# EchoState: PDF Report Generation Endpoint

## TL;DR

> **Quick Summary**: Add async PDF report generation endpoints to EchoState using maroto v2, storing generated PDFs in PostgreSQL and supporting both snapshot-ID and latest-per-host report sources, plus both direct download and stored download-URL delivery.
>
> **Deliverables**:
> - `internal/pdf/` package with maroto v2 report renderer
> - `internal/handlers/reports.go` with report endpoints
> - `internal/models/models.go` updates for report request/response structs
> - `internal/db/db.go` migration for `reports` table and `targets.normalized_host`
> - `internal/handlers/scan.go` updated to populate `targets.normalized_host`
> - Unit and integration tests for PDF generation and report endpoints
> - Docker Compose end-to-end verification
>
> **Estimated Effort**: Medium
> **Parallel Execution**: YES — 3 waves + final verification
> **Critical Path**: 1 → 2 → 3 → 4 → 5 → 6 → 7 → 8 → F1-F4

---

## Context

### Original Request
Add a PDF report generation endpoint to EchoState.

### Interview Summary
**Key Discussions**:
- Support two report sources: by specific `snapshot_id`, or latest snapshot for a `host`.
- Support two delivery modes: direct PDF download endpoint, and stored PDF with download URL for later viewing.
- Report should include all relevant reconnaissance data: cover page, metadata, WHOIS, ASN/BGP, web/copyright, errors, footer.
- Generation should be async with a polling status endpoint.
- Use PostgreSQL BYTEA storage for generated PDFs.

**Research Findings**:
- EchoState uses Gin with handlers in `internal/handlers/scan.go`, raw SQL via `pgxpool`, models in `internal/models/models.go`, migrations inline in `internal/db/db.go`.
- Snapshots store the full `ScanResult` as JSONB under `raw_data`; `ScanResult` has `Host`, `ScannedAt`, `WHOIS`, `ASN`, `Web`, `Errors`.
- No existing PDF or report code; maroto v2 is the recommended library.

### Metis Review
**Identified Gaps** (addressed):
- Resolve `host` to concrete `snapshot_id` at job creation time.
- Persist report status transitions in PostgreSQL.
- Add report generation timeout and concurrency limit.
- Add max PDF size guard.
- Add `normalized_host` column to `targets` so report host lookup works regardless of how the host was originally submitted.

### Momus Review
**Mode**: High Accuracy Review (user-selected)
**Verdict**: OKAY
**Notes**: All referenced files exist on disk and contain the claimed content. Every task has sufficient context for execution. QA scenarios are executable with concrete tools, steps, and expected results. No blocking issues found.

---

## Work Objectives

### Core Objective
Add async PDF report generation endpoints to EchoState that render stored reconnaissance snapshots as downloadable PDF reports.

### Concrete Deliverables
- `internal/pdf/` package with maroto v2 renderer.
- `internal/handlers/reports.go` with report endpoints.
- `internal/models/models.go` updates for report request/response structs.
- `internal/db/db.go` migration for `reports` table and `targets.normalized_host`.
- `internal/handlers/scan.go` updated to populate `targets.normalized_host`.
- Unit and integration tests.
- Docker Compose end-to-end verification.

### Definition of Done
- [ ] `go test ./...` passes.
- [ ] `go vet ./...` passes.
- [ ] Docker Compose stack builds and runs.
- [ ] `POST /api/reports` creates a report job and returns 202.
- [ ] `GET /api/reports/:id` returns status `pending|running|completed|failed`.
- [ ] `GET /api/reports/:id/download` returns valid PDF bytes for completed reports.
- [ ] `GET /api/snapshots/:id/report` returns or creates a report for a snapshot.

### Must Have
- `POST /api/reports` accepts `{ "snapshot_id": "uuid" }` or `{ "host": "example.com" }` and returns 202 with report metadata.
- Report jobs persisted in PostgreSQL with status transitions.
- PDF generated with maroto v2 including cover, metadata, WHOIS, ASN, web/copyright, errors, footer.
- Direct download endpoint returns `Content-Type: application/pdf` with attachment disposition.
- Convenience endpoint `GET /api/snapshots/:id/report` creates/reuses report and returns it.
- Bounded in-process worker concurrency and per-job timeout.
- `targets.normalized_host` column populated on scan and used for host-based report lookup.
- Tests cover happy path, missing snapshot, missing host, download before completion, and invalid input.

