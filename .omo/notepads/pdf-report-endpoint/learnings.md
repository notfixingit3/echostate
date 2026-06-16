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
