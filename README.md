# Node-RED Sensor Simulation Pipeline

A hardware sensor simulation pipeline built with Node-RED, modelling a recycling depot environment — directly relevant to real-world depot automation systems. Sensor readings are generated every 5 seconds, routed through anomaly detection logic, and persisted to a SQLite database.

---

## What This Simulates

Recycling depots use weight sensors attached to collection bins to track fill levels and detect anomalies (e.g. overfilled bins, foreign objects). This pipeline simulates that environment:

- 5 bins (`BIN_01` to `BIN_05`) generate randomised weight readings every 5 seconds
- Each reading includes `sensor_id`, `weight_kg`, `item_type` (glass/plastic/aluminium), and a `timestamp`
- Readings above **40kg** are flagged as anomalies and routed separately
- All readings (normal and anomaly) are persisted to a SQLite database for audit and analysis

---

## Architecture

```
inject (every 5s)
    └── function: simulate sensor reading
            └── switch: anomaly_detector (weight_kg > 40?)
                    ├── [anomaly] change: set topic = "anomaly"
                    │       └── function: prepare_params
                    │               └── sqlite: INSERT INTO sensor_log
                    │
                    └── [normal] change: set topic = "normal"
                            └── function: prepare_params
                                    └── sqlite: INSERT INTO sensor_log
```

### Node breakdown

| Node | Type | Purpose |
|---|---|---|
| `inject_1` | Inject | Fires every 5 seconds to trigger the pipeline |
| `inject_setup` | Inject | Fires once on deploy to create the DB table |
| `function_2` | Function | Generates randomised sensor payload |
| `anomaly_detector` | Switch | Routes messages by weight threshold (>40kg) |
| `Set msg.topic to anomaly/normal` | Change | Tags each message with its classification |
| `prepare_params` | Function | Maps `msg.payload` fields into `msg.params` for SQLite |
| `sqlite_db_node` | SQLite | Creates the `sensor_log` table on startup |
| `sqlite_db_node (INSERT)` | SQLite | Persists every sensor reading to the database |

---

## Database Schema

```sql
CREATE TABLE IF NOT EXISTS sensor_log (
  sensor_id  TEXT,
  weight_kg  REAL,
  item_type  TEXT,
  topic      TEXT,
  timestamp  TEXT
);
```

Sample output:

```
BIN_05|15.98|plastic|normal|2026-06-15T04:17:24.884Z
BIN_02|3.93|glass|normal|2026-06-15T04:17:29.889Z
BIN_04|50.06|aluminium|anomaly|2026-06-15T04:17:34.889Z
BIN_04|9.72|aluminium|normal|2026-06-15T04:17:39.893Z
```

---

## How to Run

### Prerequisites

- [Node-RED](https://nodered.org/docs/getting-started/) installed
- Node.js v18+
- On WSL/Linux, install build tools first:

```bash
sudo apt-get install -y build-essential python3 sqlite3
```

### Install the SQLite node

```bash
cd ~/.node-red
npm install node-red-node-sqlite
```

> **Note:** Install via npm directly rather than the Palette Manager — the pre-built binary often fails on WSL due to a native module compilation mismatch. Installing via npm allows it to compile from source for your specific environment.

### Import the flow

1. Start Node-RED: `node-red`
2. Open `http://127.0.0.1:1880` in your browser
3. Hamburger menu → Import → select `flow.json`
4. Click Deploy

### Verify data is flowing

```bash
sqlite3 /tmp/sensor_log.db "SELECT * FROM sensor_log LIMIT 10;"
```

---

## Go REST API

A lightweight Go API server (`sensor-api/`) reads from the same SQLite database that Node-RED writes to, exposing the sensor data over HTTP.

### Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Returns `{"status":"ok"}` |
| `GET` | `/sensors` | Last 50 rows, newest first |
| `GET` | `/sensors/anomalies` | Last 50 rows where `topic = "anomaly"`, newest first |

### How to run

```bash
cd sensor-api
go run .
# listening on :8080
```

### Example requests

```bash
curl -s http://localhost:8080/health | jq .
```
```json
{"status": "ok"}
```

```bash
curl -s http://localhost:8080/sensors | jq .
```
```json
[
  {
    "sensor_id": "BIN_03",
    "weight_kg": 39.41,
    "item_type": "aluminium",
    "topic": "normal",
    "timestamp": "2026-06-15T04:36:37.121Z"
  }
]
```

```bash
curl -s http://localhost:8080/sensors/anomalies | jq .
```
```json
[
  {
    "sensor_id": "BIN_04",
    "weight_kg": 50.06,
    "item_type": "aluminium",
    "topic": "anomaly",
    "timestamp": "2026-06-15T04:17:34.889Z"
  }
]
```

---

## Key Debugging Lessons

**Problem 1 — Native module compilation failure**

The sqlite3 package includes a C++ binary that must match your OS and Node.js version. The Palette Manager installs a pre-built binary that fails silently on WSL. Fix: install via npm so it compiles from source.

**Problem 2 — Empty rows in the database**

Node-RED's sqlite node in "Prepared Statement" mode expects SQL parameters in `msg.params`, not `msg.payload`. Fix: add a function node that explicitly maps payload fields into `msg.params` with `$`-prefixed keys before the INSERT node.

---

## Skills Demonstrated

- Node-RED flow design (inject, function, switch, change, sqlite nodes)
- Hardware/software integration patterns
- SQL schema design and parameterised queries
- Debugging native Node.js modules on Linux/WSL
- Anomaly detection routing logic