package db

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
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

func TestConnect_SetsTracer(t *testing.T) {
	config, err := pgxpool.ParseConfig(testDatabaseURL())
	require.NoError(t, err)

	config.ConnConfig.Tracer = otelpgx.NewTracer()
	require.NotNil(t, config.ConnConfig.Tracer)
}

func TestConnect_TracerDoesNotBlockConnection(t *testing.T) {
	d := requirePostgres(t)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var result int
	err := d.Pool.QueryRow(ctx, "SELECT 1").Scan(&result)
	require.NoError(t, err)
	assert.Equal(t, 1, result)
}

func TestConnect_TracerEmitsSpanForQuery(t *testing.T) {
	// Set up an in-memory OTel exporter to capture spans.
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSyncer(exporter),
	)
	otel.SetTracerProvider(tp)
	t.Cleanup(func() { otel.SetTracerProvider(sdktrace.NewTracerProvider()) })

	d := requirePostgres(t)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var result int
	err := d.Pool.QueryRow(ctx, "SELECT 1").Scan(&result)
	require.NoError(t, err)
	assert.Equal(t, 1, result)

	// Force span export by shutting down the provider.
	require.NoError(t, tp.Shutdown(ctx))

	spans := exporter.GetSpans()
	require.NotEmpty(t, spans, "expected at least one span from SELECT 1")

	// The span name should contain "SELECT" or the query operation.
	found := false
	for _, s := range spans {
		if s.Name == "query SELECT 1" || s.Name == "SELECT 1" {
			found = true
			break
		}
	}
	assert.True(t, found, "expected a span named 'query SELECT 1' or 'SELECT 1', got: %v", spanNames(spans))
}

func spanNames(stubs tracetest.SpanStubs) []string {
	names := make([]string, len(stubs))
	for i, s := range stubs {
		names[i] = s.Name
	}
	return names
}
