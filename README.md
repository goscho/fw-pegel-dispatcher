# fw-pegel-dispatcher

Reads values from a stream level and rainfall sensor via a W&T Web-IO device and pushes them to two cloud targets: [ThingSpeak](https://thingspeak.com) and a custom website API.

Written in Go, runs as a headless scheduled process every 5 minutes. Typical resource footprint: ~10 MB RAM, ~15 MB Docker image.

## How it works

```
W&T Web-IO device  →  fw-pegel-dispatcher  →  ThingSpeak API
                                           →  Website API (fw-murrhardt.de)
```

Every 5 minutes the scheduler:

1. Issues an HTTP GET to the Web-IO device and parses the plain-text response (e.g. `0,209 m;0,000 l/m²`).
2. Rounds stream level (port 1) to 2 decimal places and rainfall (port 2) to 0 decimal places.
3. POSTs both values to ThingSpeak with an ISO 8601 UTC timestamp.
4. POSTs level level to the website API at `/api/pegel` with JSON payload and `X-API-Key` authentication.

Both push targets are called independently — a ThingSpeak failure does not prevent the website update from being attempted.

## Prerequisites

- Go 1.26+
- Docker (for container builds)

Install Go via [go.dev/dl](https://go.dev/dl/) or your package manager.

## Project structure

```
cmd/
  level-updater/
    main.go              # entry point: config, wiring, cron start, signal handling
internal/
  config/
    config.go            # loads and validates configuration from environment variables
    config_test.go
  scheduler/
    scheduler.go         # orchestration logic + rounding; defines consumer-side interfaces
    scheduler_test.go
  webio/
    client.go            # HTTP GET + response parsing for W&T Web-IO
    client_test.go
  thingspeak/
    client.go            # HTTP POST form-encoded to ThingSpeak API
    client_test.go
  website/
    client.go            # HTTP POST JSON to /api/pegel with API key header
    client_test.go
Dockerfile               # multi-stage: golang:1.26-alpine → alpine:3.20
go.mod
go.sum
```

## Configuration

All configuration is passed via environment variables. There are no configuration files.

| Variable              | Required | Default              | Description                                        |
|-----------------------|----------|----------------------|----------------------------------------------------|
| `WEBIO_URL`           | yes      | —                    | URL of the W&T Web-IO `/Single` endpoint           |
| `THINGSPEAK_API_URL`  | yes      | —                    | ThingSpeak write API URL (e.g. `https://api.thingspeak.com/update`) |
| `THINGSPEAK_API_KEY`  | yes      | —                    | ThingSpeak channel write API key                   |
| `PEGEL_API_BASE_URL`  | yes      | —                    | Base URL of the pegel service (POST target is `/api/pegel`) |
| `PEGEL_API_KEY`       | yes      | —                    | API key sent as `X-API-Key` for website POST requests |
| `SCHEDULE_CRON`       | no       | `0 */5 * * * *`      | 6-field cron expression (seconds, minutes, hours, …) |

The application exits immediately with a descriptive error if any required variable is missing.

## Running locally

```bash
# 1. Install dependencies
go mod download

# 2. Run tests
go test -race ./...

# 3. Build binary
go build -o fw-pegel-dispatcher ./cmd/fw-pegel-dispatcher

# 4. Run with required environment variables
export WEBIO_URL="http://192.168.1.100/Single"
export THINGSPEAK_API_URL="https://api.thingspeak.com/update"
export THINGSPEAK_API_KEY="your_write_key"
export PEGEL_API_BASE_URL="https://pegel.fw-murrhardt.de"
export PEGEL_API_KEY="your_api_key"

./fw-pegel-dispatcher
```

Stop the process with `Ctrl+C` or by sending `SIGTERM`. The application completes any in-flight update cycle before exiting.

## Running with Docker

Build the image:

```bash
docker build -t fw-pegel-dispatcher .
```

Run the container:

```bash
docker run --rm \
  -e WEBIO_URL="http://192.168.1.100/Single" \
  -e THINGSPEAK_API_URL="https://api.thingspeak.com/update" \
  -e THINGSPEAK_API_KEY="your_write_key" \
  -e PEGEL_API_BASE_URL="https://pegel.fw-murrhardt.de" \
  -e PEGEL_API_KEY="your_api_key" \
  fw-pegel-dispatcher
```

The container runs as non-root user (uid 1001). Logs are sent to stdout.

## Testing

Run the full test suite including the data-race detector:

```bash
go test -race ./...
```

Run tests for a single package:

```bash
go test -race -v ./internal/webio/
```

Tests use `net/http/httptest` to spin up local HTTP servers — no external services or mock frameworks are required.

## Cron schedule format

`SCHEDULE_CRON` uses a 6-field cron expression where the first field is **seconds**:

```
┌──────────── second (0-59)
│  ┌─────────── minute (0-59)
│  │  ┌────────── hour (0-23)
│  │  │  ┌───────── day of month (1-31)
│  │  │  │  ┌──────── month (1-12)
│  │  │  │  │  ┌───── day of week (0-6, Sunday=0)
│  │  │  │  │  │
0  */5 *  *  *  *    ← default: at second 0 of every 5th minute
```

Examples:

| Expression        | Meaning                       |
|-------------------|-------------------------------|
| `0 */5 * * * *`   | Every 5 minutes (default)     |
| `0 */10 * * * *`  | Every 10 minutes              |
| `0 0 * * * *`     | Once per hour                 |

## CI

GitHub Actions runs on every push and pull request to `master`:

1. `go test -race ./...` — all tests with race detector
2. `go build ./cmd/fw-pegel-dispatcher` — binary compilation
3. `docker build .` — Docker image build

See [`.github/workflows/go.yml`](.github/workflows/go.yml).

## Architecture notes

- **`internal/`** — enforced by the Go compiler; packages here cannot be imported by external modules.
- **Interfaces at the consumer** — `internal/scheduler` defines `Requester`, `ThingSpeakService`, and `WebsiteUpdater`. The concrete client packages (`webio`, `thingspeak`, `website`) satisfy these interfaces implicitly. This makes the scheduler trivially unit-testable with plain struct mocks.
- **Single external dependency** — [`github.com/robfig/cron/v3`](https://github.com/robfig/cron) for the cron scheduler. Everything else (HTTP, JSON handling, logging) uses the Go standard library.
- **HTTP timeout** — all outbound requests use a shared `http.Client` with a 10-second connect and read timeout.
