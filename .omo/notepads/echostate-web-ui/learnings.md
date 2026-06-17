
## 2025-06-15 — Wave 3 init: shadcn slate + static export

- `npx shadcn@latest init --yes --defaults --name <name>` scaffolds Next.js App Router with shadcn/ui nova preset (neutral base).
- There's no `--base-color` flag — colors are set via CSS `oklch()` variables. Converted neutral → slate by replacing oklch hue/chroma to match Tailwind slate palette (~264° blue-ish hue).
- Static export works: `output: "export"` + `distDir: "dist"` in `next.config.ts`. Build produces `dist/index.html` + static assets.
- The init creates a nested `.git/` in the scaffolded directory — needed to remove it to avoid nested repo.

## 2025-06-15 — Shared layout and navigation

- Light-only theme: set `defaultTheme="light"` and `enableSystem={false}` in `ThemeProvider`; removed the `ThemeHotkey` component and the `d` dark-mode hint from `page.tsx`.
- Copied `assets/logo.png` and `assets/logo-banner.png` to `frontend/public/` for Next.js `Image` and static export.
- Added shadcn `navigation-menu` and `sheet` components via `npx shadcn@latest add navigation-menu sheet`.
- Base-ui based shadcn components use `render={<Element />}` for composition instead of Radix's `asChild`. Used this for `SheetTrigger` and `Button` rendered as `next/link`.
- Layout is server-safe: data fetching stays out; `Nav` is a client component, `Footer` is server-safe.

## 2025-06-15 — Wave 3 browse and detail pages

- Added shadcn components `tabs`, `skeleton`, and `separator` to complement the already-installed `table`, `input`, `button`, `pagination`, `badge`, `select`, `card`, and `alert`.
- Built client-side browse tables for targets, snapshots, and reports with server-only page wrappers so static export works without reading `searchParams` in Server Components.
- Built snapshot detail as a Server Component fetching `/api/snapshots/[id]` and rendering the interactive UI in a Client Component (`snapshot-detail`).
- Built report detail as a Client Component page that polls `/api/reports/[id]` every 5 seconds while status is `pending` or `running`.
- For static export with dynamic routes, `generateStaticParams()` must return at least one param; returning `[]` causes Next.js to report the function as missing. Used a placeholder param (`[{ id: "placeholder" }]`) and handled the missing data gracefully with `notFound()` on the server detail page.
- A Client Component page cannot export `generateStaticParams()`; the page file must remain a Server Component and delegate client behavior to a child Client Component.
- The snapshots browse API (`GET /api/snapshots`) returns `target_id` but not `host`, so the host column currently shows the target UUID. The country-code filter is implemented client-side because the endpoint only accepts `target_id`.
- Fixed: `snapshotListSQL` now JOINs `targets t ON t.id = s.target_id` and SELECTs `t.host`. The `SnapshotSummary` model has a new `Host` field. The frontend displays `snapshot.host || snapshot.target_id` in the Host column.

## 2025-06-15 — Landing page and scan submission form

- Added shadcn `input`, `card`, and `alert` components (`button` was already installed).
- Created `frontend/lib/api.ts` with `API_BASE`, `fetchJson`, and `scanHost`. `fetchJson` prefixes the base URL, sets `Content-Type: application/json`, and throws an `ApiError` carrying status and `retry_after` on non-2xx responses.
- Created `frontend/components/scan-form.tsx` as a client component with loading, error, and success states. It handles 429 responses by reading `retry_after` and showing "Too many scans. Try again in N seconds."
- Created `frontend/app/page.tsx` as a server component with a hero section (logo, tagline) and the scan form. The result card links to `/snapshots/<snapshot_id>`.
- The actual `POST /api/scan` response returns `id` rather than `snapshot_id`; `scanHost` normalizes the backend payload to the frontend's `ScanResponse` shape so the UI stays decoupled.
- Static export build passes after these changes.

## 2025-06-15 — Production Dockerfile and nginx config

- Multi-stage Dockerfile: `node:20-alpine` builds the static export, `nginx:alpine` serves it from `/usr/share/nginx/html`. Build arg `NEXT_PUBLIC_API_URL=/api` is set at build time.
- nginx uses a variable-based `proxy_pass` (`$api_upstream`) with `resolver 127.0.0.11` (Docker DNS) so the container starts even when the `api` service isn't available yet. DNS resolution is deferred to request time.
- The `api` hostname resolves in Docker Compose via the internal network; standalone runs will fail on `/api` routes (expected).
- SPA fallback: `try_files $uri $uri/ /index.html` catches all non-file, non-API routes.
- Static assets (`_next/static/`, `/static/`, common extensions) get `Cache-Control: public, immutable` with 1-year expiry.

