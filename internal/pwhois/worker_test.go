package pwhois

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/notfixingit3/echostate/internal/db"
)

func testDatabaseURL() string {
	if u := os.Getenv("DATABASE_URL"); u != "" {
		return u
	}
	return "postgres://echostate:echostate@localhost:5432/echostate?sslmode=disable"
}

func setupTestDB(t *testing.T) *db.DB {
	t.Helper()

	schema := fmt.Sprintf("pwhois_test_%s", uuid.New().String()[:8])

	baseURL := testDatabaseURL()
	sep := "&"
	if !strings.ContainsRune(baseURL, '?') {
		sep = "?"
	}
	schemaURL := fmt.Sprintf("%s%soptions=--search_path%%3D%s", baseURL, sep, schema)

	d, err := db.Connect(schemaURL)
	if err != nil {
		t.Skipf("database not available: %v", err)
	}

	ctx := context.Background()
	if _, err := d.Pool.Exec(ctx, fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schema)); err != nil {
		t.Fatalf("create schema: %v", err)
	}

	if err := db.Migrate(d); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	t.Cleanup(func() {
		if _, err := d.Pool.Exec(ctx, fmt.Sprintf("DROP SCHEMA %s CASCADE", schema)); err != nil {
			t.Logf("warning: dropping schema %s: %v", schema, err)
		}
		d.Close()
	})

	return d
}

// seedSnapshotWithIP inserts a target and snapshot with the given client_ip.
// Returns the snapshot ID.
func seedSnapshotWithIP(t *testing.T, d *db.DB, ip string) {
	t.Helper()

	ctx := context.Background()
	var targetID uuid.UUID
	err := d.Pool.QueryRow(ctx, `
		INSERT INTO targets (host, normalized_host)
		VALUES ($1, $2)
		RETURNING id
	`, "test-"+ip+".example.com", "test-"+ip+".example.com").Scan(&targetID)
	require.NoError(t, err)

	_, err = d.Pool.Exec(ctx, `
		INSERT INTO snapshots (target_id, data_hash, raw_data, client_ip)
		VALUES ($1, $2, $3, $4)
	`, targetID, "hash-"+ip, []byte(`{"host":"test"}`), ip)
	require.NoError(t, err)
}

// countLookedUp returns the number of snapshots with a non-null pwhois_looked_up_at.
func countLookedUp(t *testing.T, d *db.DB) int {
	t.Helper()
	var n int
	err := d.Pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM snapshots WHERE pwhois_looked_up_at IS NOT NULL`).Scan(&n)
	require.NoError(t, err)
	return n
}

// countPwhoisData returns the number of snapshots with non-null pwhois_data.
func countPwhoisData(t *testing.T, d *db.DB) int {
	t.Helper()
	var n int
	err := d.Pool.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM snapshots WHERE pwhois_data IS NOT NULL AND pwhois_data != 'null'`).Scan(&n)
	require.NoError(t, err)
	return n
}

func TestWorker_EnrichesPublicIP(t *testing.T) {
	d := setupTestDB(t)
	seedSnapshotWithIP(t, d, "8.8.8.8")

	mockLookup := func(_ context.Context, ips []string) ([]PWHOISRecord, error) {
		recs := make([]PWHOISRecord, 0, len(ips))
		for _, ip := range ips {
			if ip == "8.8.8.8" {
				recs = append(recs, PWHOISRecord{
					IP:          "8.8.8.8",
					OriginAS:    "AS15169",
					OrgName:     "Google LLC",
					CountryCode: "US",
					City:        "Mountain View",
					Prefix:      "8.8.8.0/24",
				})
			}
		}
		return recs, nil
	}

	w := NewWorker(d, mockLookup, 0)
	ctx := context.Background()
	w.process(ctx)

	require.Equal(t, 1, countLookedUp(t, d), "snapshot should be marked looked-up")
	require.Equal(t, 1, countPwhoisData(t, d), "snapshot should have pwhois_data")

	// Verify the enriched fields.
	var originAS, orgName, countryCode, city, prefix string
	err := d.Pool.QueryRow(ctx,
		`SELECT pwhois_origin_as, pwhois_org_name, pwhois_country_code, pwhois_city, pwhois_prefix
		 FROM snapshots WHERE client_ip = '8.8.8.8'`,
	).Scan(&originAS, &orgName, &countryCode, &city, &prefix)
	require.NoError(t, err)
	require.Equal(t, "AS15169", originAS)
	require.Equal(t, "Google LLC", orgName)
	require.Equal(t, "US", countryCode)
	require.Equal(t, "Mountain View", city)
	require.Equal(t, "8.8.8.0/24", prefix)
}