### Must NOT Have (Guardrails)
- No frontend UI, email delivery, scheduling, batch reporting, or report customization/templates.
- No re-running reconnaissance during report generation.
- No external object storage or background worker framework beyond in-process goroutines.
- No auth/authorization changes.
- No synchronous long-running PDF generation in request handlers (except the convenience endpoint reuses stored async result).
- No HTML-to-PDF/browser rendering stack; use maroto v2 only.

---

## Verification Strategy (MANDATORY)

> **ZERO HUMAN INTERVENTION** - ALL verification is agent-executed. No exceptions.

### Test Decision
- **Infrastructure exists**: YES (Go with `go test`, existing `*_test.go` files in scanner package).
- **Automated tests**: YES (tests-after implementation).
- **Framework**: `go test`.
- **Agent-Executed QA**: ALWAYS — every task includes concrete `curl`/Docker-based QA scenarios with expected status codes, JSON fields, and PDF byte assertions.

### QA Policy
Every task MUST include agent-executed QA scenarios. Evidence saved to `.omo/evidence/task-{N}-{scenario-slug}.{ext}`.

- **Library/Module**: Use `Bash` (`go test`, `go vet`, `go build`).
- **API/Backend**: Use `Bash` (`curl`) against the running container.
- **Docker**: Use `Bash` (`docker compose`, `docker exec`).

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 1 (Foundation - start immediately):
├── 1. Add maroto v2 dependency [quick]
├── 2. Add reports table migration + normalized_host [quick]
└── 3. Update models with report structs [quick]

Wave 2 (Core implementation - after Wave 1):
├── 4. Create internal/pdf renderer [unspecified-high]
├── 5. Create report worker/job manager [unspecified-high]
└── 6. Populate normalized_host in scan handler [quick]

Wave 3 (HTTP layer + integration - after Wave 2):
├── 7. Add reports HTTP handlers and routes [unspecified-high]
├── 8. Add unit and integration tests [unspecified-high]
└── 9. Docker Compose end-to-end verification [unspecified-high]

Wave FINAL (After ALL tasks):
├── F1. Plan compliance audit (oracle)
├── F2. Code quality review (unspecified-high)
├── F3. Real manual QA (unspecified-high)
└── F4. Scope fidelity check (deep)
-> Present results -> Get explicit user okay
```

### Dependency Matrix

- **1**: - → 4
- **2**: - → 4, 5, 7
- **3**: - → 4, 5, 7
- **4**: 1, 2, 3 → 5, 7
- **5**: 2, 3, 4 → 7
- **6**: 2 → 7
- **7**: 2, 3, 4, 5, 6 → 8, 9
- **8**: 7 → 9
- **9**: 7, 8 → F1-F4
- **F1-F4**: 1-9 → user okay

### Agent Dispatch Summary

- **Wave 1**: 1 → `quick`, 2 → `quick`, 3 → `quick`
- **Wave 2**: 4 → `unspecified-high`, 5 → `unspecified-high`, 6 → `quick`
- **Wave 3**: 7 → `unspecified-high`, 8 → `unspecified-high`, 9 → `unspecified-high`
- **FINAL**: F1 → `oracle`, F2 → `unspecified-high`, F3 → `unspecified-high`, F4 → `deep`

---

## TODOs

- [x] 1. Add maroto v2 dependency

  **What to do**:
  - Add `github.com/johnfercher/maroto/v2` to the project.
  - Run `go get github.com/johnfercher/maroto/v2`.
  - Run `go mod tidy`.
  - Verify `go build ./...` still succeeds.

  **Must NOT do**:
  - Do not add unused dependencies.
  - Do not upgrade unrelated dependencies.

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: none required

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 2, 3)
  - **Blocks**: 4
  - **Blocked By**: None

  **References**:
  - `go.mod` — current module dependencies.
  - `https://pkg.go.dev/github.com/johnfercher/maroto/v2` — maroto v2 API.

  **Acceptance Criteria**:
  - [ ] `go.mod` lists `github.com/johnfercher/maroto/v2`.
  - [ ] `go mod tidy` exits cleanly.
  - [ ] `go build ./...` succeeds.

  **QA Scenarios**:

  ```
  Scenario: maroto v2 dependency added and project builds
    Tool: Bash
    Preconditions: Working directory is /Users/house/Documents/gitlab/echostate
    Steps:
      1. Run `grep 'johnfercher/maroto/v2' go.mod`
      2. Run `go mod tidy`
      3. Run `go build ./...`
    Expected Result: grep shows maroto dependency; build exits 0.
    Failure Indicators: Missing dependency; build fails.
    Evidence: .omo/evidence/task-1-maroto-dep.txt
  ```

  **Evidence to Capture**:
  - [ ] Terminal output showing dependency and successful build.

  **Commit**: YES
  - Message: `chore(deps): add maroto v2 for PDF generation — Jinkies!`
  - Files: `go.mod`, `go.sum`