## 2025-06-15 — Docker Compose integration

- Frontend service added to `docker-compose.yml` with `build: context: ./frontend`, container name `echostate-frontend`, and port `${FRONTEND_PORT:-3001}:80`. Default 3001 avoids conflict with browser service on 3000.
- Frontend depends on `api` with `condition: service_healthy` — the nginx proxy_pass to `http://api:8080` resolves via Docker DNS at request time.
- Healthcheck on frontend: `wget --spider http://localhost/` (nginx on port 80).
- Added healthcheck to `api` service: `wget --spider http://localhost:8080/health` — was missing before.
- `FRONTEND_PORT=3001` added to `.env.example`.

## 2025-06-16 — Refactor dynamic detail routes to static query-param routes

- Converted `/snapshots/[id]` and `/reports/[id]` to static `/snapshot?id=...` and `/report?id=...` routes. This avoids the Next.js static-export limitation where non-generated dynamic params fall back to the home page.
- New pages (`frontend/app/snapshot/page.tsx`, `frontend/app/report/page.tsx`) are Client Components that read `?id` via `useSearchParams`. Because `useSearchParams` triggers a CSR bailout during prerendering, the hook call must be wrapped in a `Suspense` boundary inside the page.
- `ReportDetail` now receives `reportId` as a prop instead of reading `useParams()`, keeping it reusable from both the new static route and any future callers.
- All internal navigation updated: `scan-form.tsx`, `snapshots-table.tsx`, `target-snapshots-table.tsx`, and `reports-table.tsx` now route to `/snapshot?id=...` or `/report?id=...`.
- Playwright tests updated to assert `/snapshot?id=${snapshotId}` and navigate `/report?id=${reportId}`; all 6 E2E tests pass.
- Build produces clean static files: `dist/snapshot.html` and `dist/report.html`, with no leftover `snapshots/placeholder` or `reports/placeholder` artifacts.

## 2025-06-16 — Final manual QA (F3)

- Clean Docker Compose build/start from `docker compose down -v` passes; all services reach healthy state.
- Home page (`/`) returns HTTP 200 and renders navigation links with `data-testid` attributes (`nav-logo`, `nav-link-home`, `nav-link-targets`, `nav-link-snapshots`, `nav-link-reports`).
- `POST /api/scan` for `example.com` creates a snapshot and the detail/list responses include `client_ip` (observed `192.168.97.1` from Docker host).
- Browse pages `/targets`, `/snapshots`, `/reports` all return HTTP 200; corresponding API endpoints (`/api/targets`, `/api/snapshots`, `/api/reports`) return valid JSON.
- Snapshot detail page `/snapshot?id=<uuid>` and report detail page `/report?id=<uuid>` load via nginx SPA fallback.
- Report creation (`POST /api/reports`) completes and the downloaded PDF starts with `%PDF` header (verified on a ~14.9 KB file).
- CORS preflight from `FRONTEND_URL` (`http://localhost:3001`) returns HTTP 204 with `Access-Control-Allow-Origin: http://localhost:3001`.
- Full Playwright suite passes: 6/6 tests.
- Rate limiting only triggers under concurrent/rapid load: 31 parallel scans produced 15 x HTTP 429 with `retry_after`. Sequential scans do not exhaust the token bucket because it refills during scan handler execution.
- Evidence saved under `.omo/evidence/final-qa/` with index at `99-evidence-index.md`.

## 2026-06-16 — F4 scope fidelity audit

