# PDF Report Endpoint — Issues / Findings

## F3 Final QA Re-run — 2026-06-15

### Summary

The full F3 manual QA wave was executed against a clean Docker Compose stack. All functional scenarios for report creation, status polling, PDF download, snapshot convenience endpoint, and edge cases passed. Evidence is saved under `.omo/evidence/final-qa/`.

| Scenario | Result | Evidence |
|----------|--------|----------|
| Docker Compose down -v (clean state) | PASS | `00-compose-down.log` |
| Docker Compose up --build -d | PASS (with warning) | `01-compose-up-build.log` |
| docker compose ps (healthy) | PASS | `02-compose-ps.log` |
| Scan `example.com` creates snapshot | PASS | `05-scan-example.com-network.log` |
| `targets.normalized_host` populated | PASS | `06-normalized-host.log` |
| `POST /api/reports` by snapshot_id returns 202 | PASS | `07-create-report.log` |
| `GET /api/reports/:id` returns completed | PASS | `08-poll-report.log` |
| `GET /api/reports/:id/download` returns PDF | PASS | `09-download-verify.log` |
| `GET /api/snapshots/:id/report` returns PDF | PASS | `10-snapshot-report-verify.log` |
| Missing snapshot_id -> 404 | PASS | `11-edge-cases-part1.log` |
| Missing host -> 404 | PASS | `11-edge-cases-part1.log` |
| Both snapshot_id and host -> 400 | PASS | `11-edge-cases-part1.log` |
| Neither field -> 400 | PASS | `11-edge-cases-part1.log` |
| Download before completion -> 409 | PASS | `12-download-pending.log` |
| Invalid report id -> 400 | PASS | `11-edge-cases-part1.log` |
| docker compose down | PASS | `18-compose-down.log` |

### Issues Found

#### 1. Docker Compose port mapping is misaligned with `ECHOSTATE_PORT`

- **Severity**: Medium
- **Location**: `docker-compose.yml` line 55
- **Current mapping**: `"${ECHOSTATE_PORT:-8080}:8080"`
- **Problem**: When `.env` sets `ECHOSTATE_PORT=8081`, the published port becomes `8081:8080`, but the application inside the container binds to `ECHOSTATE_PORT` (8081). Host traffic to `localhost:8081` is forwarded to container port 8080, where nothing is listening. The API logs confirm it is listening on `:8081`.
- **Evidence**:
  - `02-compose-ps.log` shows `0.0.0.0:8081->8080/tcp`.
  - `03-api-logs-scan.log` shows `EchoState API listening on :8081`.
  - `13-host-port-mapping.log` and `16-external-scan-full.log` show external `curl` to `localhost:8081/api/scan` returns HTTP code 000 / no response.
  - `17-external-health-verbose.log` shows IPv6 localhost connects but receives an empty reply.
- **Impact**: E2E QA had to be performed via an internal Docker network (`curlimages/curl` on `echostate_default`) instead of the documented `http://localhost:8080` / `localhost:8081` host URL.
- **Recommended fix**: Change the `api` service port mapping to use the same container port as the configured app port, e.g. `"${ECHOSTATE_PORT:-8080}:${ECHOSTATE_PORT:-8080}"`.

#### 2. `docker-compose.yml` uses obsolete `version` attribute

- **Severity**: Low
- **Location**: `docker-compose.yml` line 1
- **Problem**: Docker Compose prints a warning on every command: `the attribute "version" is obsolete, it will be ignored, please remove it to avoid potential confusion`.
- **Evidence**: Visible in `00-compose-down.log`, `01-compose-up-build.log`, `02-compose-ps.log`, `18-compose-down.log`.
- **Impact**: Cosmetic only; does not break functionality.
- **Recommended fix**: Remove `version: "3.9"` from the top of `docker-compose.yml`.

#### 3. External IPv6 localhost access can fail

- **Severity**: Low
- **Problem**: `curl -v http://localhost:8081/health` resolved `localhost` to `::1` first, connected, but received an empty reply. Switching to `127.0.0.1:8081` may behave differently depending on the host binding.
- **Evidence**: `17-external-health-verbose.log`
- **Impact**: Intermittent local access issues; likely related to finding #1.

### Snapshot IDs Used

- Snapshot ID: `2d44c566-4f3c-46b0-8f49-f75e91468012`
- Report ID: `5150ac4b-b15c-4cff-b44c-7f1aa827e693`
- Pending report ID (edge case): `a1111111-1111-1111-1111-111111111111`

### Cleanup: Removed duplicate test `TestWorkerFailsRendererError` — 2026-06-15

- **Action**: Deleted `TestWorkerFailsRendererError` (lines 183-210) from `internal/reports/worker_test.go`.
- **Reason**: It was a duplicate of `TestWorkerRendererError` (same logic: renderer returns error, expect `ReportFailed` with `"renderer exploded"` message). The only difference was using `fmt.Errorf` vs `errors.New`, which is semantically identical.
- **Verification**: `go test ./internal/reports/...` passes (0.375s). All other tests intact.

### Verdict

Functional acceptance: **PASS**. All required API behaviors work correctly. The only blocking operational issue is the Docker Compose host port mapping, which prevented host-based curl access and required running QA from inside the Docker network.

## F2 Code Quality Re-run — 2026-06-15

### Scope
Re-run the final code-quality gate after the duplicate test `TestWorkerFailsRendererError` was removed from `internal/reports/worker_test.go`. Verify all static checks pass and review all Go files changed for the PDF report feature.

### Files Reviewed

- `internal/db/db.go`
- `internal/handlers/reports.go`
- `internal/handlers/reports_test.go`
- `internal/handlers/scan.go`
- `internal/handlers/scan_test.go`
- `internal/models/models.go`
- `internal/pdf/renderer.go`
- `internal/pdf/renderer_test.go`
- `internal/reports/worker.go`
- `internal/reports/worker_test.go`
- `main.go`

### Checks Performed

- **Duplicate test check**: `TestWorkerFailsRendererError` is absent from `internal/reports/worker_test.go`; the remaining test function names in that file are unique.
- **Debug print grep** (`fmt.Print`, `log.Print`) across handler/worker/renderer files: no production debug prints found. The only `log.Printf` calls are operational error logging in `internal/reports/worker.go`.
- **Unused imports / dead code**: `go build ./...` exits 0 and a manual review found no unused imports or dead functions/constants.
- **Static checks**:
  - `go vet ./...` — PASS (no output)
  - `go test -count=1 ./...` — PASS (all packages ok)
  - `go build ./...` — PASS (no output)

### Optional Style Note

`internal/reports/worker.go` uses `time.After(pollInterval)` inside the worker loop. Replacing it with `time.NewTicker` would avoid allocating a fresh timer on every iteration, but this is not a functional defect and does not fail any static check.

### Verdict

Build [PASS] | Lint [PASS] | Tests [all PASS / 0 FAIL] | Files [11 clean / 0 issues] | **APPROVE**
