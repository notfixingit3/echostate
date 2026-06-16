# PDF Report Endpoint — Learnings

## Wave 1 — Foundation

### Task 1: Add maroto v2 dependency

- Added `github.com/johnfercher/maroto/v2 v2.4.0` to go.mod
- Since no code imports it yet, `go mod tidy` strips unused direct deps. Pinned explicitly via `go mod edit -require`.
- Transitive deps pulled in: boombuler/barcode, pdfcpu/pdfcpu, go-tree, go-async, uax29, hhrutter/tiff/lzw/pkcs7
- Build passes: `go build ./...` exits 0