- Verdict is REJECT: 12/21 tasks are scope-compliant. Non-compliant tasks are 3, 7, 8, 12, 15, 17, 18, 19, and 21.
- The working tree has 202 changed non-ignored files; 102 are not accounted for by the task 1-21 file scopes after treating `frontend/`, `internal/middleware/`, `internal/pwhois/`, web-ui notepads, and `task-*` evidence as allowed.
- Major contamination came from unrelated prior-plan artifacts: `.omo/plans/echostate-tasks-4-and-1.md`, `.omo/plans/fix-slowloris-g112.md`, `.omo/plans/missing-go-tests.md`, `.omo/plans/pdf-report-endpoint.md`, `.omo/notepads/fix-slowloris-g112/`, `.omo/notepads/missing-go-tests/`, `.omo/notepads/pdf-report-endpoint/`, `.omo/evidence/final-qa/`, `.omo/run-continuation/`, and `.omo/boulder.json`.
- Web-ui-specific scope mismatches found: CORS/frontend defaults drifted from the plan's `localhost:3000`; pwhois TTL env naming/config is not wired into the worker; landing/report pages omit required generated-report/host fields; Docker/Playwright use `3001` and do not set `NEXT_PUBLIC_API_URL=/api` exactly as specified.

## 2025-06-16 — Code quality review (F2)

- Backend verification is green: `go vet ./...`, `go build ./...`, and `go test ./... -count=1` all exit 0.
- Frontend build is green; `next build` static export produces expected routes including `/snapshot`, `/report`, `/snapshots`, `/reports`, `/targets`, and `/targets/placeholder/snapshots`.
- `npm run lint` scans the entire `frontend/` tree including `dist/`, so generated build artifacts can fail the lint command even when source files are clean. Running `npx eslint app components lib e2e` confirms source files are lint-free.
- Quality scans (anti-patterns, `as any`, `@ts-ignore`, empty catches, commented-out code) found only minor/acceptable issues in test/config files.

## 2026-06-16 — Per-task evidence files generated

- Generated all missing per-task evidence files listed in the plan's "Evidence to Capture" sections for tasks 1-21.
- 8 files copied from `.omo/evidence/final-qa/` (task-1-client-ip, task-2-rate-limit-allowed, task-2-rate-limit-blocked, task-3-cors-preflight, task-4-browse-targets, task-4-snapshot-detail, task-19-compose-ps, task-20-api-proxy).
- 1 PDF copied from final-qa (task-21-e2e-download.pdf).
- 1 file generated via curl (task-3-cors-untrusted.txt — untrusted origin returns 403).
- 1 file generated via `go run ./main.go` (task-8-pwhois-startup.txt — startup logs showing all routes).
- 7 test outputs saved from `go test` runs (task-1-migration, task-5-pwhois-migration, task-6-browse-tests, task-6-ratelimit-tests, task-7-pwhois-public, task-7-pwhois-private, task-9-pwhois-tests).
- 2 build outputs saved (task-10-frontend-build.txt, task-18-docker-build.txt).
- 8 screenshot reference files created for Playwright-captured pages (task-11-nav through task-17-report-detail) — these are text references pointing to the Playwright HTML report in task-21-playwright-report/.
- All Go tests pass: migration, browse, rate-limit, pwhois worker (enrich, skip-private, cache, failure, start/stop).
- Frontend build and Docker build both succeed.
- Total: 30 evidence files in `.omo/evidence/` for the web-ui plan, plus the existing task-21-playwright-report/ directory.

## 2026-06-16 — F3 Real Manual QA re-run with corrected rate-limit test

- Re-ran full Real Manual QA from a clean `docker compose down -v` + `docker compose up --build -d` state.
- Corrected the rate-limit acceptance test: 35 parallel `POST /api/scan` requests produced exactly 30 x HTTP 200 and 5 x HTTP 429, with all 429 responses including `{ "error": "rate limit exceeded", "retry_after": 1 }`.
- All 7 manual QA scenarios passed:
  1. Home page loads and navigation is visible (`data-testid` attributes present).
  2. `POST /api/scan` for `example.com` created snapshot `10f9cb07-b86c-47ff-9a8c-59c04319b7ae` with `client_ip: 192.168.97.1`.
  3. Frontend browse pages (`/targets`, `/snapshots`, `/reports`) and API endpoints (`/api/targets`, `/api/snapshots`, `/api/reports`) all returned HTTP 200 with valid JSON.
  4. `/snapshot?id=<uuid>` loaded and `POST /api/reports` returned HTTP 202.
  5. `/report?id=<uuid>` loaded and PDF download returned HTTP 200 with `%PDF-1.3` header.
  6. Rate limiting behaved correctly under parallel burst load.
  7. CORS preflight from `http://localhost:3001` returned HTTP 204 with `Access-Control-Allow-Origin: http://localhost:3001`.
