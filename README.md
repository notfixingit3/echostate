<p align="center">
  <img src="assets/logo-banner.png" alt="EchoState" width="640">
</p>

# EchoState

EchoState is a passive reconnaissance platform: a Go API plus Next.js web UI for scanning hosts, tracking changes over time, and downloading PDF reports.

Feed it a hostname, IP, or URL and it gathers WHOIS, BGP/ASN, DNS, TLS certificate, and webpage intelligence concurrently. Snapshots are never deleted — rescans highlight diffs. Submitter IPs are enriched asynchronously via [pWhois](https://pwhois.org/).

## Features

- **Passive recon** — WHOIS, Team Cymru ASN/BGP (with RIPEstat hijack-risk heuristics), DNS, TLS + JARM, favicon MMH3, robots/sitemap crawl, cloud bucket hints, HTTP security headers, tech-stack fingerprinting, and headless Chrome web scraping.
- **Web UI** — Scan form, target detail with tags, snapshot browser, side-by-side raw diffs, settings, and report downloads.
- **Historical tracking** — Snapshots persist; identical rescans update `last_seen` and record field-level changes.
- **PDF reports** — Async worker renders HTML → PDF via headless Chrome.
- **IP enrichment** — pWhois worker enriches submitter IPs (org, ASN, geo).
- **Notifications** — Slack, Discord, MS Teams webhooks, and Pushover mobile alerts on snapshot changes.
- **Settings** — Configure DNS resolvers, pWhois server, and global scan rate limit from the UI.
- **Neo4j graph** — Optional relationship sync for infrastructure linking (Compose included).
- **Containerized** — Docker Compose: API, frontend, PostgreSQL, browserless Chrome, Neo4j.

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
| Neo4j     | bolt://localhost:7687       |

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

Returns a snapshot with `raw_data` containing `whois`, `asn`, `dns`, `tls`, `web`, and any `errors`.

### Browse & manage

| Endpoint | Description |
| -------- | ----------- |
| `GET /api/targets` | Paginated targets (`page`, `limit`, `q`) |
| `GET /api/targets/:id` | Target detail + latest snapshot |
| `PUT /api/targets/:id/tags` | Update target tags |
| `GET /api/targets/:id/snapshots` | Snapshots for a target |
| `GET /api/snapshots` | All snapshots (`target_id` filter) |
| `GET /api/snapshots/:id` | Full snapshot with pWhois data |
| `GET /api/snapshots/:id/diff` | Raw JSON diff vs previous snapshot |
| `GET /api/reports` | Report list (`status`, `snapshot_id`) |
| `POST /api/reports` | Queue PDF (`{"snapshot_id":"..."}`) |
| `GET /api/reports/:id/download` | Download completed PDF |
| `GET/POST/PUT/DELETE /api/webhooks` | Webhook management |
| `GET/PUT /api/settings` | System settings (DNS, pWhois, rate limit) |

## Development

### Requirements

- Go 1.26+
- Node.js 20+ (frontend)
- PostgreSQL 16+
- browserless/chrome or another CDP WebSocket endpoint
- Neo4j 5+ (optional; included in Compose)

### Backend

```bash
cp .env.example .env
docker compose up db browser neo4j -d   # or point env vars yourself
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
│   ├── config/             # Environment + runtime settings
│   ├── db/                 # PostgreSQL migrations + Neo4j client
│   ├── handlers/           # Gin routes (scan, browse, reports, settings)
│   ├── middleware/         # CORS, rate limiting
│   ├── models/             # Domain types
│   ├── pwhois/             # Async IP enrichment worker
│   ├── reports/            # Async PDF worker
│   ├── scanner/            # WHOIS, ASN, DNS, TLS, web gatherers
│   ├── webhooks/           # Snapshot notification dispatcher
│   └── pdf/                # Report HTML renderer
└── frontend/               # Next.js static export + nginx
```

## Releases

Tagged releases (`v*`) trigger:

- **GitHub Release** — Linux and macOS binaries (amd64 + arm64)
- **GHCR images** — `ghcr.io/notfixingit3/echostate` (API) and `ghcr.io/notfixingit3/echostate-frontend`

Pre-release tags containing `beta`, `alpha`, or `rc` are marked as GitHub pre-releases.

## License

[MIT](LICENSE)