func TestWorker_SkipsPrivateIP(t *testing.T) {
	d := setupTestDB(t)
	seedSnapshotWithIP(t, d, "127.0.0.1")
	seedSnapshotWithIP(t, d, "10.0.0.5")
	seedSnapshotWithIP(t, d, "192.168.1.1")

	callCount := 0
	mockLookup := func(_ context.Context, ips []string) ([]PWHOISRecord, error) {
		callCount++
		return nil, nil
	}

	w := NewWorker(d, mockLookup, 0)
	ctx := context.Background()
	w.process(ctx)

	// All three should be marked looked-up without calling the lookup.
	require.Equal(t, 3, countLookedUp(t, d), "all private IPs should be marked looked-up")
	require.Equal(t, 0, callCount, "lookup should not be called for private IPs")
	require.Equal(t, 0, countPwhoisData(t, d), "private IPs should not have pwhois_data")
}

func TestWorker_CachesResults(t *testing.T) {
	d := setupTestDB(t)
	seedSnapshotWithIP(t, d, "8.8.8.8")

	var mu sync.Mutex
	lookupCalls := 0
	mockLookup := func(_ context.Context, ips []string) ([]PWHOISRecord, error) {
		mu.Lock()
		lookupCalls++
		mu.Unlock()
		recs := make([]PWHOISRecord, 0, len(ips))
		for _, ip := range ips {
			if ip == "8.8.8.8" {
				recs = append(recs, PWHOISRecord{
					IP:          "8.8.8.8",
					OriginAS:    "AS15169",
					OrgName:     "Google LLC",
					CountryCode: "US",
				})
			}
		}
		return recs, nil
	}

	w := NewWorker(d, mockLookup, 0)
	ctx := context.Background()

	// First call: should look up and cache.
	w.process(ctx)
	require.Equal(t, 1, lookupCalls, "first process should call lookup once")

	// Second call: should use cache, not re-query.
	w.process(ctx)
	require.Equal(t, 1, lookupCalls, "second process should not call lookup (cached)")

	// Still only one snapshot marked looked-up.
	require.Equal(t, 1, countLookedUp(t, d))
}

func TestWorker_HandlesLookupFailure(t *testing.T) {
	d := setupTestDB(t)
	seedSnapshotWithIP(t, d, "8.8.8.8")

	mockLookup := func(_ context.Context, ips []string) ([]PWHOISRecord, error) {
		return nil, fmt.Errorf("connection refused")
	}

	w := NewWorker(d, mockLookup, 0)
	ctx := context.Background()

	// Should not panic.
	require.NotPanics(t, func() {
		w.process(ctx)
	})

	// Snapshot should NOT be marked looked-up (lookup failed, no cache set).
	require.Equal(t, 0, countLookedUp(t, d), "snapshot should not be marked looked-up after failed lookup")
	require.Equal(t, 0, countPwhoisData(t, d), "snapshot should not have pwhois_data after failed lookup")
}

func TestWorker_BatchesMultipleIPs(t *testing.T) {
	d := setupTestDB(t)
	seedSnapshotWithIP(t, d, "1.1.1.1")
	seedSnapshotWithIP(t, d, "8.8.8.8")
	seedSnapshotWithIP(t, d, "9.9.9.9")

	var mu sync.Mutex
	var seenIPs []string
	mockLookup := func(_ context.Context, ips []string) ([]PWHOISRecord, error) {
		mu.Lock()
		seenIPs = append(seenIPs, ips...)
		mu.Unlock()
		recs := make([]PWHOISRecord, 0, len(ips))
		for _, ip := range ips {
			recs = append(recs, PWHOISRecord{IP: ip, OriginAS: "AS00000", OrgName: "Test Org", CountryCode: "ZZ"})
		}
		return recs, nil
	}

	w := NewWorker(d, mockLookup, 0)
	ctx := context.Background()
	w.process(ctx)

	mu.Lock()
	got := seenIPs
	mu.Unlock()

	require.Equal(t, 3, countLookedUp(t, d), "all 3 snapshots should be marked looked-up")
	require.Equal(t, 3, countPwhoisData(t, d), "all 3 snapshots should have pwhois_data")

	require.ElementsMatch(t, []string{"1.1.1.1", "8.8.8.8", "9.9.9.9"}, got,
		"lookup should receive all 3 IPs in a single batch")
}

func TestWorker_StartStop(t *testing.T) {
	d := setupTestDB(t)
	seedSnapshotWithIP(t, d, "8.8.8.8")

	mockLookup := func(_ context.Context, ips []string) ([]PWHOISRecord, error) {
		return []PWHOISRecord{
			{IP: "8.8.8.8", OriginAS: "AS15169", OrgName: "Google LLC", CountryCode: "US"},
		}, nil
	}

	w := NewWorker(d, mockLookup, 0)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	require.NoError(t, w.Start(ctx))

	// Give the worker a moment to process.
	time.Sleep(200 * time.Millisecond)

	w.Stop()

	// Snapshot should have been enriched.
	require.Equal(t, 1, countLookedUp(t, d))
	require.Equal(t, 1, countPwhoisData(t, d))

	// Double-stop should be safe.
	require.NotPanics(t, func() { w.Stop() })
}