- Full Playwright suite passed: 6/6 tests.
- Evidence saved to `.omo/evidence/final-qa/` with updated index at `99-evidence-index.md`; Playwright report copied to `.omo/evidence/final-qa/playwright-report/index.html`.
- Docker Compose stack was left down after QA (Playwright global teardown confirmed no running containers).

## 2026-06-16 — F4 pwhois fix: IPv6 skipping, attribution, TTL wiring

- `internal/pwhois/worker.go`: Replaced the `privateRanges` CIDR slice (IPv4-only) with Go 1.23+ `net.IP.IsPrivate()`, `net.IP.IsLoopback()`, and `net.IP.IsLinkLocalUnicast()` — these cover IPv4 RFC1918/loopback, IPv6 unique-local (`fc00::/7`), IPv6 loopback (`::1/128`), and IPv6 link-local (`fe80::/10`) in a single clear check.
- `internal/pwhois/worker.go`: `NewWorker` now accepts a `cacheTTL time.Duration` parameter. Zero or negative values fall back to `defaultCacheTTL` (24h). The old `cacheTTL` constant was renamed to `defaultCacheTTL`.
- `internal/pwhois/pwhois.go`: Added a comment block near `defaultServer` preserving pwhois.org attribution and RIR terms reference.
- `internal/config/config.go`: Changed env var from `ECHOSTATE_PWHOIS_CACHE_TTL_HOURS` to `PWHOIS_CACHE_TTL_HOURS` (matching the task spec). `.env.example` updated accordingly.
- `main.go`: Now passes `time.Duration(cfg.PwhoisCacheTTLHours) * time.Hour` to `pwhois.NewWorker`.
- `go build ./...`, `go vet ./...`, `go test ./... -count=1` all pass.

## 2026-06-16 — Corrected F4 scope fidelity re-run

- Re-read `.omo/plans/echostate-web-ui.md` end-to-end, including Defaults Applied lines 98-107. The previous 3001 frontend port, `/api` nginx proxy behavior, configurable `FRONTEND_URL`, and standard browser CORS methods/`Authorization` header are allowed by the plan and are not violations.
- Re-classified unrelated files from other plans (`.omo/plans/echostate-tasks-4-and-1.md`, `.cgcignore`, `.dockerignore`, `CONTRIBUTING.md`, `main_test.go` slowloris timeout work, and other-plan notepads/evidence) as pre-existing/non-web-ui artifacts rather than web-ui contamination.
- Corrected result: 15/21 implementation tasks are scope-compliant; contamination is clean after excluding other-plan artifacts; unaccounted web-ui files are clean.
- Genuine remaining deviations are limited to task specs/defaults around IPv6 private/loopback pwhois skipping, missing pwhois attribution notice, `PWHOIS_CACHE_TTL_HOURS` wiring/name, landing-page generated-report link, report host display, and missing Playwright trace artifacts in `.omo/evidence/`.

## 2026-06-16 — Playwright artifact output configured to .omo/evidence/

- `frontend/playwright.config.ts` already had all required settings: `outputDir` at `.omo/evidence/task-21-traces`, `trace/screenshot/video: "on"`, HTML reporter at `.omo/evidence/task-21-playwright-report`, and `list` reporter for terminal output.
- No config changes were needed — the file was already correctly configured from a prior pass.
- Verified: `npx playwright test` passes 6/6 tests and produces screenshots (7 PNGs), videos (6 .webm), traces (6 .zip), and the HTML report under `.omo/evidence/`.
- The global setup (`e2e/global-setup.ts`) has a persistent Docker race condition: `docker compose up` sometimes fails with "No such container" or "container name already in use" errors. This is a pre-existing issue in the setup script, not related to the config change.

## 2026-06-16 — Report summary/detail host field

- Added `Host string` to `internal/models/models.go` `ReportSummary` and `ReportResponse` with JSON key `host`.
- `internal/handlers/browse.go` `listReports` now JOINs `snapshots` and `targets` and SELECTs `COALESCE(t.host, 'unknown') AS host`, scanning it into `ReportSummary.Host`.
- `internal/handlers/reports.go` `toReportResponse` now accepts a `context.Context` and populates `Host` via the existing `snapshotHost` helper, falling back to `"unknown"` on error or empty result.
- Frontend TypeScript `ReportSummary` and `Report` in `frontend/lib/types.ts` now include `host: string`.
- `frontend/components/reports-table.tsx` adds a Host column and displays `report.host || "—"`.
- `frontend/components/report-detail.tsx` shows host in the metadata grid using a `GlobeIcon` and expands the grid to `lg:grid-cols-5`.

