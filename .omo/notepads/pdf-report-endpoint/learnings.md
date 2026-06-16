# PDF Report Endpoint - Learnings

## Task 2: DB Migration (2026-06-15)
- Added `reports` table with FK to `snapshots(id) ON DELETE CASCADE`
- Added `normalized_host TEXT` column to `targets` (idempotent via `ALTER TABLE ... ADD COLUMN IF NOT EXISTS`)
- Indexes: `idx_targets_normalized_host` and `idx_reports_snapshot_id_status`
- All migrations remain inline in `Migrate()`, following existing pattern
- Verified: `go build ./...` and `go vet ./...` pass
