<p align="center">
  <img src="assets/logo-banner.png" alt="EchoState" width="640">
</p>

# EchoState

EchoState is a passive reconnaissance platform: a Go API plus Next.js web UI for scanning hosts, tracking changes over time, and downloading PDF reports.

Feed it a hostname, IP, or URL and it gathers WHOIS, BGP/ASN, DNS, TLS certificate, traceroute, certificate transparency, and webpage intelligence concurrently. Snapshots are never deleted — rescans highlight diffs. Submitter IPs are enriched asynchronously via [pWhois](https://pwhois.org/).

## Features

- **Passive recon** — WHOIS, Team Cymru ASN/BGP (RIPEstat hijack-risk heuristics, AS-path enrichment, PeeringDB IX data), DNS, TLS + JARM, crt.sh subdomain discovery, multi-vantage traceroute (local + HackerTarget), favicon MMH3, robots/sitemap crawl, cloud bucket hints, HTTP security headers and full response headers from the final page load, tech-stack fingerprinting, headless Chrome web scraping, and JPEG screenshot thumbnails.
- **Web UI** — Scan form, target detail with tags and rescan, intel tabs (WHOIS, ASN, DNS, TLS, Web, Favicon, Crawl, Storage, CT, Traceroute, Screenshots, Submitter), snapshot browser, side-by-side raw diffs, settings, and report downloads. Light mode default; version shown in footer.
- **Historical tracking** — Snapshots persist; identical rescans update `last_seen`. Field-level `change_details` (severity, type, summary) highlight TLS expiry, new CT subdomains, DNS shifts, and more.
- **Async scans** — `POST /api/scan` enqueues a job (`202`); poll `GET /api/scans/:id` for status and the resulting snapshot.
- **Scheduled rescans** — Background scheduler re-scans stale targets on a configurable interval.
- **PDF reports** — Async worker renders HTML → PDF via headless Chrome.
- **IP enrichment** — pWhois worker enriches submitter IPs (org, ASN, geo).
- **Notifications** — Slack, Discord, MS Teams webhooks, and Pushover mobile alerts with structured change payloads and alert-rule filtering.
- **Settings** — DNS resolvers, pWhois server, rate limit, scan concurrency/timeouts, scheduler, retention, alert rules (with per-webhook filters), and optional Shodan/Censys/HIBP/RiskIQ API keys.
- **Optional API key** — Set `ECHOSTATE_API_KEY` (or configure in Settings) to protect write endpoints (`scan`, `reports`, `webhooks`, `settings`).
- **Neo4j graph** — Relationship sync on scan plus interactive `/graph` UI with infra, CT, DNS, cert SAN, BGP, traceroute, and peering views; route diff, shared hops, and infra clusters.
- **Containerized** — Docker Compose: API, frontend, PostgreSQL, browserless Chrome, Neo4j.

> **Production note:** Write endpoints accept an optional API key (`ECHOSTATE_API_KEY` or Settings). Browse endpoints and submitter IPs remain publicly visible without additional auth. Deploy behind a reverse proxy with TLS, set an API key, and restrict network access. See [SECURITY.md](SECURITY.md).

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
# {"status":"ok","env":"development","version":"0.0.1-beta.13"}
```

### Scan a target

```bash
# Enqueue scan (202 Accepted)
curl -X POST http://localhost:8080/api/scan \
  -H "Content-Type: application/json" \
  -d '{"host":"example.com"}'
# {"id":"<job-uuid>","status":"pending",...}

# Poll until completed
curl http://localhost:8080/api/scans/<job-uuid>
```

When `status` is `completed`, the response includes the snapshot with `raw_data` (`whois`, `asn`, `dns`, `tls`, `web`, `ct`, `traceroute`, `screenshot`, `favicon`, `crawl`, `storage`, `errors`) and `change_details` vs the previous snapshot.

If `ECHOSTATE_API_KEY` is set, pass `Authorization: Bearer <key>` or `X-API-Key: <key>` on write requests.

Rescanning a target uses the same endpoint — there is no separate rescan API.

### Browse & manage

| Endpoint | Description |
| -------- | ----------- |
| `GET /api/targets` | Paginated targets (`page`, `limit`, `q`) |
| `GET /api/targets/:id` | Target detail + latest snapshot |
| `PUT /api/targets/:id/tags` | Update target tags |
| `GET /api/targets/:id/snapshots` | Snapshots for a target |
| `GET /api/targets/:id/screenshots` | Screenshot timeline (base64 JPEG thumbnails) |
| `GET /api/snapshots` | All snapshots (`target_id` filter) |
| `GET /api/snapshots/:id` | Full snapshot with pWhois data |
| `GET /api/snapshots/:id/diff` | Raw JSON diff vs previous snapshot |
| `GET /api/snapshots/:id/report` | Queue or fetch report for a snapshot |
| `GET /api/reports` | Report list (`status`, `snapshot_id`) |
| `POST /api/reports` | Queue PDF (`{"snapshot_id":"..."}`) |
| `GET /api/reports/:id/download` | Download completed PDF |
| `GET/POST/PUT/DELETE /api/webhooks` | Webhook management |
| `GET /api/scans/:id` | Async scan job status (+ snapshot when complete) |
| `GET/PUT /api/settings` | System settings (DNS, pWhois, rate limit, scheduler, retention, alert rules, API keys) |
| `GET /api/graph` | Infrastructure graph (`?target_id=` optional; `?view=` = `infra`, `ct`, `dns`, `cert`, `bgp`, `traceroute`, `peering`) |

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
│   ├── db/                 # PostgreSQL migrations + Neo4j graph sync
│   ├── handlers/           # Gin routes (scan, browse, graph, reports, settings)
│   ├── middleware/         # CORS, rate limiting
│   ├── models/             # Domain types
│   ├── pwhois/             # Async IP enrichment worker
│   ├── reports/            # Async PDF worker
│   ├── scanner/            # WHOIS, ASN, DNS, TLS, CT, traceroute, favicon, crawl, storage, web, screenshot gatherers
│   ├── version/            # Build-time version injection
│   ├── webhooks/           # Snapshot notification dispatcher (Slack, Discord, Teams, Pushover)
│   └── pdf/                # Report HTML renderer
└── frontend/               # Next.js static export + nginx
```

## Releases

Tagged releases (`v*`) trigger:

- **GitHub Release** — Linux and macOS binaries (amd64 + arm64)
- **GHCR images** — `ghcr.io/notfixingit3/echostate` (API) and `ghcr.io/notfixingit3/echostate-frontend`

Pre-release tags containing `beta`, `alpha`, or `rc` are marked as GitHub pre-releases.

Bump the root `VERSION` file and run `./scripts/sync-version.sh` before tagging. Set `ECHOSTATE_VERSION` in `.env` for local Docker builds to match.

## License

[MIT](LICENSE)