# BlackBox Monitor

![Status](https://img.shields.io/badge/status-v1.3-purple?style=for-the-badge)
![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go)
![License](https://img.shields.io/badge/license-MIT-green?style=for-the-badge)
[![Go Report Card](https://goreportcard.com/badge/github.com/MarceloAdan73/blackbox-monitor)](https://goreportcard.com/report/github.com/MarceloAdan73/blackbox-monitor)
[![CI](https://github.com/MarceloAdan73/blackbox-monitor/actions/workflows/ci.yml/badge.svg)](https://github.com/MarceloAdan73/blackbox-monitor/actions/workflows/ci.yml)
![Go Version](https://img.shields.io/github/go-mod/go-version/MarceloAdan73/blackbox-monitor?style=for-the-badge)

> Web health monitoring with terminal UI and web dashboard.

![Demo](screenshots/demo.gif)

## Overview

BlackBox Monitor periodically checks HTTP endpoints, displays real-time status in a Lip Gloss terminal UI and a dark-mode web dashboard, stores check history (optionally in SQLite), and sends Telegram alerts on state changes — all in a single Go binary with zero runtime dependencies.

## Architecture

- **Main loop** runs on a `time.Ticker`, triggers concurrent `CheckSite()` calls via goroutines, collects results, and updates shared state protected by `sync.RWMutex`
- **Web server** (`net/http`) serves the dashboard, API, and CSV export from the same process — no reverse proxy needed
- **Telegram notifier** runs async via goroutines so network latency never blocks the check cycle
- **Storage** uses a `Store` interface with two implementations selected at compile time via Go build tags (`sqlite` / `!sqlite`), keeping the binary small by default
- **SQLite** storage uses CGo but is optional — the default build produces a fully static binary
- **Static assets** (HTML, CSS, JS) are embedded using `//go:embed` for a single deployment artifact

## Features

- HTTP check with latency and SSL cert expiry detection
- Terminal UI with Lip Gloss (colored boxes, borders, real-time updates)
- Web dashboard with CSS Grid + ApexCharts (bar chart, donut, auto-refresh)
- Telegram alerts on state transitions (async, non-blocking)
- SQLite persistence (optional build tag, retains check history)
- CSV export, healthcheck endpoint, JSON API
- YAML configuration with `-config`, `-interval`, `-port` flags

## Quick start

```bash
cp config.example.yaml config.yaml
go build -o bin/blackbox-monitor . && ./bin/blackbox-monitor
# Open http://localhost:8080
```

## Configuration

```yaml
interval: 60
telegram:
  enabled: false
  bot_token: ""
  chat_id: ""
sites:
  - name: "My Site"
    url: "https://example.com"
    timeout: 5000
```

## Telegram alerts

When `enabled: true`, the monitor sends a notification on startup and whenever a site transitions between Online and Offline. Delivery happens in a goroutine — it never blocks the check cycle.

## Dashboard

![Dashboard](screenshots/dashboard.png)

- Summary cards, bar chart, donut chart (ApexCharts)
- Status table with HTTP code, latency, SSL info
- Auto-refresh every 10s, CSV download
- Responsive dark mode, no external CSS framework

### API

```
GET /api/status   → JSON (sites, online count, avg latency, uptime)
GET /api/export   → CSV download
GET /health       → {"status":"ok","uptime":"...","version":"1.3"}
```

## Project structure

```
├── main.go               # Entry point, flag parsing, signal handling
├── server.go             # HTTP server, /api/status, /api/export, /health
├── state.go              # Thread-safe shared state (sync.RWMutex)
├── storage_entry.go      # Store interface
├── storage_sqlite.go     # SQLite impl (build tag: sqlite)
├── storage_nosqlite.go   # In-memory impl (build tag: !sqlite)
├── internal/
│   ├── monitor/checker.go    # HTTP check + TLS cert extraction
│   ├── notifier/telegram.go  # Async Telegram alerts
│   ├── storage/sqlite.go     # SQLite storage internals
│   └── ui/styles.go          # Lip Gloss terminal styles
└── web/static/
    └── index.html            # Dashboard (CSS Grid + ApexCharts)
```

## Build variants

```bash
# Default — static binary, in-memory storage
go build -o bin/blackbox-monitor .

# With SQLite (requires gcc / CGo)
go build -tags sqlite -o bin/blackbox-monitor .

# Cross-compile (static, no CGo needed)
GOOS=linux GOARCH=amd64 go build -o bin/blackbox-monitor .
```

## Tests

```bash
go test ./... -v -count=1
go test -race ./... -count=1
go test ./... -cover
```

## License

MIT. See [LICENSE](LICENSE).

## Author

**Marcelo Adan** — [GitHub](https://github.com/MarceloAdan73)