- [x] 2. Add reports table migration and normalized_host column

  **What to do**:
  - Update `internal/db/db.go` `Migrate()` function:
    - Add `normalized_host TEXT` column to `targets` table.
    - Add index on `targets(normalized_host)`.
    - Create `reports` table:
      ```sql
      CREATE TABLE IF NOT EXISTS reports (
          id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
          snapshot_id UUID NOT NULL REFERENCES snapshots(id) ON DELETE CASCADE,
          status TEXT NOT NULL DEFAULT 'pending',
          error_message TEXT,
          pdf BYTEA,
          created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
          updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
          completed_at TIMESTAMPTZ
      );
      ```
    - Add index on `reports(snapshot_id, status)`.
  - For existing deployments, add `ALTER TABLE` statements to add `normalized_host` if not present (idempotent).

  **Must NOT do**:
  - Do not create separate `.sql` migration files (keep inline pattern).
  - Do not drop existing tables.

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: none required

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 3)
  - **Blocks**: 4, 5, 7
  - **Blocked By**: None

  **References**:
  - `internal/db/db.go` — existing `Migrate()` function and table definitions.
  - `internal/scanner/normalize.go` — `NormalizeHost` for normalized_host values.

  **Acceptance Criteria**:
  - [ ] `go build ./...` succeeds.
  - [ ] `go vet ./...` passes.
  - [ ] Migration SQL is idempotent (uses `IF NOT EXISTS` / `ALTER TABLE IF NOT EXISTS`).

  **QA Scenarios**:

  ```
  Scenario: Database migration creates reports table and normalized_host
    Tool: Bash
    Preconditions: Docker Compose stack running with fresh DB
    Steps:
      1. Run `docker compose up --build -d`
      2. Run `docker exec echostate-db psql -U echostate -d echostate -c "\dt"`
      3. Run `docker exec echostate-db psql -U echostate -d echostate -c "SELECT column_name FROM information_schema.columns WHERE table_name='targets' AND column_name='normalized_host';"`
      4. Run `docker exec echostate-db psql -U echostate -d echostate -c "SELECT column_name FROM information_schema.columns WHERE table_name='reports';"`
    Expected Result: `reports` table exists; `targets` has `normalized_host` column.
    Failure Indicators: Table or column missing; migration errors.
    Evidence: .omo/evidence/task-2-db-migration.txt
  ```

  **Evidence to Capture**:
  - [ ] `\dt` output.
  - [ ] Column verification output.

  **Commit**: YES
  - Message: `chore(db): add reports table and normalized_host column — Ruh-roh!`
  - Files: `internal/db/db.go`

- [x] 3. Update models with report structs

  **What to do**:
  - Add to `internal/models/models.go`:
    - `CreateReportRequest` struct with optional `SnapshotID *uuid.UUID` and `Host *string`.
    - `ReportResponse` struct with `ID`, `Status`, `SnapshotID`, `DownloadURL`, `Error`, `CreatedAt`, `CompletedAt`.
    - `ReportStatus` constants: `ReportPending`, `ReportRunning`, `ReportCompleted`, `ReportFailed`.
    - `Report` internal model matching the DB row.
  - Use `uuid.UUID` from existing `github.com/google/uuid` dependency.

  **Must NOT do**:
  - Do not create separate request/response packages; keep all models in `internal/models/models.go`.

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: none required

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 1 (with Tasks 1, 2)
  - **Blocks**: 4, 5, 7
  - **Blocked By**: None

  **References**:
  - `internal/models/models.go` — existing model patterns.
  - `internal/scanner/scanner.go` — uses `models.ScanResult`.

  **Acceptance Criteria**:
  - [ ] New structs compile.
  - [ ] `go vet ./...` passes.
  - [ ] No JSON tag conflicts with existing fields.

  **QA Scenarios**:

  ```
  Scenario: New report models compile
    Tool: Bash
    Preconditions: Task 1 dependencies present
    Steps:
      1. Run `go build ./...`
      2. Run `go vet ./...`
    Expected Result: Build and vet exit 0.
    Failure Indicators: Compile errors; vet warnings.
    Evidence: .omo/evidence/task-3-models-compile.txt
  ```

  **Evidence to Capture**:
  - [ ] Build/vet output.

  **Commit**: YES
  - Message: `chore(models): add report request and response structs — Zoinks!`
  - Files: `internal/models/models.go`

