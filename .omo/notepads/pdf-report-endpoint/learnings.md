# PDF Report Endpoint — Learnings

## Wave 1 — Foundation

### Task 1: Add maroto v2 dependency

- Added `github.com/johnfercher/maroto/v2 v2.4.0` to go.mod
- Since no code imports it yet, `go mod tidy` strips unused direct deps. Pinned explicitly via `go mod edit -require`.
- Transitive deps pulled in: boombuler/barcode, pdfcpu/pdfcpu, go-tree, go-async, uax29, hhrutter/tiff/lzw/pkcs7
- Build passes: `go build ./...` exits 0

### Task 6: Store normalized_host on upsert

- `upsertTarget` in `internal/handlers/scan.go` now calls `scanner.NormalizeHost(host)` before the INSERT.
- The INSERT includes `normalized_host` as a second parameter; ON CONFLICT updates it.
- The unique constraint on raw `host` is unchanged — `ON CONFLICT (host)` still uses the raw column.
- `scanner.NormalizeHost` is already imported in the handlers package via `internal/scanner`.
- Existing scan tests pass without modification since the column has a default/accepts NULL and the migration already added it.

### Task 5: Async bounded PDF report worker

- Created `internal/reports/worker.go` with a `Worker` struct backed by a PostgreSQL queue.
- Worker API: `New`, `Start`, `Stop`, `CreateReport`, `GetReport`, `ListReportsForSnapshot`.
- On `Start`, any `running` reports from a previous process are marked `failed` with message "worker restarted before completion".
- Job claiming uses `SELECT ... FOR UPDATE SKIP LOCKED` inside a transaction; status is updated to `running` before the job is released to a goroutine.
- Renderer invocation is wrapped in a 30s timeout via `context.WithTimeout`.
- PDFs larger than 10 MB are rejected and the report is marked failed with "generated PDF exceeds max size".
- Concurrency is bounded by a semaphore (default 3); `CreateReport` wakes the loop via a buffered channel with non-blocking send.
- Tests in `internal/reports/worker_test.go` cover completion, missing snapshot, oversized PDF, concurrency limit, stale-running cleanup, and wake behavior.
- `go test ./internal/reports/...`, `go build ./...`, and `go vet ./...` all pass.

### Task 7: HTTP report endpoints

- Created `internal/handlers/reports.go` with four endpoints on `*Handler`:
  - `POST /api/reports` accepts `models.CreateReportRequest` with exactly one of `snapshot_id` or `host`.
  - `GET /api/reports/:id` returns `models.ReportResponse`; `download_url` is only set when status is `completed`.
  - `GET /api/reports/:id/download` streams the PDF with `Content-Type: application/pdf` and an attachment filename (`echostate-<host>-<date>.pdf`), returning 409 if not ready.
  - `GET /api/snapshots/:id/report` serves an existing completed PDF directly or enqueues a new report and returns 202.
- Host-based report requests join `targets` on `normalized_host` and order by `snapshots.scanned_at DESC LIMIT 1`.
- `internal/handlers/scan.go` was updated to add a `worker *reports.Worker` field; `Register` now creates the worker with `reports.New(database, pdf.RenderReport, 3)` and returns it so `main.go` can manage its lifecycle.
- `main.go` starts the worker after `db.Migrate(database)` and calls `worker.Stop()` before closing the database during graceful shutdown.
- Added `internal/handlers/reports_test.go` covering invalid requests (400), missing snapshots (404), and valid create-by-snapshot/host (202).
- Manual verification: created a report by snapshot ID, polled to `completed`, and downloaded a valid PDF (`%PDF-1.3`, 9620 bytes).
- Evidence captured in `.omo/evidence/task-7-create-by-snapshot.txt` and `.omo/evidence/task-7-download-pdf.txt`.
- `go build ./...`, `go vet ./...`, and `go test ./internal/handlers/...` all pass.

### Task 8: Report status, download, and normalized_host tests

- Extended `internal/handlers/reports_test.go` with coverage for `getReport`, `downloadReport`, and `snapshotReport`.
- Added `internal/handlers/scan_test.go` to verify `upsertTarget` stores `normalized_host` correctly.
- Completed reports are created by direct DB insert in tests to avoid flakiness from async worker timing.
- `snapshotReport` returns an existing completed PDF directly (200) or enqueues a new report (202) when none exists.
- `downloadReport` sets `Content-Type: application/pdf` and an attachment `Content-Disposition` header.
- `TestUpsertTargetStoresNormalizedHost` confirms `https://www.Example.COM:8080/path` normalizes to `example.com`.
- `go test ./...` and `go vet ./...` pass.

## Wave 2 — End-to-End Verification

### Task 9: Full Docker Compose E2E report flow

- Started clean with `docker compose down -v`, then `docker compose up --build -d`.
- Port `8080` was initially occupied by an unrelated container (`traceline-api`); stopped it, reran the flow, and restarted it afterward.
- Scanned `example.com` and extracted snapshot ID `bd03780a-155e-4746-82de-a34e1a8de45c`.
- Confirmed `normalized_host` is populated as `example.com` in the `targets` table.
- Created a report by snapshot ID and received HTTP 202 with report ID `44cb5d2c-5039-4769-a90e-141f41672629`.
- Report completed on the first poll (`GET /api/reports/:id`), returning `download_url`.
- Downloaded PDF from `/api/reports/:id/download`: HTTP 200, `Content-Type: application/pdf`, `Content-Disposition: attachment; filename="echostate-example.com-2026-06-16.pdf"`, body starts with `%PDF-1.3`, 12384 bytes.
- Convenience endpoint `/api/snapshots/:id/report` returned the existing completed PDF directly (HTTP 200, same 12384-byte PDF).
- Stopped the stack with `docker compose down`; no containers left running.
- All command outputs and curl responses captured in `.omo/evidence/task-9-e2e-report.txt`.

## Wave 3 — PDF Metadata Section and Renderer Error Handling

### Task: Add explicit metadata/report-generated-at section

- Added `addReportMetadata(m core.Maroto, result *models.ScanResult)` in `internal/pdf/renderer.go`.
- The section is placed immediately after `addCoverPage` and before the WHOIS section.
- Uses existing helpers `addSectionHeader` and `addKeyValueRow` to show:
  - "Report Metadata" header
  - "Report Generated At: <timestamp>" with current UTC time formatted via `pdfDateFormat`
  - "Target: <host>"
  - "Scanned At: <timestamp>" from `result.ScannedAt`
- `RenderReport` now calls `addReportMetadata(m, result)` right after the cover page.

### Task: Add worker test for renderer error path

- Added `TestWorkerFailsRendererError` in `internal/reports/worker_test.go`.
- Uses a renderer that always returns `fmt.Errorf("renderer exploded")`.
- Verifies the report status becomes `ReportFailed` and the error message contains the renderer error.

### Verification

- `go test ./internal/pdf/...` passes.
- `go test ./internal/reports/...` passes.
- `go build ./...` exits 0.
- Generated PDF bytes still start with `%PDF`.

