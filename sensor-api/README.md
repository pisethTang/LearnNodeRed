# sensor-api

A minimal Go REST API that serves data from the `/tmp/sensor_log.db` SQLite database.

## Prerequisites

- Go 1.21+
- GCC (required by `go-sqlite3` for CGO compilation)
  - Ubuntu/Debian: `sudo apt install gcc`
  - Arch: `sudo pacman -S gcc`

## Run

```bash
cd sensor-api
go run .
```

The server starts on **http://localhost:8080**.

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Returns `{"status":"ok"}` |
| `GET` | `/sensors` | Last 50 rows, newest first |
| `GET` | `/sensors/anomalies` | Last 50 rows where `topic = 'anomaly'`, newest first |

### Example response (`/sensors`)

```json
[
  {
    "sensor_id": "BIN_01",
    "weight_kg": 47.86,
    "item_type": "aluminium",
    "topic": "anomaly",
    "timestamp": "2026-06-15T04:17:34.889Z"
  }
]
```

## Build a binary

```bash
go build -o sensor-api .
./sensor-api
```
