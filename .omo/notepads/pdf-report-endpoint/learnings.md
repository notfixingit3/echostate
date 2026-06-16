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