- [x] 4. Create `internal/pdf` report renderer

  **What to do**:
  - Create `internal/pdf/renderer.go`.
  - Implement `RenderReport(result *models.ScanResult) ([]byte, error)` using maroto v2.
  - Report layout must include:
    - Cover page: title "EchoState Reconnaissance Report", host, scan timestamp.
    - Metadata section: host, scanned at, report generated at.
    - WHOIS section: registrar, expiration date, name servers, status, registrar.
    - ASN/BGP section: ASN, prefix, country, registry, allocated, AS name, IP.
    - Web section: title, final URL, copyright snippets.
    - Errors section: list of gatherer errors (if any).
    - Footer with page numbers.
  - Handle missing/empty sections gracefully (skip or show "N/A").
  - Truncate long fields (e.g., raw WHOIS, copyright list) to avoid oversized PDFs.
  - Return PDF bytes.

  **Must NOT do**:
  - Do not use chromedp or HTML-to-PDF.
  - Do not embed images, charts, or logos.
  - Do not expose internal error details in rendered output.

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: none required

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 2 (with Tasks 5, 6)
  - **Blocks**: 5, 7
  - **Blocked By**: 1, 2, 3

  **References**:
  - `internal/models/models.go` — `ScanResult` fields.
  - `internal/scanner/whois.go`, `asn.go`, `web.go` — expected map keys.
  - `https://pkg.go.dev/github.com/johnfercher/maroto/v2` — maroto v2 API.

  **Acceptance Criteria**:
  - [ ] `internal/pdf/renderer.go` compiles.
  - [ ] `RenderReport` returns non-empty bytes starting with `%PDF` for a sample `ScanResult`.
  - [ ] Unit test passes for renderer with sample data.

  **QA Scenarios**:

  ```
  Scenario: Renderer produces a valid PDF for sample scan data
    Tool: Bash
    Preconditions: maroto v2 dependency added
    Steps:
      1. Run `go test ./internal/pdf/ -run TestRenderReport -v`
    Expected Result: Test passes; output bytes start with `%PDF`; no panic.
    Failure Indicators: Test fails; bytes do not start with `%PDF`; panic.
    Evidence: .omo/evidence/task-4-pdf-renderer.txt
  ```

  **Evidence to Capture**:
  - [ ] Test output.
  - [ ] Optional: first 16 bytes of generated PDF.

  **Commit**: YES
  - Message: `feat(pdf): add maroto report renderer — Jeepers!`
  - Files: `internal/pdf/renderer.go`, `internal/pdf/renderer_test.go`

