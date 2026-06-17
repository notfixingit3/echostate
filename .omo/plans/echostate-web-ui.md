# EchoState: Web Frontend + IP Tracking + pwhois Enrichment

## TL;DR

> **Quick Summary**: Add a React + shadcn/ui web frontend to EchoState, capture the submitter's IP on every scan snapshot, asynchronously enrich that IP via pwhois.org, expose browsable read-only API endpoints, and protect scan submissions with per-IP rate limiting.
>
> **Deliverables**:
> - Next.js 14+ frontend with shadcn/ui in `frontend/`
> - `client_ip` column on `snapshots` + capture in scan handler
> - In-memory per-IP rate limiting (30 scans/min)
> - CORS + trusted proxy configuration
> - Browse API endpoints: `GET /api/targets`, `/api/snapshots`, `/api/reports`
> - Async pwhois worker in `internal/pwhois/` with DB persistence
> - Docker Compose frontend service
> - Playwright end-to-end tests
>
> **Estimated Effort**: Medium
> **Parallel Execution**: YES — 4 waves + final verification
> **Critical Path**: 1 → 4 → 7 → 10 → 18 → 19 → 21 → F1-F4

---

## Context

### Original Request
Make EchoState more usable with a professional web frontend, track who submitted scans by IP, and enrich those IPs with pwhois.org data.

### Interview Summary
**Key Decisions**:
- Frontend: Next.js 14+ App Router + shadcn/ui, separate Docker Compose service.
- Pages: landing scan form, browse targets, browse snapshots, browse reports, snapshot detail, report detail/download.
- IP storage: `client_ip` on `snapshots` table, nullable for existing rows.
- pwhois: async worker using `github.com/georgestarcher/pwhois`, batch lookups, DB cache, not in PDFs.
- Rate limiting: 30 `POST /api/scan` requests per IP per minute, in-memory, no Redis.
- No authentication.

### Research Findings
- EchoState is a pure Go/Gin JSON API with no static file serving or CORS.
- Existing async worker pattern in `internal/reports/worker.go` can be mirrored for pwhois.
- pwhois.org uses TCP whois on port 43 to `whois.pwhois.org`; `georgestarcher/pwhois` library available.

### Metis Review
**Identified Gaps** (addressed):
- API route contracts will be explicit in the plan.
- Client IP trust handled via Gin `TrustedProxies` for Docker/internal networks.
- Existing snapshots remain nullable; old rows display as "unknown".
- pwhois failures are non-blocking, retried 3× with backoff, skip private/loopback IPs.
- Browse endpoints are read-only with pagination.

### Momus Review
**Mode**: High Accuracy Review (user-selected)
**Verdict**: OKAY
**Notes**: All referenced files exist on disk and contain the claimed content. Every task has sufficient context for execution. QA scenarios are executable with concrete tools, steps, and expected results. No blocking issues found.

---

## Work Objectives

### Core Objective
Add a React + shadcn/ui web UI, per-scan client-IP capture, and async pwhois enrichment to the existing EchoState JSON API — all publicly browsable with no user authentication.

### Concrete Deliverables
- `frontend/` Next.js 14+ App Router project with shadcn/ui.
- `internal/middleware/ratelimit.go` per-IP rate limiter.
- `internal/middleware/cors.go` CORS configuration.
- `internal/handlers/browse.go` browse endpoints.
- `internal/pwhois/worker.go` async pwhois worker.
- DB migration adding `client_ip` and pwhois columns.
- Updated `docker-compose.yml` with frontend service.
- Playwright tests in `frontend/e2e/`.

### Definition of Done
- [ ] Frontend container builds and serves all pages.
- [ ] `POST /api/scan` stores `client_ip` and returns 429 after 30 requests/min.
- [ ] pwhois worker enriches snapshot IPs asynchronously.
- [ ] Browse endpoints return paginated targets, snapshots (with IP + pwhois), reports.
- [ ] Playwright tests pass against the full Docker Compose stack.

### Must Have
- Next.js + shadcn/ui frontend with all 6 pages.
- `client_ip` captured and stored per snapshot.
- Per-IP rate limiting on scan submission.
- CORS enabled for browser API access.
- Browse endpoints with pagination.
- Async pwhois worker with DB persistence.
- Docker Compose includes frontend service.
- Playwright QA scenarios.

### Must NOT Have (Guardrails)
- No user accounts, authentication, or authorization.
- No Redis or external message queue.
- No historical backfill of existing snapshots.
- No charts, maps, WebSockets, or email.
- No pwhois data in generated PDF reports.
- No TLS, domain, or CI/CD changes.
- No custom design system beyond shadcn/ui.

### Defaults Applied
- **Frontend-to-API**: Next.js static export served by nginx, proxying `/api` to the Go API. `NEXT_PUBLIC_API_URL=/api`. This avoids CORS configuration inside Docker Compose.
- **Client IP trust**: Gin `TrustedProxies` configured for Docker/internal networks (`172.16.0.0/12`, `10.0.0.0/8`, `192.168.0.0/16`, `127.0.0.1`). `c.ClientIP()` derives the submitter IP.
- **Existing snapshots**: `client_ip` and pwhois fields are nullable; old rows display as "unknown".
- **pwhois cache**: 24-hour TTL in worker + DB persistence; batch size 100; 3 retries with exponential backoff; skip private/loopback/invalid IPs.
- **Rate limiting**: Token bucket per normalized IP; 30 `POST /api/scan` per minute; in-memory map with periodic cleanup.
- **Data retention**: IP addresses stored indefinitely (no automatic purge).
- **Privacy notice**: Not included in the UI by default; browse pages publicly expose submitter IPs and pwhois data.
- **pwhois attribution**: Copyright notice from pwhois.org terms will be preserved in code comments and documentation.
- **Light/dark mode**: Light mode default; no toggle.

---

## Verification Strategy (MANDATORY)

> **ZERO HUMAN INTERVENTION** - ALL verification is agent-executed.

### Test Decision
- **Infrastructure exists**: YES (Go `go test`, existing `*_test.go` files).
- **Automated tests**: YES (tests-after implementation).
- **Framework**: `go test` for backend; Playwright for frontend.
- **Agent-Executed QA**: ALWAYS.

### QA Policy
Every task MUST include agent-executed QA scenarios. Evidence saved to `.omo/evidence/task-{N}-{scenario-slug}.{ext}`.

- **Backend/API**: Use `Bash` (`go test`, `curl`).
- **Frontend/UI**: Use Playwright — navigate, interact, assert DOM, screenshot.
- **Docker**: Use `Bash` (`docker compose`).

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 1 (Backend foundation — start immediately):
├── 1. Add client_ip to snapshots + model + scan handler [quick]
├── 2. Add per-IP rate limiting middleware [quick]
├── 3. Add CORS + trusted proxy config [quick]
├── 4. Add browse API endpoints (targets, snapshots, reports) [unspecified-high]
├── 5. Add pwhois DB migration + models [quick]
└── 6. Add API contract tests for browse + rate limit [quick]

