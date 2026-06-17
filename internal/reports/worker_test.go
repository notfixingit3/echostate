package reports

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/notfixingit3/echostate/internal/db"
	"github.com/notfixingit3/echostate/internal/models"
	"github.com/notfixingit3/echostate/internal/pdf"
)

func schemaName(prefix string) string {
	return fmt.Sprintf("%s_%s", prefix, uuid.New().String()[:8])
}

func testDatabaseURL() string {
	if u := os.Getenv("DATABASE_URL"); u != "" {
		return u
	}
	return "postgres://echostate:echostate@localhost:5432/echostate?sslmode=disable"
}

func setupTestDB(t *testing.T) *db.DB {
	t.Helper()

	schema := schemaName("reports_test")

	// Connect with search_path set via connection options so every pool
	// connection automatically targets the isolated schema.
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

func seedSnapshot(t *testing.T, d *db.DB) uuid.UUID {
	return seedSnapshotWithHost(t, d, "example.com")
}

func seedSnapshotWithHost(t *testing.T, d *db.DB, host string) uuid.UUID {
	t.Helper()

	ctx := context.Background()
	var targetID uuid.UUID
	err := d.Pool.QueryRow(ctx, `
		INSERT INTO targets (host, normalized_host)
		VALUES ($1, $2)
		RETURNING id
	`, host, host).Scan(&targetID)
	require.NoError(t, err)

	result := models.ScanResult{
		Host:      host,
		ScannedAt: time.Now().UTC(),
		WHOIS: map[string]any{
			"domain":    host,
			"registrar": "Example Registrar",
		},
		ASN: map[string]any{
			"asn": "15169",
		},
		Web: map[string]any{
			"title": "Example Domain",
		},
	}
	raw, err := json.Marshal(result)
	require.NoError(t, err)

	var snapshotID uuid.UUID
	err = d.Pool.QueryRow(ctx, `
		INSERT INTO snapshots (target_id, data_hash, raw_data)
		VALUES ($1, $2, $3)
		RETURNING id
	`, targetID, "deadbeef", raw).Scan(&snapshotID)
	require.NoError(t, err)

	return snapshotID
}

func TestWorkerCompletes(t *testing.T) {
	d := setupTestDB(t)
	snapshotID := seedSnapshot(t, d)

	w := New(d, pdf.RenderReport, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	require.NoError(t, w.Start(ctx))
	defer w.Stop()

	reportID, err := w.CreateReport(ctx, snapshotID)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, reportID)

	var report *models.Report
	require.Eventually(t, func() bool {
		var err error
		report, err = w.GetReport(ctx, reportID)
		require.NoError(t, err)
		return report.Status == models.ReportCompleted || report.Status == models.ReportFailed
	}, 15*time.Second, 100*time.Millisecond)

	require.Equal(t, models.ReportCompleted, report.Status)
	require.NotEmpty(t, report.PDF)
	require.Greater(t, len(report.PDF), 4)
	require.Equal(t, "%PDF", string(report.PDF[:4]))
	require.NotNil(t, report.CompletedAt)
}

func insertReportWithSnapshotID(t *testing.T, d *db.DB, snapshotID uuid.UUID) uuid.UUID {
	t.Helper()
	ctx := context.Background()

	_, err := d.Pool.Exec(ctx, "ALTER TABLE reports DISABLE TRIGGER ALL")
	require.NoError(t, err)
	defer func() {
		_, err := d.Pool.Exec(ctx, "ALTER TABLE reports ENABLE TRIGGER ALL")
		require.NoError(t, err)
	}()

	var reportID uuid.UUID
	err = d.Pool.QueryRow(ctx, `
		INSERT INTO reports (snapshot_id, status)
		VALUES ($1, $2)
		RETURNING id
	`, snapshotID, models.ReportPending).Scan(&reportID)
	require.NoError(t, err)

	return reportID
}

func TestWorkerFailsMissingSnapshot(t *testing.T) {
	d := setupTestDB(t)

	renderer := func(*models.ScanResult) ([]byte, error) {
		return []byte("should not be called"), nil
	}

	w := New(d, renderer, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	reportID := insertReportWithSnapshotID(t, d, uuid.New())

	require.NoError(t, w.Start(ctx))
	defer w.Stop()

	var report *models.Report
	require.Eventually(t, func() bool {
		var err error
		report, err = w.GetReport(ctx, reportID)
		require.NoError(t, err)
		if report == nil {
			return false
		}
		return report.Status == models.ReportCompleted || report.Status == models.ReportFailed
	}, 15*time.Second, 100*time.Millisecond)

	require.Equal(t, models.ReportFailed, report.Status)
	require.Contains(t, report.ErrorMessage, "snapshot not found")
}

func TestWorkerFailsOversizedPDF(t *testing.T) {
	d := setupTestDB(t)
	snapshotID := seedSnapshot(t, d)

	renderer := func(*models.ScanResult) ([]byte, error) {
		return make([]byte, maxPDFSize+1), nil
	}

	w := New(d, renderer, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	require.NoError(t, w.Start(ctx))
	defer w.Stop()

	reportID, err := w.CreateReport(ctx, snapshotID)
	require.NoError(t, err)

	var report *models.Report
	require.Eventually(t, func() bool {
		var err error
		report, err = w.GetReport(ctx, reportID)
		require.NoError(t, err)
		return report.Status == models.ReportCompleted || report.Status == models.ReportFailed
	}, 15*time.Second, 100*time.Millisecond)

	require.Equal(t, models.ReportFailed, report.Status)
	require.Contains(t, report.ErrorMessage, "generated PDF exceeds max size")
}

func TestWorkerRendererError(t *testing.T) {
	d := setupTestDB(t)
	snapshotID := seedSnapshot(t, d)

	rendererErr := errors.New("renderer exploded")
	renderer := func(*models.ScanResult) ([]byte, error) {
		return nil, rendererErr
	}

	w := New(d, renderer, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	require.NoError(t, w.Start(ctx))
	defer w.Stop()

	reportID, err := w.CreateReport(ctx, snapshotID)
	require.NoError(t, err)

	var report *models.Report
	require.Eventually(t, func() bool {
		var err error
		report, err = w.GetReport(ctx, reportID)
		require.NoError(t, err)
		return report.Status == models.ReportCompleted || report.Status == models.ReportFailed
	}, 15*time.Second, 100*time.Millisecond)

	require.Equal(t, models.ReportFailed, report.Status)
	require.Contains(t, report.ErrorMessage, rendererErr.Error())
}

func TestWorkerConcurrencyLimit(t *testing.T) {
	d := setupTestDB(t)

	const (
		jobCount       = 6
		maxConcurrency = 2
	)

	snapshotIDs := make([]uuid.UUID, jobCount)
	for i := 0; i < jobCount; i++ {
		snapshotIDs[i] = seedSnapshotWithHost(t, d, fmt.Sprintf("host-%d.example.com", i))
	}

	var active int64
	var maxActive int64
	var mu sync.Mutex
	gate := make(chan struct{})

	renderer := func(*models.ScanResult) ([]byte, error) {
		current := atomic.AddInt64(&active, 1)

		mu.Lock()
		if current > maxActive {
			maxActive = current
		}
		mu.Unlock()

		// Block until the test releases the gate so concurrency is observable.
		<-gate

		atomic.AddInt64(&active, -1)
		return []byte("%PDF fake pdf"), nil
	}

	w := New(d, renderer, maxConcurrency)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	require.NoError(t, w.Start(ctx))
	defer w.Stop()

	reportIDs := make([]uuid.UUID, jobCount)
	for i := 0; i < jobCount; i++ {
		id, err := w.CreateReport(ctx, snapshotIDs[i])
		require.NoError(t, err)
		reportIDs[i] = id
	}

	// Wait briefly for workers to pick up jobs and block on the gate.
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	observedMax := maxActive
	mu.Unlock()
	require.LessOrEqual(t, observedMax, int64(maxConcurrency), "concurrency exceeded limit")

	// Release all blocked renders.
	for i := 0; i < jobCount; i++ {
		gate <- struct{}{}
	}

	require.Eventually(t, func() bool {
		for _, id := range reportIDs {
			r, err := w.GetReport(ctx, id)
			require.NoError(t, err)
			if r.Status != models.ReportCompleted && r.Status != models.ReportFailed {
				return false
			}
		}
		return true
	}, 15*time.Second, 100*time.Millisecond)

	for _, id := range reportIDs {
		r, err := w.GetReport(ctx, id)
		require.NoError(t, err)
		require.Equal(t, models.ReportCompleted, r.Status)
	}
}

func TestWorkerMarksStaleRunningAsFailed(t *testing.T) {
	d := setupTestDB(t)
	snapshotID := seedSnapshot(t, d)

	ctx := context.Background()
	var reportID uuid.UUID
	err := d.Pool.QueryRow(ctx, `
		INSERT INTO reports (snapshot_id, status)
		VALUES ($1, $2)
		RETURNING id
	`, snapshotID, models.ReportRunning).Scan(&reportID)
	require.NoError(t, err)

	renderer := func(*models.ScanResult) ([]byte, error) {
		return []byte("%PDF"), nil
	}

	w := New(d, renderer, 1)
	workerCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	require.NoError(t, w.Start(workerCtx))
	defer w.Stop()

	report, err := w.GetReport(ctx, reportID)
	require.NoError(t, err)
	require.Equal(t, models.ReportFailed, report.Status)
	require.Contains(t, report.ErrorMessage, "worker restarted before completion")
}

func TestCreateReportWakesWorker(t *testing.T) {
	d := setupTestDB(t)
	snapshotID := seedSnapshot(t, d)

	called := make(chan struct{}, 1)
	renderer := func(*models.ScanResult) ([]byte, error) {
		called <- struct{}{}
		return []byte("%PDF"), nil
	}

	w := New(d, renderer, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	require.NoError(t, w.Start(ctx))
	defer w.Stop()

	_, err := w.CreateReport(ctx, snapshotID)
	require.NoError(t, err)

	select {
	case <-called:
		// worker woke and processed the job promptly
	case <-time.After(5 * time.Second):
		t.Fatal("worker did not wake and process report")
	}
}