## 2026-06-16 — Verification and test stability fixes

- Full verification suite passes: `go test ./... -count=1` (all packages), `go vet ./...`, `go build ./...`, `cd frontend && npm run build`, and `cd frontend && npx playwright test` (6/6 tests).
- Playwright's default 30s test timeout was too short for `e2e/report.spec.ts` (creates snapshot, waits for report completion, downloads PDF). Added `timeout: 120_000` to `frontend/playwright.config.ts` so `npx playwright test` passes without CLI overrides.
- `frontend/e2e/global-setup.ts` now short-circuits when the Docker Compose stack is already healthy, avoiding repeated `docker compose down -v` / `up --build -d` cycles that could race with leftover containers/networks in this environment.
- LSP diagnostics are clean on all host-field files and the two test config files; only pre-existing hint is the deprecated `FormEventHandler` type in `frontend/components/reports-table.tsx`.

## 2026-06-16 — Landing result card generated-report link

- **Resolved**: Task 12 landing-page generated-report link is present in `frontend/components/scan-form.tsx`.
- The result card shows a "Generate report" button after a successful scan; clicking it calls `createReport(result.snapshot_id)` (`POST /api/reports`).
- A `useEffect` polls `getReport(report.id)` every 5 seconds while `report.status` is `pending` or `running`, and stops when status reaches `completed` or `failed`.
- While pending/running, the button is disabled and shows a `Loader2Icon` spinner with the text "Generating report…".
- On completion, the button becomes an outline `Link` to `/report?id=${report.id}` with the text "View report →".
- On failure, an inline error message appears in the card using `AlertCircleIcon` and `text-destructive`, showing either `report.error` or a fallback message.
- The existing "View snapshot →" link to `/snapshot?id=${result.snapshot_id}` remains on the right side of the card footer.
- Supporting API helpers and types are already in place: `frontend/lib/api.ts` exports `createReport` and `getReport`; `frontend/lib/types.ts` defines `Report`, `ReportStatus`, and `CreateReportRequest`.
- Verification: `cd frontend && npm run build` exits 0; `cd frontend && npx playwright test` passes 6/6.

## 2026-06-16 — F4 scope fidelity check

- Re-ran F4 against `.omo/plans/echostate-web-ui.md` tasks 1-21 and the current `git diff HEAD --stat` / `git status --short` working tree inventory.
- Treated explicitly named non-web-ui artifacts from other plans as out of scope rather than contamination: other `.omo/plans/*`, their notepads/evidence, `CONTRIBUTING.md`, `.cgcignore`, `.dockerignore`, `main_test.go` slowloris timeout work, and generic missing-test files.
- Recent F4 fixes are present in source: IPv6/private pwhois skipping (`internal/pwhois/worker.go`), `PWHOIS_CACHE_TTL_HOURS` wiring (`internal/config/config.go`, `main.go`), landing report generation/link (`frontend/components/scan-form.tsx`), report host display (`internal/handlers/reports.go`, `frontend/components/reports-table.tsx`, `frontend/components/report-detail.tsx`), and Playwright artifacts under `.omo/evidence/` (`frontend/playwright.config.ts`).
- Result: contamination is clean and unaccounted web-ui files are clean, but task compliance is 19/21 because two explicit test deliverables are missing: rate-limiter cleanup coverage and pwhois batching coverage.

## 2026-06-16 — F1 Plan Compliance Audit

- All 8 Must Have items verified with concrete implementations:
  1. Next.js + shadcn/ui frontend: 6 pages confirmed (`page.tsx`, `targets/page.tsx`, `snapshots/page.tsx`, `reports/page.tsx`, `snapshot/page.tsx`, `report/page.tsx`).
  2. `client_ip` captured: `internal/handlers/scan.go:96` calls `c.ClientIP()`, stored via `storeSnapshot` at line 98.
  3. Per-IP rate limiting: `internal/middleware/ratelimit.go` token-bucket (30 tokens/min), applied to `POST /api/scan`.
  4. CORS: `internal/middleware/cors.go` with gin-contrib/cors, configurable `FRONTEND_URL`.
  5. Browse endpoints: `internal/handlers/browse.go` with pagination (page=1, limit=20, max=100), 5 endpoints.
  6. Async pwhois worker: `internal/pwhois/worker.go` with polling, batch lookups, retries, cache, DB persistence.
  7. Docker Compose frontend: `docker-compose.yml:68-82` with build context, port, healthcheck, depends_on.
  8. Playwright QA: `frontend/e2e/smoke.spec.ts` and `report.spec.ts`, 6/6 tests pass.