Wave 2 (pwhois worker — after Wave 1):
├── 7. Create internal/pwhois worker [unspecified-high]
├── 8. Wire pwhois worker into main.go [quick]
└── 9. Add pwhois worker tests [quick]

Wave 3 (Frontend — after Wave 1):
├── 10. Initialize Next.js + shadcn/ui in frontend/ [quick]
├── 11. Create shared layout + navigation [visual-engineering]
├── 12. Build landing page with scan form [visual-engineering]
├── 13. Build browse targets page [visual-engineering]
├── 14. Build browse snapshots page [visual-engineering]
├── 15. Build browse reports page [visual-engineering]
├── 16. Build snapshot detail page [visual-engineering]
└── 17. Build report detail/download page [visual-engineering]

Wave 4 (Integration + Docker + E2E — after Waves 2-3):
├── 18. Add frontend Dockerfile [quick]
├── 19. Update docker-compose.yml with frontend service [quick]
├── 20. Add frontend API client + env config [quick]
└── 21. Add Playwright tests and run full stack [unspecified-high]

Wave FINAL (After ALL tasks):
├── F1. Plan compliance audit (oracle)
├── F2. Code quality review (unspecified-high)
├── F3. Real manual QA — Playwright + curl (unspecified-high)
└── F4. Scope fidelity check (deep)
-> Present results -> Get explicit user okay
```

### Dependency Matrix

- **1**: - → 4, 6
- **2**: - → 6
- **3**: - → 6, 21
- **4**: 1, 5 → 6
- **5**: - → 7, 9
- **6**: 1-4 → F1-F4
- **7**: 5 → 8, 9
- **8**: 7 → 21
- **9**: 7 → F1-F4
- **10**: - → 11-17
- **11**: 10 → 12-17
- **12**: 11 → 21
- **13**: 11 → 21
- **14**: 11 → 21
- **15**: 11 → 21
- **16**: 11 → 21
- **17**: 11 → 21
- **18**: 10 → 19
- **19**: 18 → 21
- **20**: 10 → 21
- **21**: 3, 8, 12-17, 19, 20 → F1-F4
- **F1-F4**: 1-21 → user okay

### Agent Dispatch Summary

- **Wave 1**: 1 → `quick`, 2 → `quick`, 3 → `quick`, 4 → `unspecified-high`, 5 → `quick`, 6 → `quick`
- **Wave 2**: 7 → `unspecified-high`, 8 → `quick`, 9 → `quick`
- **Wave 3**: 10 → `quick`, 11-17 → `visual-engineering`
- **Wave 4**: 18 → `quick`, 19 → `quick`, 20 → `quick`, 21 → `unspecified-high`
- **FINAL**: F1 → `oracle`, F2 → `unspecified-high`, F3 → `unspecified-high`, F4 → `deep`

---

## TODOs

- [x] 1. Add `client_ip` to snapshots, update model, and capture IP in scan handler

  **What to do**:
  - Add `client_ip TEXT` column to `snapshots` table in `internal/db/db.go:Migrate()`.
  - Add `ClientIP string` field to `models.Snapshot` in `internal/models/models.go`.
  - In `internal/handlers/scan.go:createScan`, capture `c.ClientIP()` after binding and pass it to `storeSnapshot`.
  - Update `storeSnapshot` signature and INSERT/UPDATE SQL to include `client_ip`.
  - Update `upsertTarget` and `storeSnapshot` tests in `internal/handlers/scan_test.go`.

  **Must NOT do**:
  - Do not make `client_ip` non-nullable (existing rows must remain valid).
  - Do not store IP on `targets` or `reports` tables.
  - Do not log IPs to application logs.

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: none required

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 2, 3, 4, 5, 6)
  - **Blocks**: 4, 6
  - **Blocked By**: None

  **References**:
  - `internal/db/db.go:47-99` — existing `Migrate()` schema.
  - `internal/models/models.go:17-25` — `Snapshot` struct.
  - `internal/handlers/scan.go:60-95` — `createScan` flow.
  - `internal/handlers/scan.go:112-182` — `storeSnapshot` SQL.
  - `internal/handlers/scan_test.go` — existing handler test patterns.

  **Acceptance Criteria**:
  - [ ] `go test ./internal/handlers/... -run TestStoreSnapshot -count=1` passes.
  - [ ] `go test ./internal/db/... -run TestMigrate_CreatesTablesAndIsIdempotent -count=1` passes.
  - [ ] A scan request from `127.0.0.1` stores `client_ip` = `127.0.0.1`.

  **QA Scenarios**:

  ```
  Scenario: Scan stores submitter IP
    Tool: Bash (curl + psql)
    Preconditions: Docker Compose stack running
    Steps:
      1. Run `curl -s -X POST http://localhost:8080/api/scan -H "Content-Type: application/json" -d '{"host":"example.com"}'`
      2. Query DB: `docker exec echostate-db psql -U echostate -d echostate -c "SELECT client_ip FROM snapshots ORDER BY scanned_at DESC LIMIT 1;"`
    Expected Result: `client_ip` is non-null and matches the curl source IP.
    Evidence: .omo/evidence/task-1-client-ip.txt

  Scenario: Existing snapshots remain valid with null client_ip
    Tool: Bash (go test)
    Preconditions: Migration applied before this code change
    Steps:
      1. Run `go test ./internal/db/... -run TestMigrate_CreatesTablesAndIsIdempotent -count=1 -v`
    Expected Result: Test passes; nullable column does not break existing rows.
    Evidence: .omo/evidence/task-1-migration.txt
  ```

  **Evidence to Capture**:
  - [ ] DB query showing stored `client_ip`.
  - [ ] Migration test output.

  **Commit**: YES
  - Message: `feat(db,handlers): capture and store submitter IP on each snapshot`
  - Files: `internal/db/db.go`, `internal/models/models.go`, `internal/handlers/scan.go`, `internal/handlers/scan_test.go`

- [x] 2. Add per-IP rate limiting middleware

  **What to do**:
  - Create `internal/middleware/ratelimit.go` with a token-bucket rate limiter keyed by normalized client IP.
  - Limit: 30 `POST /api/scan` requests per IP per minute.
  - Use in-memory map with RWMutex and periodic cleanup (no Redis).
  - Return HTTP 429 with JSON body `{"error":"rate limit exceeded","retry_after":N}`.
  - Apply only to `POST /api/scan` in `internal/handlers/scan.go:Register()`.
  - Add tests for allowed/blocked requests and cleanup.

  **Must NOT do**:
  - Do not apply rate limiting to browse/read endpoints.
  - Do not use Redis or external storage.
  - Do not rate limit by target host instead of IP.

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: none required

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 3, 4, 5, 6)
  - **Blocks**: 6
  - **Blocked By**: None

  **References**:
  - `internal/handlers/scan.go:37-54` — route registration.
  - `main.go:41-43` — existing middleware pattern.
  - Standard library `net/netip` for IP normalization.

  **Acceptance Criteria**:
  - [ ] `go test ./internal/middleware/... -count=1` passes.
  - [ ] 30 rapid scan requests from same IP succeed.
  - [ ] 31st request returns HTTP 429.

  **QA Scenarios**:

  ```
  Scenario: Rate limit allows 30 scans per minute
    Tool: Bash (curl loop)
    Preconditions: API running
    Steps:
      1. Run a loop of 30 `POST /api/scan` from the same IP.
      2. Assert all return 200/202.
    Expected Result: All 30 requests succeed.
    Evidence: .omo/evidence/task-2-rate-limit-allowed.txt

  Scenario: Rate limit blocks 31st scan
    Tool: Bash (curl)
    Preconditions: 30 scans already submitted in the same minute
    Steps:
      1. Run 31st `POST /api/scan`.
    Expected Result: HTTP 429 with JSON error and retry_after field.
    Evidence: .omo/evidence/task-2-rate-limit-blocked.txt
  ```

  **Evidence to Capture**:
  - [ ] Loop output showing 30 successes.
  - [ ] 31st request 429 response.

  **Commit**: YES
  - Message: `feat(api): add per-IP rate limiting to scan submissions`
  - Files: `internal/middleware/ratelimit.go`, `internal/middleware/ratelimit_test.go`, `internal/handlers/scan.go`

- [x] 3. Add CORS and trusted proxy configuration

  **What to do**:
  - Add `github.com/gin-contrib/cors` dependency.
  - Create `internal/middleware/cors.go` allowing the frontend origin (`http://localhost:3000` by default, configurable via `FRONTEND_URL`).
  - Configure Gin `TrustedProxies` in `main.go` to trust Docker/internal networks (`172.16.0.0/12`, `10.0.0.0/8`, `192.168.0.0/16`, `127.0.0.1`).
  - Add `FRONTEND_URL` to `internal/config/config.go` and `.env.example`.
  - Add tests verifying CORS headers on API responses.

  **Must NOT do**:
  - Do not allow `*` origins in production config.
  - Do not trust all proxies (`SetTrustedProxies(nil)`) in production.
  - Do not add CORS to non-API routes if not needed.

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: none required

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 2, 4, 5, 6)
  - **Blocks**: 21
  - **Blocked By**: None

  **References**:
  - `main.go:41-43` — middleware setup.
  - `internal/config/config.go` — config loading pattern.
  - `.env.example` — existing env vars.
  - `https://github.com/gin-contrib/cors` — CORS middleware docs.

  **Acceptance Criteria**:
  - [ ] `go test ./internal/middleware/... -count=1` passes.
  - [ ] Preflight `OPTIONS` request to `/api/scan` returns `Access-Control-Allow-Origin` header.
  - [ ] `gin.TrustedProxies` includes Docker networks.

  **QA Scenarios**:

  ```
  Scenario: CORS preflight succeeds for frontend origin
    Tool: Bash (curl)
    Preconditions: API running with FRONTEND_URL=http://localhost:3000
    Steps:
      1. Run `curl -s -i -X OPTIONS http://localhost:8080/api/scan -H "Origin: http://localhost:3000" -H "Access-Control-Request-Method: POST"`
    Expected Result: HTTP 204 with Access-Control-Allow-Origin: http://localhost:3000.
    Evidence: .omo/evidence/task-3-cors-preflight.txt

  Scenario: Untrusted origin is not reflected
    Tool: Bash (curl)
    Preconditions: API running
    Steps:
      1. Run `curl -s -i -X OPTIONS http://localhost:8080/api/scan -H "Origin: http://evil.example"`
    Expected Result: No Access-Control-Allow-Origin header (or not matching evil.example).
    Evidence: .omo/evidence/task-3-cors-untrusted.txt
  ```

  **Evidence to Capture**:
  - [ ] Preflight response headers.
  - [ ] Untrusted origin response.

  **Commit**: YES
  - Message: `feat(api): enable CORS and configure trusted proxies for frontend`
  - Files: `internal/middleware/cors.go`, `internal/middleware/cors_test.go`, `internal/config/config.go`, `.env.example`, `main.go`

- [x] 4. Add browse API endpoints

  **What to do**:
  - Create `internal/handlers/browse.go` with handler methods:
    - `GET /api/targets` — list targets with `page`, `limit`, `q` (host search).
    - `GET /api/targets/:id/snapshots` — list snapshots for a target.
    - `GET /api/snapshots` — list all snapshots with `page`, `limit`, `target_id`.
    - `GET /api/snapshots/:id` — get snapshot detail including `raw_data`, `changes`, `client_ip`, `pwhois_data`.
    - `GET /api/reports` — list reports with `page`, `limit`, `snapshot_id`, `status`.
  - Add `PaginatedResponse` and summary models to `internal/models/models.go`.
  - Use `COALESCE` for `client_ip` fallback to `"unknown"`.
  - Default pagination: page=1, limit=20, max limit=100.
  - Add tests for each endpoint.

  **Must NOT do**:
  - Do not make browse endpoints writeable.
  - Do not expose raw PDF bytes in list endpoints.
  - Do not allow unbounded `limit` params.

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: none required

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 2, 3, 5, 6)
  - **Blocks**: 6
  - **Blocked By**: 1, 5

  **References**:
  - `internal/handlers/scan.go:37-54` — route registration pattern.
  - `internal/handlers/reports.go` — existing read handler patterns.
  - `internal/models/models.go` — existing model patterns.
  - `internal/handlers/reports_test.go` — DB seeding and httptest patterns.

  **Acceptance Criteria**:
  - [ ] `go test ./internal/handlers/... -count=1` passes.
  - [ ] `GET /api/targets` returns paginated list.
  - [ ] `GET /api/snapshots/:id` includes `client_ip` and `pwhois_data`.

  **QA Scenarios**:

  ```
  Scenario: Browse targets returns paginated results
    Tool: Bash (curl)
    Preconditions: API running with seeded data
    Steps:
      1. Run `curl -s "http://localhost:8080/api/targets?page=1&limit=5"`
    Expected Result: HTTP 200 with JSON containing data, page, limit, total_count.
    Evidence: .omo/evidence/task-4-browse-targets.txt

  Scenario: Snapshot detail includes client IP and pwhois
    Tool: Bash (curl)
    Preconditions: Snapshot with client_ip and pwhois enrichment exists
    Steps:
      1. Run `curl -s "http://localhost:8080/api/snapshots/<snapshot_id>"`
    Expected Result: HTTP 200 with client_ip and pwhois_data fields.
    Evidence: .omo/evidence/task-4-snapshot-detail.txt
  ```

  **Evidence to Capture**:
  - [ ] Targets list response.
  - [ ] Snapshot detail response.

  **Commit**: YES
  - Message: `feat(api): add paginated browse endpoints for targets, snapshots, and reports`
  - Files: `internal/handlers/browse.go`, `internal/handlers/browse_test.go`, `internal/models/models.go`, `internal/handlers/scan.go`

- [x] 5. Add pwhois DB migration and models

  **What to do**:
  - Add to `internal/db/db.go:Migrate()`:
    - `pwhois_data JSONB` on `snapshots`.
    - `pwhois_looked_up_at TIMESTAMPTZ` on `snapshots`.
    - Extracted key fields: `pwhois_origin_as TEXT`, `pwhois_org_name TEXT`, `pwhois_country_code TEXT`, `pwhois_city TEXT`, `pwhois_prefix TEXT`.
    - Index on `snapshots(pwhois_country_code)` and `snapshots(pwhois_origin_as)`.
  - Add `PWHOISData` struct and fields to `models.Snapshot` in `internal/models/models.go`.
  - Update `internal/db/db_test.go` to assert new columns/indexes exist.

  **Must NOT do**:
  - Do not create a separate `pwhois_lookups` table (keep denormalized on snapshots).
  - Do not make pwhois fields non-nullable.
  - Do not add pwhois fields to `ScanResult` or PDF renderer.

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: none required

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1-4, 6)
  - **Blocks**: 7, 9
  - **Blocked By**: None

  **References**:
  - `internal/db/db.go:47-99` — existing `Migrate()`.
  - `internal/models/models.go` — model patterns.
  - `internal/db/db_test.go` — schema verification tests.

  **Acceptance Criteria**:
  - [ ] `go test ./internal/db/... -count=1` passes.
  - [ ] `pwhois_data`, `pwhois_looked_up_at`, and key fields exist on `snapshots`.

  **QA Scenarios**:

  ```
  Scenario: Migration creates pwhois columns and indexes
    Tool: Bash (go test)
    Preconditions: None
    Steps:
      1. Run `go test ./internal/db/... -run TestMigrate_CreatesTablesAndIsIdempotent -count=1 -v`
    Expected Result: Test passes and asserts pwhois columns/indexes.
    Evidence: .omo/evidence/task-5-pwhois-migration.txt
  ```

  **Evidence to Capture**:
  - [ ] DB test output.

  **Commit**: YES
  - Message: `feat(db): add pwhois columns and indexes to snapshots`
  - Files: `internal/db/db.go`, `internal/db/db_test.go`, `internal/models/models.go`

- [x] 6. Add API contract tests for browse and rate limit

  **What to do**:
  - Add integration-style tests in `internal/handlers/browse_test.go` for:
    - Pagination defaults and bounds.
    - Filtering snapshots by `target_id`.
    - Filtering reports by `status`.
    - 404 for missing target/snapshot/report.
  - Add tests in `internal/handlers/scan_test.go` verifying rate limit middleware returns 429.
  - Ensure tests use isolated DB schemas (follow `reports_test.go` pattern).

  **Must NOT do**:
  - Do not require real pwhois responses (mock or null).
  - Do not test frontend here.

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: none required

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1-5)
  - **Blocks**: F1-F4
  - **Blocked By**: 1-4

  **References**:
  - `internal/handlers/reports_test.go` — DB test setup pattern.
  - `internal/handlers/scan_test.go` — mock scanner pattern.

  **Acceptance Criteria**:
  - [ ] `go test ./internal/handlers/... -count=1` passes.
  - [ ] All browse endpoints have happy-path and 404 tests.
  - [ ] Rate limit 429 test passes.

  **QA Scenarios**:

  ```
  Scenario: Browse endpoints return paginated data
    Tool: Bash (go test)
    Preconditions: None
    Steps:
      1. Run `go test ./internal/handlers/... -run TestBrowse -count=1 -v`
    Expected Result: All browse tests pass.
    Evidence: .omo/evidence/task-6-browse-tests.txt

  Scenario: Rate limit returns 429
    Tool: Bash (go test)
    Preconditions: None
    Steps:
      1. Run `go test ./internal/handlers/... -run TestRateLimit -count=1 -v`
    Expected Result: 429 response test passes.
    Evidence: .omo/evidence/task-6-ratelimit-tests.txt
  ```

  **Evidence to Capture**:
  - [ ] Handler test output.

  **Commit**: YES
  - Message: `test(api): add browse and rate limit contract tests`
  - Files: `internal/handlers/browse_test.go`, `internal/handlers/scan_test.go`

- [x] 7. Create async pwhois worker

  **What to do**:
  - Create `internal/pwhois/worker.go`:
    - `Worker` struct with `*db.DB`, `*pwhois.WhoisServer` from `github.com/georgestarcher/pwhois`, in-memory cache.
    - `Start(ctx context.Context)` polls `snapshots` for rows where `client_ip` is set but `pwhois_looked_up_at` is older than 24h or null.
    - `Enqueue(snapshotID, clientIP)` or use polling-only model.
    - Skip private/loopback/invalid IPs.
    - Batch up to 100 IPs per query.
    - On success: parse response, update `pwhois_data` + key fields + `pwhois_looked_up_at`.
    - On failure: log error, do not block scan/report; retry up to 3 times with exponential backoff.
    - Use 30s per-job timeout.
  - Create `internal/pwhois/pwhois.go` with parsing helpers if needed.
  - Add package-level `whoisServer` shim for testability.

  **Must NOT do**:
  - Do not run pwhois synchronously in scan handler.
  - Do not fail scan/report if pwhois fails.
  - Do not expose pwhois data in PDF renderer.

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: none required

  **Parallelization**:
  - **Can Run In Parallel**: NO (after Wave 1)
  - **Parallel Group**: Wave 2 (with Tasks 8, 9)
  - **Blocks**: 8, 9
  - **Blocked By**: 5

  **References**:
  - `internal/reports/worker.go` — existing worker pattern.
  - `internal/scanner/whois.go` — gatherer pattern and shim approach.
  - `github.com/georgestarcher/pwhois` — library API.

  **Acceptance Criteria**:
  - [ ] `go test ./internal/pwhois/... -count=1` passes.
  - [ ] Worker updates snapshot with parsed pwhois data.
  - [ ] Worker skips private IPs.

  **QA Scenarios**:

  ```
  Scenario: Worker enriches a public IP
    Tool: Bash (go test)
    Preconditions: None (mock pwhois server)
    Steps:
      1. Run `go test ./internal/pwhois/... -run TestWorkerEnrichesPublicIP -count=1 -v`
    Expected Result: Snapshot row updated with pwhois_data and key fields.
    Evidence: .omo/evidence/task-7-pwhois-public.txt

  Scenario: Worker skips loopback IP
    Tool: Bash (go test)
    Preconditions: None
    Steps:
      1. Run `go test ./internal/pwhois/... -run TestWorkerSkipsPrivateIP -count=1 -v`
    Expected Result: Snapshot row not updated; no external query made.
    Evidence: .omo/evidence/task-7-pwhois-private.txt
  ```

  **Evidence to Capture**:
  - [ ] Worker test output.

  **Commit**: YES
  - Message: `feat(pwhois): add async worker for IP enrichment with caching and retries`
  - Files: `internal/pwhois/worker.go`, `internal/pwhois/pwhois.go`

- [x] 8. Wire pwhois worker into main.go

  **What to do**:
  - In `main.go`, instantiate pwhois worker after DB setup.
  - Start worker with `worker.Start(ctx)` in the same goroutine pattern as the report worker.
  - Stop worker on shutdown signal.
  - Add `ECHOSTATE_PWHOIS_ENABLED` env var (default true) and `PWHOIS_CACHE_TTL_HOURS`.

  **Must NOT do**:
  - Do not start pwhois worker if disabled.
  - Do not block main thread on worker startup.

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: none required

  **Parallelization**:
  - **Can Run In Parallel**: NO (Wave 2)
  - **Parallel Group**: Wave 2 (with Tasks 7, 9)
  - **Blocks**: 21
  - **Blocked By**: 7

  **References**:
  - `main.go` — worker lifecycle pattern.
  - `internal/reports/worker.go` — how report worker is started/stopped.
  - `internal/config/config.go` — env config pattern.

  **Acceptance Criteria**:
  - [ ] `go build ./...` succeeds.
  - [ ] `go test ./... -count=1` passes.
  - [ ] Worker starts/stops cleanly in main.

  **QA Scenarios**:

  ```
  Scenario: Application starts with pwhois worker
    Tool: Bash (go run)
    Preconditions: Postgres running
    Steps:
      1. Run `ECHOSTATE_PWHOIS_ENABLED=true go run ./main.go`
      2. Wait 2s, then interrupt with Ctrl-C.
    Expected Result: Clean startup and shutdown; no panic.
    Evidence: .omo/evidence/task-8-pwhois-startup.txt
  ```

  **Evidence to Capture**:
  - [ ] Startup/shutdown logs.

  **Commit**: YES
  - Message: `feat(main): wire pwhois worker into application lifecycle`
  - Files: `main.go`, `internal/config/config.go`, `.env.example`

- [x] 9. Add pwhois worker tests

  **What to do**:
  - Create `internal/pwhois/worker_test.go`:
    - Mock `pwhois.WhoisServer` via package-level shim.
    - Test successful enrichment updates DB.
    - Test failure does not block and logs error.
    - Test cache prevents duplicate lookups within TTL.
    - Test invalid/private IP is skipped.
    - Test batching aggregates multiple pending snapshots.

  **Must NOT do**:
  - Do not make real network calls to pwhois.org in tests.
  - Do not test pwhois library internals.

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: none required

  **Parallelization**:
  - **Can Run In Parallel**: NO (Wave 2)
  - **Parallel Group**: Wave 2 (with Tasks 7, 8)
  - **Blocks**: F1-F4
  - **Blocked By**: 7

  **References**:
  - `internal/reports/worker_test.go` — worker test pattern.
  - `internal/scanner/whois_test.go` — shim/mocking pattern.

  **Acceptance Criteria**:
  - [ ] `go test ./internal/pwhois/... -count=1` passes.
  - [ ] Coverage includes success, failure, cache, private IP, batch.

  **QA Scenarios**:

  ```
  Scenario: pwhois unit tests pass
    Tool: Bash (go test)
    Preconditions: None
    Steps:
      1. Run `go test ./internal/pwhois/... -count=1 -v`
    Expected Result: All tests pass.
    Evidence: .omo/evidence/task-9-pwhois-tests.txt
  ```

  **Evidence to Capture**:
  - [ ] Test output.

  **Commit**: YES
  - Message: `test(pwhois): cover worker enrichment, caching, and error paths`
  - Files: `internal/pwhois/worker_test.go`

- [x] 10. Initialize Next.js 14+ App Router + shadcn/ui in `frontend/`

  **What to do**:
  - Run `npx shadcn@latest init --yes --template next --base-color slate` in repo root to create `frontend/`.
  - Configure `next.config.js` for static export (`output: 'export'`, `distDir: 'dist'`).
  - Add `.gitignore` for `node_modules/`, `.next/`, `dist/`.
  - Add `README.md` in `frontend/` with dev commands.
  - Verify `npm run build` succeeds and outputs to `frontend/dist/`.

  **Must NOT do**:
  - Do not use Next.js Pages Router.
  - Do not add unnecessary meta-frameworks (e.g., tRPC, Zustand, TanStack Query) unless required.
  - Do not commit `node_modules/`.

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: [`shadcn`]

  **Parallelization**:
  - **Can Run In Parallel**: YES (after Wave 1)
  - **Parallel Group**: Wave 3 (with Tasks 11-17)
  - **Blocks**: 11-17, 18, 20
  - **Blocked By**: None

  **References**:
  - `https://ui.shadcn.com/docs/installation/next` — shadcn Next.js install.
  - Existing `README.md` — project conventions.

  **Acceptance Criteria**:
  - [ ] `frontend/package.json` exists with Next.js and shadcn dependencies.
  - [ ] `npm run build` exits 0 and creates `frontend/dist/`.
  - [ ] No `node_modules/` committed.

  **QA Scenarios**:

  ```
  Scenario: Frontend builds successfully
    Tool: Bash
    Preconditions: Node 20+ installed
    Steps:
      1. Run `cd frontend && npm install`
      2. Run `cd frontend && npm run build`
    Expected Result: Build succeeds; `frontend/dist/` created.
    Evidence: .omo/evidence/task-10-frontend-build.txt
  ```

  **Evidence to Capture**:
  - [ ] Build output.

  **Commit**: YES
  - Message: `feat(frontend): initialize Next.js + shadcn/ui project`
  - Files: `frontend/`