- [x] 5. Create async report worker/job manager

  **What to do**:
  - Create `internal/reports/worker.go` (or add to `internal/pdf/` if preferred; keep separate for clarity).
  - Implement a `Worker` struct that:
    - Accepts `*db.DB` and a renderer function.
    - Has a `Start(ctx context.Context)` method that processes pending reports from DB.
    - Uses a semaphore/channel to limit concurrent jobs (e.g., max 3 concurrent).
    - Loads the snapshot row, unmarshals `raw_data` into `models.ScanResult`, calls renderer, writes PDF bytes back, updates status to `completed`.
    - On error: updates status to `failed` and stores `error_message`.
    - Uses a per-job timeout (e.g., 30 seconds) via `context.WithTimeout`.
  - Implement `CreateReport(ctx, snapshotID) (reportID, error)` that inserts a pending report row and triggers processing.
  - Implement `GetReport(ctx, reportID)` to fetch report status and PDF.
  - On startup, call `worker.Start(ctx)` in `main.go` after DB setup.
  - On restart, mark any `running` reports as `failed` with a timeout message.

  **Must NOT do**:
  - Do not use an external queue framework.
  - Do not block HTTP handlers during PDF generation (except convenience endpoint may poll/wait briefly).
  - Do not store more than 10 MB of PDF bytes per report.

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: none required

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 2 (with Tasks 4, 6)
  - **Blocks**: 7
  - **Blocked By**: 2, 3, 4

  **References**:
  - `internal/db/db.go` — `DB` struct and migration.
  - `internal/models/models.go` — `Report`, `Snapshot`, `ScanResult` structs.
  - `internal/pdf/renderer.go` — `RenderReport`.
  - `internal/handlers/scan.go` — raw SQL patterns.

  **Acceptance Criteria**:
  - [ ] `CreateReport` inserts a pending row.
  - [ ] `Start` processes pending rows and updates status to `completed`/`failed`.
  - [ ] Concurrent jobs are bounded.
  - [ ] `go test ./internal/reports/...` passes.

  **QA Scenarios**:

  ```
  Scenario: Worker completes a report job
    Tool: Bash
    Preconditions: Test DB seeded with a snapshot
    Steps:
      1. Run `go test ./internal/reports/ -run TestWorkerCompletes -v`
    Expected Result: Test passes; report status transitions pending -> running -> completed; PDF bytes non-empty.
    Failure Indicators: Status stuck pending; PDF empty; panic.
    Evidence: .omo/evidence/task-5-worker-completes.txt

  Scenario: Worker fails a report job for missing snapshot
    Tool: Bash
    Preconditions: None
    Steps:
      1. Run `go test ./internal/reports/ -run TestWorkerFailsMissingSnapshot -v`
    Expected Result: Test passes; report status becomes failed with error message.
    Failure Indicators: Status remains pending/running; no error message.
    Evidence: .omo/evidence/task-5-worker-fails.txt
  ```

  **Evidence to Capture**:
  - [ ] Test output for both scenarios.

  **Commit**: YES
  - Message: `feat(reports): add async report worker — Scooby-Dooby-Doo!`
  - Files: `internal/reports/worker.go`, `internal/reports/worker_test.go`

