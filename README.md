# BlackBox Monitor

![Status](https://img.shields.io/badge/status-v1.3-purple?style=for-the-badge)
![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go)
![License](https://img.shields.io/badge/license-MIT-green?style=for-the-badge)
[![Go Report Card](https://goreportcard.com/badge/github.com/MarceloAdan73/blackbox-monitor)](https://goreportcard.com/report/github.com/MarceloAdan73/blackbox-monitor)
[![CI](https://github.com/MarceloAdan73/blackbox-monitor/actions/workflows/ci.yml/badge.svg)](https://github.com/MarceloAdan73/blackbox-monitor/actions/workflows/ci.yml)

> Web health monitoring with terminal UI and web dashboard.

![Demo](screenshots/demo.gif)

## Features

- HTTP check + latency measurement
- Terminal UI with Lip Gloss styles
- Web dashboard (CSS Grid + ApexCharts)
- SSL certificate expiry check
- Telegram alerts (async, non-blocking)
- SQLite history (optional, build tag)
- CSV export, healthcheck endpoint
- YAML configuration

## Quick start

```bash
cp config.example.yaml config.yaml   # edit with your sites
go build -o bin/blackbox-monitor . && ./bin/blackbox-monitor
# Open http://localhost:8080
```

Flags: `-config path`, `-interval N`, `-port :8080` (empty disables web).

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

Sends notifications when a site changes state (Online ↔ Offline). Runs asynchronously — doesn't block monitoring.

Set `telegram.enabled: true` with your bot token and chat ID.

## Dashboard

![Dashboard](screenshots/dashboard.png)

- 4 summary cards (Total, Online, Offline, Avg latency)
- Bar chart + donut chart (ApexCharts)
- Status table with HTTP code, SSL info
- Auto-refresh every 10s, CSV download
- Responsive dark mode

### API

```
GET /api/status   → JSON with all sites, stats, latency
GET /api/export   → CSV download
GET /health       → {"status":"ok","uptime":"...","version":"1.3"}
```

## Project structure

```
├── main.go               # Entry point
├── server.go             # HTTP server + endpoints
├── state.go              # Thread-safe shared state
├── storage_entry.go      # Store interface
├── storage_sqlite.go     # SQLite impl (tag: sqlite)
├── storage_nosqlite.go   # In-memory impl (tag: !sqlite)
├── internal/
│   ├── monitor/checker.go
│   ├── notifier/telegram.go
│   ├── storage/sqlite.go
│   └── ui/styles.go
└── web/static/
    └── index.html
```

## Tests

```bash
go test ./... -v
go test -race ./...
```

## License

MIT. See [LICENSE](LICENSE).

## Author

**Marcelo Adan** — [GitHub](https://github.com/MarceloAdan73)
