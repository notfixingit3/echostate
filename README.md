<p align="center">
  <img src="assets/logo-banner.png" alt="EchoState" width="640">
</p>

# EchoState

EchoState is a passive reconnaissance platform: a Go API that gathers WHOIS, BGP/ASN, and webpage intelligence for any host, IP, or URL, plus a Next.js web UI to scan, browse history, and download PDF reports.

Snapshots are never deleted — rescans highlight what changed. Submitter IPs are enriched asynchronously via [pWhois](https://pwhois.org/).

## Features

- **Passive recon** — WHOIS, Team Cymru ASN/BGP, and headless Chrome web scraping run concurrently.
- **Web UI** — Scan form, target detail pages, snapshot browser, and report downloads.
- **Historical tracking** — Snapshots persist; identical rescans update `last_seen` and diffs are recorded.
- **PDF reports** — Async worker renders HTML → PDF via headless Chrome.
- **IP enrichment** — pWhois worker enriches submitter IPs (org, ASN, geo).
- **Rate limiting** — 30 scans per IP per minute on `POST /api/scan`.
- **Containerized** — Docker Compose stack: API, frontend, PostgreSQL, browserless Chrome.

## Quick Start

```bash
cp .env.example .env
docker compose up --build
```

| Service   | URL                         |
| --------- | --------------------------- |
| Web UI    | http://localhost:3001       |
| API       | http://localhost:8080       |
| Browser   | ws://localhost:3000/        |

The UI proxies `/api` to the Go backend inside Docker — no CORS setup required for local use.

## API

### Health

```bash
curl http://localhost:8080/health
```

### Scan a target

```bash
curl -X POST http://localhost:8080/api/scan \
  -H "Content-Type: application/json" \
  -d '{"host":"example.com"}'
```

Returns a snapshot with `raw_data` containing `whois`, `asn`, `web`, and any `errors`.

### Browse

| Endpoint | Description |
| -------- | ----------- |
| `GET /api/targets` | Paginated targets (`page`, `limit`, `q`) |
| `GET /api/targets/:id` | Target detail + latest snapshot |
| `GET /api/targets/:id/snapshots` | Snapshots for a target |
| `GET /api/snapshots` | All snapshots (`target_id` filter) |
| `GET /api/snapshots/:id` | Full snapshot with pWhois data |
| `GET /api/reports` | Report list (`status`, `snapshot_id`) |
| `POST /api/reports` | Queue PDF (`{"snapshot_id":"..."}`) |
| `GET /api/reports/:id/download` | Download completed PDF |

## Development

### Requirements

- Go 1.26+
- Node.js 20+ (frontend)
- PostgreSQL 16+
- browserless/chrome or another CDP WebSocket endpoint

### Backend

```bash
cp .env.example .env
# Start db + browser via compose, or point DATABASE_URL / BROWSER_WS_URL yourself
go run ./main.go
```

### Frontend

```bash
cd frontend
npm install
cp .env.local.example .env.local   # NEXT_PUBLIC_API_URL=http://localhost:8080
npm run dev
```

### Tests

```bash
# Backend (excludes frontend/node_modules Go shim)
go test $(go list ./... | grep -v '/frontend/')

# Frontend
cd frontend && npm run typecheck && npm run build

# E2E (requires full docker compose stack)
cd frontend && npx playwright test
```

## Architecture

```
echostate/
├── main.go                 # Entry point, workers, HTTP server
├── internal/
│   ├── config/             # Environment configuration
│   ├── db/                 # PostgreSQL + migrations
│   ├── handlers/         # Gin routes (scan, browse, reports)
│   ├── middleware/         # CORS, rate limiting
│   ├── models/             # Domain types
│   ├── pwhois/             # Async IP enrichment worker
│   ├── reports/            # Async PDF worker
│   ├── scanner/            # WHOIS, ASN, web gatherers
│   └── pdf/                # Report HTML renderer
└── frontend/               # Next.js static export + nginx
```

## Releases

Tagged releases (`v*`) trigger:

- **GitHub Release** — Linux and macOS binaries (amd64 + arm64)
- **GHCR images** — `ghcr.io/notfixingit3/echostate` (API) and `ghcr.io/notfixingit3/echostate-frontend`

Pre-release tags containing `beta` or `dev` are marked as GitHub pre-releases.

## License

[MIT](LICENSE)