- [x] 6. Populate `normalized_host` in scan handler

  **What to do**:
  - Update `internal/handlers/scan.go` `upsertTarget` function:
    - Normalize `host` with `scanner.NormalizeHost(host)`.
    - Update INSERT/ON CONFLICT to also set `normalized_host`.
    - Update the RETURNING/SELECT to include normalized_host if needed.
  - Add idempotent migration in `internal/db/db.go` to backfill `normalized_host` for existing targets (optional; if time permits).

  **Must NOT do**:
  - Do not change the unique constraint on raw `host`.
  - Do not break existing scan endpoint behavior.

  **Recommended Agent Profile**:
  - **Category**: `quick`
  - **Skills**: none required

  **Parallelization**:
  - **Can Run In Parallel**: YES
  - **Parallel Group**: Wave 2 (with Tasks 4, 5)
  - **Blocks**: 7
  - **Blocked By**: 2

  **References**:
  - `internal/handlers/scan.go` — `upsertTarget` function.
  - `internal/scanner/normalize.go` — `NormalizeHost`.

  **Acceptance Criteria**:
  - [ ] Scanning a host stores `normalized_host` in `targets`.
  - [ ] `go test ./...` passes.
  - [ ] Existing scan tests still pass.

  **QA Scenarios**:

  ```
  Scenario: Scan stores normalized_host for URL input
    Tool: Bash
    Preconditions: Docker Compose stack running
    Steps:
      1. Run `curl -s -X POST http://localhost:8080/api/scan -H "Content-Type: application/json" -d '{"host":"https://www.example.com"}'`
      2. Run `docker exec echostate-db psql -U echostate -d echostate -c "SELECT host, normalized_host FROM targets;"`
    Expected Result: `normalized_host` is `example.com` regardless of raw `host`.
    Failure Indicators: normalized_host is NULL or equals raw host.
    Evidence: .omo/evidence/task-6-normalized-host.txt
  ```

  **Evidence to Capture**:
  - [ ] DB query output.

  **Commit**: YES
  - Message: `refactor(scan): populate normalized_host on target upsert — Would you do it for a Scooby Snack?`
  - Files: `internal/handlers/scan.go`

- [x] 7. Add reports HTTP handlers and routes

  **What to do**:
  - Create `internal/handlers/reports.go`.
  - Implement handler methods on `*Handler`:
    - `createReport(c *gin.Context)` — handles `POST /api/reports`.
      - Bind JSON to `models.CreateReportRequest`.
      - Validate that exactly one of `snapshot_id` or `host` is provided; return 400 otherwise.
      - If `snapshot_id` provided, verify snapshot exists; if not, return 404.
      - If `host` provided, normalize it, find latest snapshot by `normalized_host` (join targets → snapshots, order by `scanned_at DESC LIMIT 1`); if none, return 404.
      - Call worker `CreateReport(ctx, snapshotID)`.
      - Return 202 with `models.ReportResponse`.
    - `getReport(c *gin.Context)` — handles `GET /api/reports/:id`.
      - Parse UUID param.
      - Fetch report from DB.
      - Return 200 with `ReportResponse`; include `download_url` only when status is `completed`.
      - Return 404 if report not found.
    - `downloadReport(c *gin.Context)` — handles `GET /api/reports/:id/download`.
      - Fetch report.
      - If status is not `completed`, return 409 with error JSON.
      - If completed, return PDF bytes with `Content-Type: application/pdf` and `Content-Disposition: attachment; filename="echostate-<host>-<date>.pdf"`.
    - `snapshotReport(c *gin.Context)` — handles `GET /api/snapshots/:id/report`.
      - Parse snapshot UUID param.
      - Check for existing completed report for this snapshot; if found, return it directly.
      - If no completed report exists, create a pending report and return 202 with `ReportResponse` pointing to status endpoint (do not block long).
  - Register routes in `internal/handlers/scan.go` `Register()`:
    - `router.POST("/api/reports", h.createReport)`
    - `router.GET("/api/reports/:id", h.getReport)`
    - `router.GET("/api/reports/:id/download", h.downloadReport)`
    - `router.GET("/api/snapshots/:id/report", h.snapshotReport)`
  - Add `*reports.Worker` to the `Handler` struct and initialize it in `Register()`.
  - Start worker in `main.go` after DB setup.

  **Must NOT do**:
  - Do not perform synchronous PDF generation inside handlers.
  - Do not leak internal error details to client.

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: none required

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 3 (with Tasks 8, 9)
  - **Blocks**: 8, 9
  - **Blocked By**: 2, 3, 4, 5, 6

  **References**:
  - `internal/handlers/scan.go` — existing handler patterns, `Register()` function.
  - `internal/models/models.go` — report structs.
  - `internal/reports/worker.go` — `CreateReport`, `GetReport`.
  - `internal/db/db.go` — DB access pattern.

  **Acceptance Criteria**:
  - [ ] All four routes are registered and compile.
  - [ ] `POST /api/reports` returns 202 for valid input.
  - [ ] `GET /api/reports/:id/download` returns PDF bytes for completed reports and 409 otherwise.
  - [ ] `go test ./internal/handlers/...` passes.

  **QA Scenarios**:

  ```
  Scenario: Create report by snapshot_id returns 202
    Tool: Bash
    Preconditions: Docker Compose stack running; a snapshot exists with id 00000000-0000-0000-0000-000000000001
    Steps:
      1. Run `curl -s -i -X POST http://localhost:8080/api/reports -H "Content-Type: application/json" -d '{"snapshot_id":"00000000-0000-0000-0000-000000000001"}'`
    Expected Result: HTTP 202; JSON contains id, status, snapshot_id.
    Failure Indicators: Non-202 status; missing fields.
    Evidence: .omo/evidence/task-7-create-by-snapshot.txt

  Scenario: Download completed report returns PDF bytes
    Tool: Bash
    Preconditions: A completed report exists with id <report_id>
    Steps:
      1. Run `curl -s -i http://localhost:8080/api/reports/<report_id>/download`
    Expected Result: HTTP 200; Content-Type is application/pdf; body starts with `%PDF`.
    Failure Indicators: Non-200 status; wrong Content-Type; body does not start with `%PDF`.
    Evidence: .omo/evidence/task-7-download-pdf.txt
  ```

  **Evidence to Capture**:
  - [ ] curl output for create and download scenarios.

  **Commit**: YES
  - Message: `feat(handlers): add report endpoints and routes — And I would have gotten away with it too!`
  - Files: `internal/handlers/reports.go`, `internal/handlers/scan.go`, `main.go`

- [x] 8. Add unit and integration tests

  **What to do**:
  - Add `internal/handlers/reports_test.go`:
    - Test `createReport` with snapshot_id, with host, with neither/both (400).
    - Test `getReport` for pending/completed/failed/not-found.
    - Test `downloadReport` for completed report, not-completed (409), not-found (404).
    - Test `snapshotReport` creates/reuses report.
  - Add `internal/reports/worker_test.go`:
    - Test worker processes pending reports.
    - Test worker handles renderer errors.
    - Test concurrency limit.
  - Add `internal/pdf/renderer_test.go`:
    - Test renderer output starts with `%PDF`.
    - Test renderer handles empty/missing sections.
  - Add `internal/handlers/scan_test.go` or extend existing tests to verify `normalized_host` is populated.
  - Use `pgxpool` test setup or in-memory mocks for DB-dependent tests where practical.

  **Must NOT do**:
  - Do not require a real browser for report tests.
  - Do not rely on external network services for unit tests.

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: none required

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 3 (with Tasks 7, 9)
  - **Blocks**: 9
  - **Blocked By**: 7

  **References**:
  - `internal/scanner/*_test.go` — existing test patterns.
  - `internal/handlers/reports.go` — handlers under test.
  - `internal/reports/worker.go` — worker under test.
  - `internal/pdf/renderer.go` — renderer under test.

  **Acceptance Criteria**:
  - [ ] `go test ./...` passes.
  - [ ] Report handler tests cover happy path and error cases.
  - [ ] Worker tests cover status transitions and errors.
  - [ ] Renderer tests verify PDF output.

  **QA Scenarios**:

  ```
  Scenario: All unit tests pass
    Tool: Bash
    Preconditions: Implementation complete
    Steps:
      1. Run `go test ./...`
    Expected Result: All tests PASS.
    Failure Indicators: Any FAIL.
    Evidence: .omo/evidence/task-8-unit-tests.txt
  ```

  **Evidence to Capture**:
  - [ ] Test output.

  **Commit**: YES
  - Message: `test(reports): add unit and integration tests — Let's split up, gang!`
  - Files: `internal/handlers/reports_test.go`, `internal/reports/worker_test.go`, `internal/pdf/renderer_test.go`

- [x] 9. Docker Compose end-to-end verification

  **What to do**:
  - Start the full stack:
    - `docker compose down -v`
    - `docker compose up --build -d`
  - Wait for services healthy.
  - Scan a host:
    - `curl -s -X POST http://localhost:8080/api/scan -H "Content-Type: application/json" -d '{"host":"example.com"}'`
  - Extract snapshot ID from response.
  - Create report by snapshot_id:
    - `curl -s -X POST http://localhost:8080/api/reports -H "Content-Type: application/json" -d '{"snapshot_id":"<snapshot_id>"}'`
  - Extract report ID.
  - Poll status until completed (with timeout):
    - `curl -s http://localhost:8080/api/reports/<report_id>`
  - Download PDF:
    - `curl -s -i http://localhost:8080/api/reports/<report_id>/download`
  - Verify body starts with `%PDF`.
  - Test convenience endpoint:
    - `curl -s -i http://localhost:8080/api/snapshots/<snapshot_id>/report`
  - Stop the stack:
    - `docker compose down`

  **Must NOT do**:
  - Do not leave containers running after verification.
  - Do not modify production data.

  **Recommended Agent Profile**:
  - **Category**: `unspecified-high`
  - **Skills**: none required

  **Parallelization**:
  - **Can Run In Parallel**: NO
  - **Parallel Group**: Wave 3 (after Tasks 7, 8)
  - **Blocks**: F1-F4
  - **Blocked By**: 7, 8

  **References**:
  - `docker-compose.yml` — service definitions.
  - `internal/handlers/reports.go` — report endpoints.

  **Acceptance Criteria**:
  - [ ] Stack builds and starts.
  - [ ] Scan creates snapshot and target with normalized_host.
  - [ ] Report creation returns 202.
  - [ ] Polling eventually returns completed.
  - [ ] Download returns valid PDF bytes.
  - [ ] Convenience endpoint returns valid PDF or 202.
  - [ ] Containers stopped after verification.

  **QA Scenarios**:

  ```
  Scenario: Full stack report generation and download
    Tool: Bash
    Preconditions: Docker running
    Steps:
      1. Run `docker compose down -v`
      2. Run `docker compose up --build -d`
      3. Wait for health: `docker compose ps`
      4. Scan example.com and capture snapshot_id
      5. POST /api/reports with snapshot_id and capture report_id
      6. Poll GET /api/reports/<report_id> until status=completed
      7. GET /api/reports/<report_id>/download and save to /tmp/report.pdf
      8. Run `file /tmp/report.pdf` or check first 4 bytes
      9. Run `docker compose down`
    Expected Result: Downloaded file starts with `%PDF`; polling completes within 60s; containers stop cleanly.
    Failure Indicators: Status never completes; file is not PDF; containers fail to stop.
    Evidence: .omo/evidence/task-9-e2e-report.txt
  ```

  **Evidence to Capture**:
  - [ ] `docker compose ps` output.
  - [ ] Scan response.
  - [ ] Report creation response.
  - [ ] Final status response.
  - [ ] Download headers and PDF first bytes.

  **Commit**: YES
  - Message: `test(e2e): verify report endpoints end-to-end — Ruh-roh, Raggy!`
  - Files: no source changes (evidence only)

---

## Final Verification Wave (MANDATORY — after ALL implementation tasks)

> 4 review agents run in PARALLEL. ALL must APPROVE. Present consolidated results to user and get explicit "okay" before completing.

- [x] F1. **Plan Compliance Audit** — `oracle`
  Read the plan end-to-end. For each "Must Have": verify implementation exists (read file, curl endpoint, run command). For each "Must NOT Have": search codebase for forbidden patterns — reject with file:line if found. Check evidence files exist in `.omo/evidence/`. Compare deliverables against plan.
  Output: `Must Have [N/N] | Must NOT Have [N/N] | Tasks [N/N] | VERDICT: APPROVE/REJECT`

- [x] F2. **Code Quality Review** — `unspecified-high`
  Run `go vet ./...` + `go test ./...` + `go build ./...`. Review all changed files for: empty catches, `fmt.Println` in prod, dead code, unused imports, AI slop (excessive comments, over-abstraction, generic names).
  Output: `Build [PASS/FAIL] | Lint [PASS/FAIL] | Tests [N pass/N fail] | Files [N clean/N issues] | VERDICT`

- [x] F3. **Real Manual QA** — `unspecified-high`
  Start from clean state. Execute EVERY QA scenario from EVERY task — follow exact steps, capture evidence. Test cross-task integration (features working together, not isolation). Test edge cases: missing snapshot, missing host, download before completion, invalid report ID. Save to `.omo/evidence/final-qa/`.
  Output: `Scenarios [N/N pass] | Integration [N/N] | Edge Cases [N tested] | VERDICT`

- [x] F4. **Scope Fidelity Check** — `deep`
  For each task: read "What to do", read actual diff (`git log/dev..HEAD`). Verify 1:1 — everything in spec was built (no missing), nothing beyond spec was built (no creep). Check "Must NOT do" compliance. Detect cross-task contamination.
  Output: `Tasks [N/N compliant] | Contamination [CLEAN/N issues] | Unaccounted [CLEAN/N files] | VERDICT`

---

## Commit Strategy

- Use conventional commit prefixes (`chore:`, `feat:`, `test:`, `refactor:`).
- Keep messages concise and written in natural, human language — avoid robotic or overly verbose phrasing.
- A short Scooby-Doo quote at the end is fine for tone.

Example commits for this plan:

- **Task 1**: `chore(deps): pull in maroto v2 for PDF reports`
- **Task 2**: `chore(db): add reports table and normalized_host column`
- **Task 3**: `chore(models): report request/response structs`
- **Task 4**: `feat(pdf): render recon reports with maroto`
- **Task 5**: `feat(reports): async PDF worker with bounded concurrency`
- **Task 6**: `refactor(scan): store normalized_host on upsert`
- **Task 7**: `feat(handlers): report endpoints for snapshots and hosts`
- **Task 8**: `test(reports): cover report create, status, download, and errors`
- **Task 9**: `test(e2e): docker compose report generation flow`

---

## Success Criteria

### Verification Commands
```bash
go test ./...           # Expected: PASS
go vet ./...            # Expected: PASS
go build -o echostate ./main.go  # Expected: success
```

### Final Checklist
- [ ] maroto v2 dependency added and builds cleanly.
- [ ] `reports` table and `targets.normalized_host` migration applied.
- [ ] Report endpoints return correct status codes and PDF bytes.
- [ ] Async job status transitions persisted in PostgreSQL.
- [ ] Tests pass and cover happy path + error cases.
- [ ] Docker Compose E2E passes.
- [ ] No frontend, email, scheduling, or HTML-to-PDF code added.
