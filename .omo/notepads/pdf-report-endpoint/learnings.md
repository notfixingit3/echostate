# PDF Report Endpoint — Learnings

## Task 3 — Report Models (2026-06-15)

- Added `ReportStatus` string type with four constants: `pending`, `running`, `completed`, `failed`.
- `CreateReportRequest` uses pointer fields (`*uuid.UUID`, `*string`) with `omitempty` so callers can provide either a snapshot ID or a host.
- `ReportResponse` is the JSON envelope for API responses; `DownloadURL` and `Error` are omitempty since they're only set after completion or on failure.
- `Report` is the internal DB model; `PDF` uses `json:"-"` to never serialize the binary blob over the wire.
- All types live in `internal/models/models.go` alongside `Target`, `Snapshot`, `ScanRequest`, and `ScanResult`.
- `go build ./...` and `go vet ./...` both pass cleanly.