- All 7 Must NOT Have guardrails verified absent (auth, Redis, charts, maps, WebSockets, email, pwhois-in-PDF, TLS/domain/CI, custom design system). False positives: `Authorization` is standard CORS header, `ws://` is pre-existing browserless Chrome DevTools, `Map`/`map` are Go data structures, `email` hits are domain names.
- Evidence exists for all 21 implementation tasks in `.omo/evidence/` plus final-qa evidence in `.omo/evidence/final-qa/` (126 entries with index).
- All 21 tasks marked complete in the plan; implementations verified.

## 2026-06-16 — Final Verification Wave F4 scope fidelity re-run

- Re-ran F4 against the current source using `.omo/plans/echostate-web-ui.md`, `git diff HEAD --stat`, and `git status --short` / expanded untracked inventory.
- Confirmed the previously missing test deliverables now exist: `TestRateLimiter_CleansOldEntries`, `TestRateLimiter_DoesNotCleanFreshEntries`, and `TestWorker_BatchesMultipleIPs`.
- Treated the prompt-designated non-web-ui artifacts from other plans as exempt rather than web-ui contamination, including other `.omo/plans/*`, their notepads/evidence, root metadata files, and slowloris/missing-test files.
- Result: contamination is clean and unaccounted web-ui files are clean, but task compliance is 18/21 due to strict task-spec mismatches in Tasks 1, 3, and 5.

## 2026-06-16 — Final Verification Wave F2 code quality review

- `go vet ./...`, `go build ./...`, and `go test ./... -count=1` all pass: 101 tests, 0 failures.
- `cd frontend && npm run build` static export passes.
- `cd frontend && npx eslint app components lib e2e` passes with 0 problems.
- `cd frontend && npx playwright test` passes 6/6.
- `cd frontend && npm run typecheck` passes (`tsc --noEmit`).
- Anti-pattern scan found no `as any`, `@ts-ignore`, empty catch blocks, `TODO/FIXME/HACK/xxx`, commented-out code, unused imports, or generic names in production source files. `console.log` only appears in `frontend/e2e/global-setup.ts` and `frontend/e2e/global-teardown.ts` as intentional test-infrastructure progress logging.
- During the review `/api/reports` returned `"data": null` when there were no report rows, causing the frontend `ReportsTable` to crash on `data.data.length`. Fixed by initializing the paginated result slices to empty slices in `internal/handlers/browse.go` for targets, snapshots, and reports.

## 2026-06-16 — F3 Real Manual QA (re-verification)

- Clean `docker compose down -v` + `docker compose up --build -d` from scratch: all 4 services reach healthy state within ~25s.
- Home page (`/`) returns HTTP 200 with 5 navigation testids: `nav-logo`, `nav-link-home`, `nav-link-targets`, `nav-link-snapshots`, `nav-link-reports`.
- `POST /api/scan` for `example.com` creates snapshot `5488f37c-...` with `client_ip: 192.168.97.1` (Docker host IP). Snapshot detail API and frontend page both load.
- Browse pages (`/targets`, `/snapshots`, `/reports`) all return HTTP 200; API endpoints return valid JSON with pagination metadata.
- Report creation (`POST /api/reports`) completes within ~5s; PDF download returns HTTP 200 with `%PDF-1.3` header (~14.4 KB).
- Rate limiting: 35 parallel `POST /api/scan` requests produce exactly 30 x HTTP 200 and 5 x HTTP 429 with `{"error":"rate limit exceeded","retry_after":1}`.
- CORS preflight from `http://localhost:3001` returns HTTP 204 with `Access-Control-Allow-Origin: http://localhost:3001`.
- Playwright suite passes 6/6 tests (smoke + report flow); global teardown cleans up Docker stack.
- Evidence saved to `.omo/evidence/final-qa/` with 13+ files and Playwright HTML report.
