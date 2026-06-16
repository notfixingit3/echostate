<p align="center">
  <img src="assets/logo-banner.png" alt="EchoState" width="640">
</p>

# EchoState

EchoState is an unobtrusive, passive reconnaissance web API and reporting tool. Feed it an IP address, hostname, or web address and it gathers deep, passive intelligence — WHOIS records, BGP/ASN routing, hosting provider hints, and webpage copyright data — then tracks changes over time and generates downloadable PDF reports.

## Features

- **Passive Reconnaissance:** WHOIS, ASN/BGP, and DOM scraping run concurrently.
- **Historical Tracking:** Old snapshots are never deleted; differences are highlighted.
- **PDF Reporting:** Renders a clean HTML report and prints it to PDF via headless Chrome.
- **Containerized:** Docker Compose stack with Go API, PostgreSQL, and browserless Chrome.

## Quick Start

```bash
cp .env.example .env
docker compose up --build
```

The API will be available at `http://localhost:8080`.

## API

### Health Check

```bash
curl http://localhost:8080/health
```

### Start a Scan

```bash
curl -X POST http://localhost:8080/api/scan \
  -H "Content-Type: application/json" \
  -d '{"host":"example.com"}'
```

## Development

Requirements:

- Go 1.23+
- PostgreSQL 16+
- browserless/chrome or another Chrome DevTools Protocol endpoint

Run locally:

```bash
go run ./main.go
```

## Architecture

- `cmd/` — application entry point
- `internal/config/` — environment configuration
- `internal/db/` — PostgreSQL connection and migrations
- `internal/handlers/` — Gin HTTP handlers
- `internal/models/` — domain models
- `internal/scanner/` — reconnaissance orchestration

## License

[MIT](LICENSE)