- [x] 11. Create shared layout + navigation

  **What to do**:
  - Update `frontend/app/layout.tsx` with shadcn `ThemeProvider` (light default), project logo, and global styles.
  - Create `frontend/components/nav.tsx` with links to: Home, Targets, Snapshots, Reports.
  - Use shadcn `NavigationMenu`, `Button`, and `Card` components.
  - Add responsive mobile menu.
  - Add footer with link to GitHub repo.

  **Must NOT do**:
  - Do not create a custom design system beyond shadcn components.
  - Do not add dark mode toggle (light default only).
  - Do not fetch data in layout.

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
  - **Skills**: [`shadcn`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 10, 12-17)
  - **Blocks**: 12-17
  - **Blocked By**: 10

  **References**:
  - `assets/logo.png` and `assets/logo-banner.png` — brand assets.
  - shadcn docs for NavigationMenu, Button, Sheet (mobile menu).

  **Acceptance Criteria**:
  - [ ] Navigation renders on all pages.
  - [ ] Links navigate to `/`, `/targets`, `/snapshots`, `/reports`.
  - [ ] Mobile menu works.

  **QA Scenarios**:

  ```
  Scenario: Navigation renders and links work
    Tool: Playwright
    Preconditions: Frontend dev server running
    Steps:
      1. Navigate to `http://localhost:3000`.
      2. Click "Targets" link.
    Expected Result: URL changes to `/targets`; Targets heading visible.
    Evidence: .omo/evidence/task-11-nav.png
  ```

  **Evidence to Capture**:
  - [ ] Playwright screenshot of nav on desktop and mobile.

  **Commit**: YES
  - Message: `feat(frontend): add shared layout and navigation`
  - Files: `frontend/app/layout.tsx`, `frontend/components/nav.tsx`, `frontend/app/globals.css`

- [x] 12. Build landing page with scan form

  **What to do**:
  - Create `frontend/app/page.tsx`:
    - Hero section with logo and tagline.
    - Input field for host/IP/URL with submit button.
    - Result card showing snapshot summary (host, scanned_at, snapshot_id).
    - Link to generated report once available.
  - Create `frontend/lib/api.ts` helper for `fetch` to API (reads `NEXT_PUBLIC_API_URL`).
  - Use shadcn `Input`, `Button`, `Card`, `Alert` components.
  - Handle loading, error, and success states.

  **Must NOT do**:
  - Do not implement live progress/polling beyond basic report status check.
  - Do not add authentication gating.

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
  - **Skills**: [`shadcn`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 10-11, 13-17)
  - **Blocks**: 21
  - **Blocked By**: 11

  **References**:
  - shadcn docs for Input, Button, Card, Alert.
  - `internal/handlers/scan.go` — scan response shape.

  **Acceptance Criteria**:
  - [ ] Landing page renders.
  - [ ] Submitting a host returns snapshot summary.
  - [ ] Rate limit 429 shows user-friendly error.

  **QA Scenarios**:

  ```
  Scenario: Submit scan from landing page
    Tool: Playwright
    Preconditions: API and frontend running
    Steps:
      1. Navigate to `http://localhost:3000`.
      2. Enter "example.com" in host input.
      3. Click "Scan".
    Expected Result: Result card appears with snapshot ID and link to report.
    Evidence: .omo/evidence/task-12-landing-scan.png

  Scenario: Rate limit shows error
    Tool: Playwright
    Preconditions: API running, rate limit configured
    Steps:
      1. Submit 31 scans rapidly via API.
      2. Navigate to landing and submit one more.
    Expected Result: Error alert displays "Too many scans. Try again in N seconds."
    Evidence: .omo/evidence/task-12-rate-limit-ui.png
  ```

  **Evidence to Capture**:
  - [ ] Playwright screenshots of success and error states.

  **Commit**: YES
  - Message: `feat(frontend): landing page with scan submission form`
  - Files: `frontend/app/page.tsx`, `frontend/lib/api.ts`, `frontend/components/scan-form.tsx`

- [x] 13. Build browse targets page

  **What to do**:
  - Create `frontend/app/targets/page.tsx`:
    - Table listing targets with host, normalized_host, created_at, scan count.
    - Search input filtering by host.
    - Pagination controls.
    - Link to target snapshots page.
  - Use shadcn `Table`, `Input`, `Button`, `Pagination` components.
  - Server Component fetching `GET /api/targets`.

  **Must NOT do**:
  - Do not add charts or maps.
  - Do not expose client IPs on this page.

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
  - **Skills**: [`shadcn`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 10-12, 14-17)
  - **Blocks**: 21
  - **Blocked By**: 11

  **References**:
  - shadcn docs for Table, Pagination.
  - `internal/handlers/browse.go` — targets response shape.

  **Acceptance Criteria**:
  - [ ] Targets page renders paginated table.
  - [ ] Search filters results.
  - [ ] Clicking a target navigates to snapshots.

  **QA Scenarios**:

  ```
  Scenario: Browse targets with pagination
    Tool: Playwright
    Preconditions: API seeded with targets
    Steps:
      1. Navigate to `http://localhost:3000/targets`.
      2. Assert table shows targets.
      3. Click page 2 if available.
    Expected Result: Table updates; pagination state changes.
    Evidence: .omo/evidence/task-13-targets.png
  ```

  **Evidence to Capture**:
  - [ ] Playwright screenshot.

  **Commit**: YES
  - Message: `feat(frontend): targets browse page with search and pagination`
  - Files: `frontend/app/targets/page.tsx`, `frontend/components/targets-table.tsx`

- [x] 14. Build browse snapshots page

  **What to do**:
  - Create `frontend/app/snapshots/page.tsx`:
    - Table listing snapshots with host, scanned_at, client_ip, pwhois country, pwhois org.
    - Filters: target_id dropdown, country code.
    - Pagination.
    - Link to snapshot detail.
  - Create `frontend/components/snapshots-table.tsx`.
  - Display "unknown" for null client_ip.

  **Must NOT do**:
  - Do not expose raw pwhois JSON in table.
  - Do not show pwhois if not yet enriched.

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
  - **Skills**: [`shadcn`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 10-13, 15-17)
  - **Blocks**: 21
  - **Blocked By**: 11

  **References**:
  - shadcn docs for Table, Select (filters), Pagination.
  - `internal/handlers/browse.go` — snapshots response shape.

  **Acceptance Criteria**:
  - [ ] Snapshots page renders.
  - [ ] client_ip and pwhois summary columns visible.
  - [ ] Filters work.

  **QA Scenarios**:

  ```
  Scenario: Browse snapshots with IP and pwhois
    Tool: Playwright
    Preconditions: Snapshots with client_ip and pwhois exist
    Steps:
      1. Navigate to `http://localhost:3000/snapshots`.
      2. Assert table shows client_ip and pwhois country/org.
    Expected Result: Table displays enrichment data.
    Evidence: .omo/evidence/task-14-snapshots.png
  ```

  **Evidence to Capture**:
  - [ ] Playwright screenshot.

  **Commit**: YES
  - Message: `feat(frontend): snapshots browse page with IP and pwhois filters`
  - Files: `frontend/app/snapshots/page.tsx`, `frontend/components/snapshots-table.tsx`

- [x] 15. Build browse reports page

  **What to do**:
  - Create `frontend/app/reports/page.tsx`:
    - Table listing reports with host, status, created_at, completed_at.
    - Filters: status, snapshot_id.
    - Pagination.
    - Download button for completed reports.
    - Link to report detail.
  - Create `frontend/components/reports-table.tsx`.

  **Must NOT do**:
  - Do not expose PDF bytes inline.
  - Do not allow deleting reports.

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
  - **Skills**: [`shadcn`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 10-14, 16-17)
  - **Blocks**: 21
  - **Blocked By**: 11

  **References**:
  - shadcn docs for Table, Badge (status), Pagination.
  - `internal/handlers/browse.go` — reports response shape.

  **Acceptance Criteria**:
  - [ ] Reports page renders.
  - [ ] Status badges display correctly.
  - [ ] Download link triggers PDF download.

  **QA Scenarios**:

  ```
  Scenario: Browse reports and download PDF
    Tool: Playwright
    Preconditions: Completed report exists
    Steps:
      1. Navigate to `http://localhost:3000/reports`.
      2. Click download button on a completed report.
    Expected Result: Browser downloads PDF starting with %PDF.
    Evidence: .omo/evidence/task-15-reports.png
  ```

  **Evidence to Capture**:
  - [ ] Playwright screenshot and downloaded file check.

  **Commit**: YES
  - Message: `feat(frontend): reports browse page with status and download`
  - Files: `frontend/app/reports/page.tsx`, `frontend/components/reports-table.tsx`

- [x] 16. Build snapshot detail page

  **What to do**:
  - Create `frontend/app/snapshots/[id]/page.tsx`:
    - Header: host, scanned_at, client_ip, pwhois summary card.
    - Sections: WHOIS, ASN, Web, Errors, Changes (from `raw_data`).
    - pwhois full data rendered as key-value list.
    - Button to create/view report for this snapshot.
  - Use shadcn `Card`, `Tabs`, `Badge`, `Separator` components.

  **Must NOT do**:
  - Do not include pwhois data in report generation.
  - Do not expose internal error stack traces.

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
  - **Skills**: [`shadcn`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 10-15, 17)
  - **Blocks**: 21
  - **Blocked By**: 11

  **References**:
  - shadcn docs for Card, Tabs, Badge.
  - `internal/handlers/browse.go` — snapshot detail response shape.

  **Acceptance Criteria**:
  - [ ] Snapshot detail renders all sections.
  - [ ] pwhois data visible when enriched.
  - [ ] Report button works.

  **QA Scenarios**:

  ```
  Scenario: View snapshot detail
    Tool: Playwright
    Preconditions: Snapshot with pwhois exists
    Steps:
      1. Navigate to `http://localhost:3000/snapshots/<id>`.
      2. Assert host, client_ip, and pwhois country visible.
    Expected Result: All sections render correctly.
    Evidence: .omo/evidence/task-16-snapshot-detail.png
  ```

  **Evidence to Capture**:
  - [ ] Playwright screenshot.

  **Commit**: YES
  - Message: `feat(frontend): snapshot detail page with recon data and pwhois`
  - Files: `frontend/app/snapshots/[id]/page.tsx`, `frontend/components/snapshot-detail.tsx`

- [x] 17. Build report detail / download page

  **What to do**:
  - Create `frontend/app/reports/[id]/page.tsx`:
    - Header: report status, host, created_at, completed_at.
    - Status badge (pending/running/completed/failed).
    - Download button (disabled unless completed).
    - Auto-refresh status every 5s while pending/running.
    - Error message if failed.
  - Use shadcn `Card`, `Button`, `Badge`, `Skeleton` components.

  **Must NOT do**:
  - Do not poll faster than every 5 seconds.
  - Do not render PDF inline.

  **Recommended Agent Profile**:
  - **Category**: `visual-engineering`
  - **Skills**: [`shadcn`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 3 (with Tasks 10-16)
  - **Blocks**: 21
  - **Blocked By**: 11

  **References**:
  - shadcn docs for Card, Button, Badge, Skeleton.
  - `internal/handlers/reports.go` — report response shape.

  **Acceptance Criteria**:
  - [ ] Report detail renders status.
  - [ ] Download button works for completed reports.
  - [ ] Polling stops when completed/failed.

  **QA Scenarios**:

  ```
  Scenario: View report detail and download
    Tool: Playwright
    Preconditions: Completed report exists
    Steps:
      1. Navigate to `http://localhost:3000/reports/<id>`.
      2. Click download.
    Expected Result: PDF downloads; status badge shows completed.
    Evidence: .omo/evidence/task-17-report-detail.png
  ```

  **Evidence to Capture**:
  - [ ] Playwright screenshot and downloaded file check.

  **Commit**: YES
  - Message: `feat(frontend): report detail page with status polling and download`
  - Files: `frontend/app/reports/[id]/page.tsx`, `frontend/components/report-detail.tsx`

- [x] 18. Add frontend Dockerfile

  **What to do**:
  - Create `frontend/Dockerfile`:
    - Multi-stage build: `node:20-alpine` for install/build, `nginx:alpine` for static serve.
    - Copy `frontend/dist/` to nginx html dir.
    - Expose port 80.
    - Add `frontend/nginx.conf` to proxy `/api` to `http://api:8080` and serve static files.
  - Ensure build uses `NEXT_PUBLIC_API_URL=/api` so browser hits same-origin proxy.

  **Must NOT do**:
  - Do not use `next start` in production (static export + nginx is simpler).
  - Do not run Node in production stage.

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: none required

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4 (with Tasks 19-21)
  - **Blocks**: 21
  - **Blocked By**: 10

  **References**:
  - `Dockerfile` — existing Go API Dockerfile pattern.
  - `docker-compose.yml` — service definitions.

  **Acceptance Criteria**:
  - [ ] `docker build -t echostate-frontend ./frontend` succeeds.
  - [ ] Container serves `index.html` on port 80.

  **QA Scenarios**:

  ```
  Scenario: Frontend Docker image builds
    Tool: Bash
    Preconditions: None
    Steps:
      1. Run `docker build -t echostate-frontend ./frontend`
    Expected Result: Build succeeds.
    Evidence: .omo/evidence/task-18-docker-build.txt
  ```

  **Evidence to Capture**:
  - [ ] Docker build output.

  **Commit**: YES
  - Message: `chore(docker): add production Dockerfile for frontend`
  - Files: `frontend/Dockerfile`, `frontend/nginx.conf`

