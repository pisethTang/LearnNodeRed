package main

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const dbPath = "./load_test.db"

type scenario struct {
	name     string
	bins     int
	interval time.Duration
	duration time.Duration
}

func main() {
	scenarios := []scenario{
		{name: "baseline (5 bins / 5s)", bins: 5, interval: 5 * time.Second, duration: 30 * time.Second},
		{name: "10 msg/s", bins: 10, interval: 1 * time.Second, duration: 30 * time.Second},
		{name: "100 msg/s", bins: 100, interval: 1 * time.Second, duration: 30 * time.Second},
		{name: "500 msg/s", bins: 500, interval: 1 * time.Second, duration: 30 * time.Second},
		{name: "500 bins / 5s", bins: 500, interval: 5 * time.Second, duration: 30 * time.Second},
	}

	for _, s := range scenarios {
		runScenario(s)
	}
}

func runScenario(s scenario) {
	cleanup()
	defer cleanup()

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS sensor_log (
		sensor_id TEXT,
		weight_kg REAL,
		item_type TEXT,
		topic TEXT,
		timestamp TEXT
	)`); err != nil {
		log.Fatalf("create table: %v", err)
	}

	insertStmt, err := db.Prepare(`INSERT INTO sensor_log VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		log.Fatalf("prepare insert: %v", err)
	}
	defer insertStmt.Close()

	queryStmt, err := db.Prepare(`SELECT sensor_id, weight_kg, item_type, topic, timestamp
		FROM sensor_log ORDER BY timestamp DESC LIMIT 50`)
	if err != nil {
		log.Fatalf("prepare query: %v", err)
	}
	defer queryStmt.Close()

	insertRate := float64(s.bins) / s.interval.Seconds()
	fmt.Printf("\n=== %s (%.1f inserts/sec, %d bins) ===\n", s.name, insertRate, s.bins)

	var inserted int64
	var errors int64
	var latencies []time.Duration
	var mu sync.Mutex
	stop := make(chan struct{})
	var wg sync.WaitGroup

	// Start workers, one per bin
	for b := 0; b < s.bins; b++ {
		wg.Add(1)
		go func(binID int) {
			defer wg.Done()
			ticker := time.NewTicker(s.interval)
			defer ticker.Stop()

			for {
				select {
				case <-stop:
					return
				case <-ticker.C:
					start := time.Now()
					_, err := insertStmt.Exec(
						fmt.Sprintf("BIN_%03d", binID+1),
						rand.Float64()*60,
						[]string{"glass", "plastic", "aluminium"}[rand.Intn(3)],
						"normal",
						time.Now().UTC().Format(time.RFC3339Nano),
					)
					elapsed := time.Since(start)

					mu.Lock()
					latencies = append(latencies, elapsed)
					mu.Unlock()

					if err != nil {
						atomic.AddInt64(&errors, 1)
					} else {
						atomic.AddInt64(&inserted, 1)
					}
				}
			}
		}(b)
	}

	// Query load generator: run the same query as the Go API every 5 seconds
	var queryLatencies []time.Duration
	var queryErrors int64
	queryStop := make(chan struct{})
	var queryWg sync.WaitGroup
	queryWg.Add(1)
	go func() {
		defer queryWg.Done()
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-queryStop:
				return
			case <-ticker.C:
				start := time.Now()
				rows, err := queryStmt.Query()
				if err != nil {
					atomic.AddInt64(&queryErrors, 1)
					continue
				}
				count := 0
				for rows.Next() {
					count++
				}
				rows.Close()
				elapsed := time.Since(start)
				mu.Lock()
				queryLatencies = append(queryLatencies, elapsed)
				mu.Unlock()
			}
		}
	}()

	time.Sleep(s.duration)
	close(stop)
	wg.Wait()
	close(queryStop)
	queryWg.Wait()

	// Stats
	mu.Lock()
	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
	sort.Slice(queryLatencies, func(i, j int) bool { return queryLatencies[i] < queryLatencies[j] })
	insertP50, insertP95, insertP99 := percentile(latencies, 0.50), percentile(latencies, 0.95), percentile(latencies, 0.99)
	queryP50, queryP95, queryP99 := percentile(queryLatencies, 0.50), percentile(queryLatencies, 0.95), percentile(queryLatencies, 0.99)
	mu.Unlock()

	info, _ := os.Stat(dbPath)
	fileSizeMB := float64(info.Size()) / (1024 * 1024)

	fmt.Printf("inserted rows:      %d\n", atomic.LoadInt64(&inserted))
	fmt.Printf("insert errors:      %d\n", atomic.LoadInt64(&errors))
	fmt.Printf("insert latency:     p50=%s p95=%s p99=%s\n", insertP50, insertP95, insertP99)
	fmt.Printf("query latency:      p50=%s p95=%s p99=%s\n", queryP50, queryP95, queryP99)
	fmt.Printf("query errors:       %d\n", atomic.LoadInt64(&queryErrors))
	fmt.Printf("database file size: %.2f MB\n", fileSizeMB)
	fmt.Printf("actual insert rate: %.1f rows/sec\n", float64(atomic.LoadInt64(&inserted))/s.duration.Seconds())
}

func percentile(sorted []time.Duration, p float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(math.Ceil(p*float64(len(sorted)-1)))
	return sorted[idx]
}

func cleanup() {
	os.Remove(dbPath)
}
