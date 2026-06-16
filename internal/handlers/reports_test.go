package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/notfixingit3/echostate/internal/db"
	"github.com/notfixingit3/echostate/internal/models"
	"github.com/notfixingit3/echostate/internal/reports"
)

func testDatabaseURL() string {
	if u := os.Getenv("DATABASE_URL"); u != "" {
		return u
	}
	return "postgres://echostate:echostate@localhost:5432/echostate?sslmode=disable"
}

func setupTestDB(t *testing.T) *db.DB {
	t.Helper()

	d, err := db.Connect(testDatabaseURL())
	if err != nil {
		t.Skipf("database not available: %v", err)
	}
	t.Cleanup(func() { d.Close() })

	if err := db.Migrate(d); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	ctx := context.Background()
	if _, err := d.Pool.Exec(ctx, "DELETE FROM reports"); err != nil {
		t.Fatalf("clean reports: %v", err)
	}
	if _, err := d.Pool.Exec(ctx, "DELETE FROM snapshots"); err != nil {
		t.Fatalf("clean snapshots: %v", err)
	}
	if _, err := d.Pool.Exec(ctx, "DELETE FROM targets"); err != nil {
		t.Fatalf("clean targets: %v", err)
	}

	return d
}

func seedSnapshot(t *testing.T, d *db.DB) uuid.UUID {
	t.Helper()

	ctx := context.Background()
	var targetID uuid.UUID
	err := d.Pool.QueryRow(ctx, `
		INSERT INTO targets (host, normalized_host)
		VALUES ($1, $2)
		RETURNING id
	`, "example.com", "example.com").Scan(&targetID)
	require.NoError(t, err)

	result := models.ScanResult{
		Host:      "example.com",
		ScannedAt: time.Now().UTC(),
		WHOIS:     map[string]any{"domain": "example.com"},
		ASN:       map[string]any{"asn": "15169"},
		Web:       map[string]any{"title": "Example"},
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

func newTestHandler(t *testing.T, d *db.DB) *Handler {
	t.Helper()
	renderer := func(*models.ScanResult) ([]byte, error) {
		return []byte("%PDF fake"), nil
	}
	return &Handler{
		db:     d,
		worker: reports.New(d, renderer, 1),
	}
}

func TestCreateReport_InvalidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/reports", bytes.NewBufferString("{}"))
	c.Request.Header.Set("Content-Type", "application/json")

	h := &Handler{}
	h.createReport(c)

	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Contains(t, w.Body.String(), "provide exactly one of snapshot_id or host")
}

func TestCreateReport_SnapshotNotFound(t *testing.T) {
	d := setupTestDB(t)
	h := newTestHandler(t, d)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body, _ := json.Marshal(map[string]any{"snapshot_id": uuid.New()})
	c.Request = httptest.NewRequest(http.MethodPost, "/api/reports", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.createReport(c)

	require.Equal(t, http.StatusNotFound, w.Code)
	require.Contains(t, w.Body.String(), "snapshot not found")
}

func TestCreateReport_BySnapshot(t *testing.T) {
	d := setupTestDB(t)
	snapshotID := seedSnapshot(t, d)
	h := newTestHandler(t, d)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	require.NoError(t, h.worker.Start(ctx))
	defer h.worker.Stop()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body, _ := json.Marshal(map[string]any{"snapshot_id": snapshotID})
	c.Request = httptest.NewRequest(http.MethodPost, "/api/reports", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.createReport(c)

	require.Equal(t, http.StatusAccepted, w.Code)
	require.Contains(t, w.Body.String(), "pending")
}

func TestCreateReport_ByHost(t *testing.T) {
	d := setupTestDB(t)
	seedSnapshot(t, d)
	h := newTestHandler(t, d)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	require.NoError(t, h.worker.Start(ctx))
	defer h.worker.Stop()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body, _ := json.Marshal(map[string]any{"host": "example.com"})
	c.Request = httptest.NewRequest(http.MethodPost, "/api/reports", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.createReport(c)

	require.Equal(t, http.StatusAccepted, w.Code)
	require.Contains(t, w.Body.String(), "pending")
}

func TestGetReport_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/reports/bad-id", nil)
	c.Params = gin.Params{{Key: "id", Value: "bad-id"}}

	h := &Handler{}
	h.getReport(c)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSnapshotReport_SnapshotNotFound(t *testing.T) {
	d := setupTestDB(t)
	h := newTestHandler(t, d)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/snapshots/"+uuid.New().String()+"/report", nil)
	c.Params = gin.Params{{Key: "id", Value: uuid.New().String()}}

	h.snapshotReport(c)

	require.Equal(t, http.StatusNotFound, w.Code)
}
