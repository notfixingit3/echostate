package db

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func testDatabaseURL() string {
	if u := os.Getenv("DATABASE_URL"); u != "" {
		return u
	}
	return "postgres://echostate:echostate@localhost:5432/echostate?sslmode=disable"
}

func requirePostgres(t *testing.T) *DB {
	t.Helper()

	d, err := Connect(testDatabaseURL())
	if err != nil {
		t.Skipf("database not available: %v", err)
	}
	t.Cleanup(func() { d.Close() })
	return d
}

func TestConnect_Success(t *testing.T) {
	d := requirePostgres(t)
	require.NotNil(t, d)
	require.NotNil(t, d.Pool)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	require.NoError(t, d.Pool.Ping(ctx))
}

func TestConnect_InvalidURL(t *testing.T) {
	d, err := Connect("not-a-valid-postgres-url")
	require.Error(t, err)
	require.Nil(t, d)
}

func TestConnect_UnreachableHost(t *testing.T) {
	// Port 1 is unlikely to accept connections, so the ping should fail.
	d, err := Connect("postgres://echostate:echostate@localhost:1/echostate?sslmode=disable")
	require.Error(t, err)
	require.Nil(t, d)
}

func TestMigrate_CreatesTablesAndIsIdempotent(t *testing.T) {
	d := requirePostgres(t)

	require.NoError(t, Migrate(d))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, table := range []string{"targets", "snapshots", "reports"} {
		var exists bool
		err := d.Pool.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1
				FROM information_schema.tables
				WHERE table_schema = 'public'
				  AND table_name = $1
			)
		`, table).Scan(&exists)
		require.NoError(t, err)
		require.True(t, exists, "table %q should exist after migration", table)
	}

	// Verify expected columns exist.
	type columnCheck struct {
		table  string
		column string
	}
	for _, c := range []columnCheck{
		{"targets", "host"},
		{"targets", "normalized_host"},
		{"snapshots", "target_id"},
		{"snapshots", "raw_data"},
		{"snapshots", "data_hash"},
		{"snapshots", "pwhois_data"},
		{"snapshots", "pwhois_looked_up_at"},
		{"snapshots", "pwhois_origin_as"},
		{"snapshots", "pwhois_org_name"},
		{"snapshots", "pwhois_country_code"},
		{"snapshots", "pwhois_city"},
		{"snapshots", "pwhois_prefix"},
		{"reports", "snapshot_id"},
		{"reports", "status"},
		{"reports", "pdf"},
	} {
		var exists bool
		err := d.Pool.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1
				FROM information_schema.columns
				WHERE table_schema = 'public'
				  AND table_name = $1
				  AND column_name = $2
			)
		`, c.table, c.column).Scan(&exists)
		require.NoError(t, err)
		require.True(t, exists, "column %q.%q should exist after migration", c.table, c.column)
	}

	// Verify expected indexes exist.
	for _, idx := range []string{
		"idx_snapshots_target_id_scanned_at",
		"idx_snapshots_data_hash",
		"idx_targets_normalized_host",
		"idx_reports_snapshot_id_status",
		"idx_snapshots_pwhois_country_code",
		"idx_snapshots_pwhois_origin_as",
	} {
		var exists bool
		err := d.Pool.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1
				FROM pg_indexes
				WHERE schemaname = 'public'
				  AND indexname = $1
			)
		`, idx).Scan(&exists)
		require.NoError(t, err)
		require.True(t, exists, "index %q should exist after migration", idx)
	}

	// Running migrate a second time must succeed without side effects.
	require.NoError(t, Migrate(d))
}

func TestClose_DoesNotPanic(t *testing.T) {
	d := requirePostgres(t)

	require.NotPanics(t, func() {
		d.Close()
	})
}
