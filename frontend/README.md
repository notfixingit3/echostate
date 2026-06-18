# EchoState Web UI

Next.js App Router frontend for EchoState. Static export served by nginx in Docker; dev server for local work.

## Stack

- Next.js 16 (App Router, static export)
- shadcn/ui (base-nova)
- Tailwind CSS 4
- Playwright for E2E tests

## Development

```bash
npm install
cp .env.local.example .env.local
npm run dev
```

Set `NEXT_PUBLIC_API_URL` to your API origin (default `http://localhost:8080` when running the API locally).

## Build

```bash
npm run build
```

Output lands in `dist/` for nginx serving.

## Docker

Built as `echostate-frontend` via `docker compose up --build`. Inside Compose, `NEXT_PUBLIC_API_URL` is empty so the UI proxies `/api` through nginx to the API container.

`NEXT_PUBLIC_APP_VERSION` is set at build time from `ECHOSTATE_VERSION` (see root `.env.example`) and shown in the footer version badge.

## Pages

| Route | Purpose |
| ----- | ------- |
| `/` | Scan form + intel summary |
| `/targets` | Browse targets |
| `/target?id=` | Target detail, intel tabs, rescan button |
| `/target/snapshots?id=` | Snapshot history for a target |
| `/snapshots` | Browse all snapshots |
| `/snapshot?id=` | Snapshot detail |
| `/reports` | Browse reports |
| `/report?id=` | Report status + PDF download |
| `/settings` | DNS resolvers, pWhois, rate limit, webhooks (Slack, Discord, Teams, Pushover) |
| `/graph` | Interactive Neo4j graph (views: infra, CT, DNS, cert, BGP, traceroute, peering) |

Target and snapshot detail pages use query-param routes because the app is a static export — UUIDs are not baked into the build.

## Tests

```bash
npm run typecheck
npm run build
npx playwright test    # requires docker compose stack running
```