- [x] 19. Update docker-compose.yml with frontend service

  **What to do**:
  - Add `frontend` service to `docker-compose.yml`:
    - Build context `./frontend`.
    - Port `${FRONTEND_PORT:-3000}:80`.
    - `depends_on` api.
    - Healthcheck via `wget --spider http://localhost/`.
  - Add `FRONTEND_PORT` to `.env.example`.
  - Update `api` healthcheck if missing.
  - Ensure `api` service name is reachable as `http://api:8080` from nginx.

  **Must NOT do**:
  - Do not expose the frontend on the same port as the API.
  - Do not make frontend depend on browser service.

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: none required

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4 (with Tasks 18, 20, 21)
  - **Blocks**: 21
  - **Blocked By**: 18

  **References**:
  - `docker-compose.yml` — existing services.
  - `.env.example` — env vars.

  **Acceptance Criteria**:
  - [ ] `docker compose config` validates.
  - [ ] `docker compose up --build -d` starts frontend, api, db, browser.
  - [ ] Frontend reachable at `http://localhost:3000`.

  **QA Scenarios**:

  ```
  Scenario: Full stack starts with frontend
    Tool: Bash
    Preconditions: Docker running
    Steps:
      1. Run `docker compose down -v`
      2. Run `docker compose up --build -d`
      3. Run `docker compose ps`
    Expected Result: All services healthy/running.
    Evidence: .omo/evidence/task-19-compose-ps.txt
  ```

  **Evidence to Capture**:
  - [ ] `docker compose ps` output.

  **Commit**: YES
  - Message: `chore(docker): add frontend service to compose stack`
  - Files: `docker-compose.yml`, `.env.example`

