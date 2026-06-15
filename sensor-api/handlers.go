package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

type Handlers struct {
	db *sql.DB
}

type SensorRow struct {
	SensorID  string  `json:"sensor_id"`
	WeightKg  float64 `json:"weight_kg"`
	ItemType  string  `json:"item_type"`
	Topic     string  `json:"topic"`
	Timestamp string  `json:"timestamp"`
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *Handlers) Sensors(w http.ResponseWriter, r *http.Request) {
	const q = `SELECT sensor_id, weight_kg, item_type, topic, timestamp
	           FROM sensor_log
	           ORDER BY timestamp DESC
	           LIMIT 50`
	h.writeRows(w, q)
}

func (h *Handlers) Anomalies(w http.ResponseWriter, r *http.Request) {
	const q = `SELECT sensor_id, weight_kg, item_type, topic, timestamp
	           FROM sensor_log
	           WHERE topic = 'anomaly'
	           ORDER BY timestamp DESC
	           LIMIT 50`
	h.writeRows(w, q)
}

func (h *Handlers) writeRows(w http.ResponseWriter, query string) {
	rows, err := h.db.Query(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	result := make([]SensorRow, 0)
	for rows.Next() {
		var s SensorRow
		if err := rows.Scan(&s.SensorID, &s.WeightKg, &s.ItemType, &s.Topic, &s.Timestamp); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		result = append(result, s)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
