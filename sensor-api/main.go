package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "/tmp/sensor_log.db")
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	h := &Handlers{db: db}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", h.Health)
	mux.HandleFunc("/sensors/anomalies", h.Anomalies)
	mux.HandleFunc("/sensors", h.Sensors)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", withCORS(mux)))
}
