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
| `/` | Scan form + intel summary; scan result card polls reports and supports PDF download |
| `/targets` | Browse targets |
| `/target?id=` | Target detail, intel tabs, rescan button |
| `/target/snapshots?id=` | Snapshot history for a target |
| `/snapshots` | Browse all snapshots |
| `/snapshot?id=` | Snapshot detail; hydrates latest report on load; **Create report** polls until PDF is ready, then download |
| `/reports` | Browse reports (`?snapshot_id=` filter supported) |
| `/report?id=` | Report status (auto-refresh) + credentialed PDF download |
| `/profile` | Display name, theme, timezone, passkeys, device enrollment codes |
| `/admin` | Server settings: integrations, system, authentication, user management |
| `/settings/*` | Legacy redirects to `/profile` or `/admin` |
| `/graph` | Interactive Neo4j graph (views: infra, CT, DNS, cert, BGP, traceroute, peering) |

Target and snapshot detail pages use query-param routes because the app is a static export — UUIDs are not baked into the build.

## Tests

```bash
npm run typecheck
npm run build
npx playwright test    # builds local dev images; may recreate volumes in setup — avoid on stacks with data you need
```