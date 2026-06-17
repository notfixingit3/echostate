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

## Pages

| Route | Purpose |
| ----- | ------- |
| `/` | Scan form + intel summary |
| `/targets` | Browse targets |
| `/targets/[id]` | Target detail + latest intel |
| `/targets/[id]/snapshots` | Snapshot history for a target |
| `/snapshots` | Browse all snapshots |
| `/snapshot?id=` | Snapshot detail |
| `/reports` | Browse reports |
| `/report?id=` | Report status + PDF download |

## Tests

```bash
npm run typecheck
npm run build
npx playwright test    # requires docker compose stack running
```