- [x] 20. Add frontend API client and environment config

  **What to do**:
  - Create `frontend/lib/api.ts`:
    - `fetchApi(path, options)` helper.
    - Reads `NEXT_PUBLIC_API_URL` env var.
    - In Docker, uses `/api` (same-origin nginx proxy).
    - In dev, uses `http://localhost:8080`.
    - Handles JSON parsing and error messages.
  - Create `frontend/lib/types.ts` with TypeScript interfaces mirroring Go models.
  - Add `.env.local.example` documenting required vars.
  - Update all frontend pages to use `fetchApi`.

  **Must NOT do**:
  - Do not hardcode API URLs in pages.
  - Do not add complex state management.

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: [`shadcn`]

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 4 (with Tasks 18, 19, 21)
  - **Blocks**: 21
  - **Blocked By**: 10

  **References**:
  - `internal/models/models.go` — response shapes.
  - Next.js docs for environment variables.

  **Acceptance Criteria**:
  - [ ] `npm run build` succeeds with env vars.
  - [ ] Frontend calls API successfully in Docker.

  **QA Scenarios**:

  ```
  Scenario: Frontend calls API through nginx proxy
    Tool: Bash (curl)
    Preconditions: Docker Compose running
    Steps:
      1. Run `curl -s http://localhost:3000/api/health`
    Expected Result: Returns API health JSON.
    Evidence: .omo/evidence/task-20-api-proxy.txt
  ```

  **Evidence to Capture**:
  - [ ] Proxy response.

  **Commit**: YES
  - Message: `feat(frontend): add API client and environment config`
  - Files: `frontend/lib/api.ts`, `frontend/lib/types.ts`, `frontend/.env.local.example`

- [x] 21. Add Playwright tests and run full stack

  **What to do**:
  - Install Playwright in `frontend/` (`npm init playwright@latest`).
  - Create `frontend/e2e/smoke.spec.ts`:
    - Navigate to `/`, assert logo and nav.
    - Submit scan, assert result card.
    - Navigate to `/targets`, assert table.
    - Navigate to `/snapshots`, assert client_ip column.
    - Navigate to `/reports`, assert status badges.
  - Create `frontend/e2e/report.spec.ts`:
    - Submit scan, create report, poll UI until download enabled, download PDF.
  - Configure Playwright to start Docker Compose before tests.
  - Run tests and capture screenshots/traces.

  **Must NOT do**:
  - Do not rely on external live websites for tests (use `example.com` or local test host).
  - Do not commit Playwright binaries.

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: [`playwright`]

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 4
  - **Blocks**: F1-F4
  - **Blocked By**: 3, 8, 12-17, 19, 20

  **References**:
  - Playwright docs for web-first assertions, traces, screenshots.
  - `.omo/plans/pdf-report-endpoint.md` — existing Docker Compose E2E pattern.

  **Acceptance Criteria**:
  - [ ] `npx playwright test` passes.
  - [ ] Screenshots/traces saved to `.omo/evidence/`.
  - [ ] All 6 frontend pages covered.

  **QA Scenarios**:

  ```
  Scenario: Full stack smoke test
    Tool: Playwright
    Preconditions: Docker Compose running
    Steps:
      1. Run `cd frontend && npx playwright test`
    Expected Result: All tests pass.
    Evidence: .omo/evidence/task-21-playwright-report/

  Scenario: End-to-end scan and report download
    Tool: Playwright
    Preconditions: Docker Compose running
    Steps:
      1. Submit scan on landing page.
      2. Create report from snapshot detail.
      3. Wait for completed status.
      4. Download PDF.
    Expected Result: PDF file starts with %PDF.
    Evidence: .omo/evidence/task-21-e2e-download.pdf
  ```

  **Evidence to Capture**:
  - [ ] Playwright HTML report.
  - [ ] Screenshots for each page.
  - [ ] Downloaded PDF.

  **Commit**: YES
  - Message: `test(e2e): add Playwright tests for full stack`
  - Files: `frontend/e2e/`, `frontend/playwright.config.ts`

---

## Final Verification Wave (MANDATORY — after ALL implementation tasks)

> 4 review agents run in PARALLEL. ALL must APPROVE. Present consolidated results to user and get explicit "okay" before completing.

- [x] F1. **Plan Compliance Audit** — `oracle`
  Read the plan end-to-end. For each "Must Have": verify implementation exists (read file, curl endpoint, run command). For each "Must NOT Have": search codebase for forbidden patterns — reject with file:line if found. Check evidence files exist in `.omo/evidence/`. Compare deliverables against plan.
  Output: `Must Have [N/N] | Must NOT Have [N/N] | Tasks [N/N] | VERDICT: APPROVE/REJECT`

- [x] F2. **Code Quality Review** — `unspecified-high`
  Run `go vet ./...` + `go test ./...` + `go build ./...` + `npm run lint` (frontend). Review for `as any`/`@ts-ignore`, empty catches, console.log in prod, commented-out code, unused imports. Check AI slop: excessive comments, over-abstraction, generic names.
  Output: `Build [PASS/FAIL] | Lint [PASS/FAIL] | Tests [N pass/N fail] | Files [N clean/N issues] | VERDICT`

- [x] F3. **Real Manual QA** — `unspecified-high` (+ `playwright` skill)
  Start from clean state. Execute EVERY QA scenario from EVERY task — follow exact steps, capture evidence. Run Playwright against Docker Compose. Test cross-task integration. Save to `.omo/evidence/final-qa/`.
  Output: `Scenarios [N/N pass] | Integration [N/N] | Edge Cases [N tested] | VERDICT`

- [x] F4. **Scope Fidelity Check** — `deep`
  For each task: read "What to do", read actual diff. Verify 1:1 — everything in spec was built, nothing beyond spec was built. Check "Must NOT do" compliance. Detect cross-task contamination.
  Output: `Tasks [N/N compliant] | Contamination [CLEAN/N issues] | Unaccounted [CLEAN/N files] | VERDICT`

---

## Commit Strategy

- Use conventional commit prefixes (`feat:`, `fix:`, `test:`, `chore:`, `docs:`).
- Keep messages natural and concise.
- Example commits:
  - `feat(db): add client_ip and pwhois columns to snapshots`
  - `feat(api): add browse endpoints for targets, snapshots, reports`
  - `feat(api): per-IP rate limiting on scan submissions`
  - `feat(pwhois): async worker for IP enrichment`
  - `feat(frontend): initialize Next.js + shadcn/ui`
  - `feat(frontend): landing page scan form`
  - `feat(frontend): browse targets/snapshots/reports pages`
  - `chore(docker): add frontend service to compose`
  - `test(e2e): Playwright verification of full stack`

---

## Success Criteria

### Verification Commands
```bash
# Backend
go test ./... -count=1
go vet ./...
go build ./...

# Frontend
cd frontend && npm run build

# Docker Compose
docker compose up --build -d

# Playwright
cd frontend && npx playwright test
```

### Final Checklist
- [ ] All frontend pages render and navigate correctly.
- [ ] `POST /api/scan` stores `client_ip` and rate limits at 30/min.
- [ ] Browse endpoints return paginated data.
- [ ] pwhois worker enriches snapshot IPs asynchronously.
- [ ] Playwright tests pass against Docker Compose.
- [ ] No auth, Redis, charts, WebSockets, or pwhois-in-PDF code added.
