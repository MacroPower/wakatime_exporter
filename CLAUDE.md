# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Development Commands

```bash
# Build the binary
CGO_ENABLED=0 go build -o wakatime_exporter .

# Run tests with coverage
go test ./... -race -coverprofile=coverage.out -v

# Run a single test
go test -run TestFunctionName ./collector/

# Lint (requires golangci-lint v2.x)
golangci-lint run

# Format code
go fmt ./...

# Vet code
go vet ./...

# Download dependencies
go mod download

# Build for Docker (Linux amd64)
mkdir -p .build/linux-amd64
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o .build/linux-amd64/wakatime_exporter .

# Run locally (requires WAKA_API_KEY)
./wakatime_exporter --wakatime.api-key="YOUR_API_KEY"
```

## Architecture

This is a Prometheus exporter that scrapes WakaTime API endpoints and exposes coding statistics as Prometheus metrics.

### Core Components

- **main.go**: Entry point, HTTP server setup, CLI flag parsing with kingpin
- **collector/collector.go**: Central collector registry and `WakaCollector` type that implements `prometheus.Collector` interface
- **collector/common.go**: Shared HTTP fetching logic (`FetchHTTP`) and JSON unmarshaling utilities

### Collector Pattern

Each metric collector follows this pattern:
1. Types file (`*_types.go`): JSON structs for WakaTime API responses
2. Implementation file (`*.go`): Collector struct with `Update(ch chan<- prometheus.Metric)` method
3. Test file (`*_test.go`): Tests using mock HTTP responses

Collectors register themselves via `init()` calling `registerCollector()`:
```go
func init() {
    registerCollector("collector-name", defaultEnabled, NewCollectorFunc)
}
```

### Available Collectors

| Collector | Endpoint | Description |
|-----------|----------|-------------|
| summary | /summaries | Daily breakdown by language, editor, project, OS, machine, category |
| all-time | /all_time_since_today | Lifetime coding totals |
| goal | /goals | Goal progress tracking |
| leader | /leaders | Public leaderboard ranking |
| stats | /stats/{range} | Pre-aggregated stats for configurable time ranges |
| durations | /durations | Duration data with AI coding metrics (lines added/deleted) |

### Metrics Namespace

All metrics use the `wakatime` namespace prefix. Metric naming follows Prometheus conventions:
- `wakatime_<subsystem>_<metric>` (e.g., `wakatime_language_seconds_total`)

### Configuration

Configuration via CLI flags or environment variables (prefixed with `WAKA_`). Collectors can be enabled/disabled individually via `--collector.<name>` flags or `WAKA_COLLECTOR_<NAME>` env vars.
