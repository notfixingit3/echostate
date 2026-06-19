# Deploying EchoState

Guide for running EchoState in production-like environments. For local development see [README.md](README.md) and [CONTRIBUTING.md](CONTRIBUTING.md).

---

## Release checklist

Before tagging a release (beta or `v1.0.0`):

1. Bump `VERSION` and run `./scripts/sync-version.sh`.
2. Update `CHANGELOG.md` and `TODO.md`.
3. Run the full test suite locally:

   ```bash
   go test $(go list ./... | grep -v '/frontend/')
   cd frontend && npm run typecheck && npm run build
   cd frontend && npm run test:e2e
   ```

4. Promote `dev` → `main` when shipping stable `:main` images.
5. Tag with a `v` prefix and push the tag to trigger GHCR binaries/images.

E2E recreates the Docker Compose stack with **fresh volumes** by default so auth, scans, and delete flows run against a clean database.

---

## Quick production start

```bash
cp .env.example .env
# Edit .env — see Required secrets below

docker compose -f docker-compose.yml -f docker-compose.prod.yml pull
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d

docker compose run --rm api auth bootstrap-admin --name "Admin"
```

Open the UI, enter the enrollment code at `/login`, and register a passkey.

Pin a specific release instead of `:main`:

```bash
ECHOSTATE_API_IMAGE=ghcr.io/notfixingit3/echostate:0.0.1-beta.39
ECHOSTATE_FRONTEND_IMAGE=ghcr.io/notfixingit3/echostate-frontend:0.0.1-beta.39
```

Set both images to the **same** version tag.

---

## Required secrets

| Variable | Purpose |
| -------- | ------- |
| `ECHOSTATE_ENV` | Set to `production` (enables Gin release mode and `Secure` session cookies). |
| `ECHOSTATE_AUTH_PEPPER` | HMAC pepper for enrollment-code and session hashing. Generate a long random string. |
| `ECHOSTATE_BREAK_GLASS_SECRET` | Secret for `auth issue-admin-code` recovery CLI. |
| `FRONTEND_URL` | Public HTTPS origin of the UI (WebAuthn RP ID/origin must match). |
| `POSTGRES_PASSWORD` | Strong database password (change from compose defaults). |
| `NEO4J_PASSWORD` | Strong Neo4j password if graph sync is enabled. |

Optional but recommended:

- Reverse proxy TLS termination (nginx, Caddy, Traefik).
- Unique `ECHOSTATE_API_KEY` only for automation before passkeys exist — remove after bootstrap.

---

## TLS and reverse proxy

EchoState expects:

- Browser → **HTTPS** → reverse proxy → frontend (`:3001`) and API (`:8080`).
- `FRONTEND_URL` set to the public origin (e.g. `https://echostate.example.com`).
- API `TrustedProxies` includes common private CIDRs (see `main.go`); extend if your proxy uses other ranges.

The frontend nginx image proxies `/api` to the API service inside Compose. External proxies can either:

- Expose only the frontend port and let it proxy `/api`, or
- Route `/api` directly to the API container (set `NEXT_PUBLIC_API_URL` at frontend build time if the API is on a separate host).

---

## Authentication hardening

- **Bootstrap before exposure** — Do not expose the stack to the internet until `auth bootstrap-admin` has run and a passkey is registered.
- **Pre-bootstrap behavior** — With zero users, write endpoints accept an optional legacy API key; reads are open. This matches early betas and is unsafe on a public URL.
- **Recovery** — Store `ECHOSTATE_BREAK_GLASS_SECRET` offline; use `docker compose run --rm api auth issue-admin-code` if admins lose passkeys.
- **Session cookies** — `HttpOnly`, `SameSite=Lax`, `Secure` when `ECHOSTATE_ENV=production`.

---

## Backup and restore

### PostgreSQL volume

Back up the `postgres_data` Docker volume on a schedule (e.g. `pg_dump` from the `db` container).

### Application export (admin)

```bash
curl -b echostate_session=... \
  "https://echostate.example.com/api/export?include_blobs=true&include_config=false" \
  -o echostate-export.json
```

Restore via `POST /api/import` (admin) from **Admin → Data** or the API. See `internal/export` for conflict options.

### Neo4j

If graph views matter, back up the `neo4j_data` volume or plan to rebuild the graph from snapshots after restore (`rebuild_graph` on import).

---

## Upgrades

1. Back up Postgres (and Neo4j if used).
2. Pin new image tags in `.env` or pull `:main` / a semver tag.
3. `docker compose pull && docker compose up -d` — migrations run automatically on API boot (`db.Migrate()`).
4. Smoke-test: `/health`, sign-in, one scan, one report download.

Test the upgrade path in staging before production (`beta.N` → `beta.N+1` or `1.0.0`).

---

## Resource sizing (starting point)

| Service | Notes |
| ------- | ----- |
| **API** | 1 CPU, 512MB–1GB; scales with scan concurrency in Settings. |
| **browserless** | 1–2 CPU, 1–2GB; one Chrome session per concurrent web/screenshot gatherer. |
| **PostgreSQL** | 1GB+ RAM; grows with snapshot `raw_data` and PDF bytea. |
| **Neo4j** | 1–2GB heap for medium graphs; optional. |
| **Frontend** | Static nginx — minimal CPU/RAM. |

Enable retention (Settings → keep-N per target) to cap snapshot growth.

---

## Observability

- **Health** — `GET /health` and `GET /api/version`.
- **Logs** — API stdout (Gin request logging); `docker compose logs -f api`.
- **Updates** — Optional upstream version check (`ECHOSTATE_UPDATE_CHECK`).

Metrics and structured audit logs are not yet first-class; plan external log aggregation for production.

---

## Security

See [SECURITY.md](SECURITY.md) for vulnerability reporting.

Third-party licenses and external API cost models: [THIRD_PARTY.md](THIRD_PARTY.md).