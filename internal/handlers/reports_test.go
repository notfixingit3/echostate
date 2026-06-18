package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
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

	schema := fmt.Sprintf("handlers_test_%s", uuid.New().String()[:8])

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
		db:           d,
		reportWorker: reports.New(d, renderer, 1),
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
	require.NoError(t, h.reportWorker.Start(ctx))
	defer h.reportWorker.Stop()

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
	require.NoError(t, h.reportWorker.Start(ctx))
	defer h.reportWorker.Stop()

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

func TestCreateReport_BothFields(t *testing.T) {
	d := setupTestDB(t)
	snapshotID := seedSnapshot(t, d)
	h := newTestHandler(t, d)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body, _ := json.Marshal(map[string]any{"snapshot_id": snapshotID, "host": "example.com"})
	c.Request = httptest.NewRequest(http.MethodPost, "/api/reports", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.createReport(c)

	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Contains(t, w.Body.String(), "provide exactly one of snapshot_id or host")
}

func TestCreateReport_NeitherField(t *testing.T) {
	d := setupTestDB(t)
	h := newTestHandler(t, d)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body, _ := json.Marshal(map[string]any{})
	c.Request = httptest.NewRequest(http.MethodPost, "/api/reports", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.createReport(c)

	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Contains(t, w.Body.String(), "provide exactly one of snapshot_id or host")
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

func insertReport(t *testing.T, d *db.DB, snapshotID uuid.UUID, status models.ReportStatus, pdf []byte) uuid.UUID {
	t.Helper()

	ctx := context.Background()
	var id uuid.UUID
	var completedAt interface{}
	if status == models.ReportCompleted {
		completedAt = time.Now().UTC()
	} else {
		completedAt = nil
	}
	err := d.Pool.QueryRow(ctx, `
		INSERT INTO reports (snapshot_id, status, pdf, completed_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, snapshotID, status, pdf, completedAt).Scan(&id)
	require.NoError(t, err)
	return id
}

func insertFailedReport(t *testing.T, d *db.DB, snapshotID uuid.UUID, errorMessage string) uuid.UUID {
	t.Helper()

	ctx := context.Background()
	var id uuid.UUID
	err := d.Pool.QueryRow(ctx, `
		INSERT INTO reports (snapshot_id, status, error_message, pdf, completed_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, snapshotID, models.ReportFailed, errorMessage, nil, nil).Scan(&id)
	require.NoError(t, err)
	return id
}

func TestGetReport_Pending(t *testing.T) {
	d := setupTestDB(t)
	snapshotID := seedSnapshot(t, d)
	h := newTestHandler(t, d)

	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body, _ := json.Marshal(map[string]any{"snapshot_id": snapshotID})
	c.Request = httptest.NewRequest(http.MethodPost, "/api/reports", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	h.createReport(c)
	require.Equal(t, http.StatusAccepted, w.Code)

	var created models.ReportResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))
	require.Equal(t, models.ReportPending, created.Status)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/reports/"+created.ID.String(), nil)
	c.Params = gin.Params{{Key: "id", Value: created.ID.String()}}
	h.getReport(c)

	require.Equal(t, http.StatusOK, w.Code)
	var resp models.ReportResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, created.ID, resp.ID)
	require.Equal(t, models.ReportPending, resp.Status)
	require.Equal(t, snapshotID, resp.SnapshotID)
	require.Empty(t, resp.DownloadURL)
}

func TestGetReport_Completed(t *testing.T) {
	d := setupTestDB(t)
	snapshotID := seedSnapshot(t, d)
	h := newTestHandler(t, d)
	reportID := insertReport(t, d, snapshotID, models.ReportCompleted, []byte("%PDF completed"))

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/reports/"+reportID.String(), nil)
	c.Params = gin.Params{{Key: "id", Value: reportID.String()}}
	h.getReport(c)

	require.Equal(t, http.StatusOK, w.Code)
	var resp models.ReportResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, reportID, resp.ID)
	require.Equal(t, models.ReportCompleted, resp.Status)
	require.Equal(t, snapshotID, resp.SnapshotID)
	require.Contains(t, resp.DownloadURL, "/api/reports/"+reportID.String()+"/download")
	require.NotNil(t, resp.CompletedAt)
}

func TestGetReport_NotFound(t *testing.T) {
	d := setupTestDB(t)
	h := newTestHandler(t, d)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/reports/"+uuid.New().String(), nil)
	c.Params = gin.Params{{Key: "id", Value: uuid.New().String()}}
	h.getReport(c)

	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetReport_Failed(t *testing.T) {
	d := setupTestDB(t)
	snapshotID := seedSnapshot(t, d)
	h := newTestHandler(t, d)
	internalError := "internal renderer crashed: connection refused"
	reportID := insertFailedReport(t, d, snapshotID, internalError)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/reports/"+reportID.String(), nil)
	c.Params = gin.Params{{Key: "id", Value: reportID.String()}}
	h.getReport(c)

	require.Equal(t, http.StatusOK, w.Code)
	var resp models.ReportResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, reportID, resp.ID)
	require.Equal(t, models.ReportFailed, resp.Status)
	require.Equal(t, "report generation failed", resp.Error)
	require.NotContains(t, resp.Error, internalError)
	require.NotContains(t, w.Body.String(), internalError)
}

func TestDownloadReport_Completed(t *testing.T) {
	d := setupTestDB(t)
	snapshotID := seedSnapshot(t, d)
	h := newTestHandler(t, d)
	pdf := []byte("%PDF fake download")
	reportID := insertReport(t, d, snapshotID, models.ReportCompleted, pdf)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/reports/"+reportID.String()+"/download", nil)
	c.Params = gin.Params{{Key: "id", Value: reportID.String()}}
	h.downloadReport(c)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "application/pdf", w.Header().Get("Content-Type"))
	require.Contains(t, w.Header().Get("Content-Disposition"), "attachment")
	require.Contains(t, w.Header().Get("Content-Disposition"), ".pdf")
	require.Equal(t, pdf, w.Body.Bytes())
}

func TestDownloadReport_NotReady(t *testing.T) {
	d := setupTestDB(t)
	snapshotID := seedSnapshot(t, d)
	h := newTestHandler(t, d)
	reportID := insertReport(t, d, snapshotID, models.ReportPending, nil)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/reports/"+reportID.String()+"/download", nil)
	c.Params = gin.Params{{Key: "id", Value: reportID.String()}}
	h.downloadReport(c)

	require.Equal(t, http.StatusConflict, w.Code)
	require.Contains(t, strings.ToLower(w.Body.String()), "not ready")
}

func TestDownloadReport_NotFound(t *testing.T) {
	d := setupTestDB(t)
	h := newTestHandler(t, d)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/reports/"+uuid.New().String()+"/download", nil)
	c.Params = gin.Params{{Key: "id", Value: uuid.New().String()}}
	h.downloadReport(c)

	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestSnapshotReport_ReusesCompleted(t *testing.T) {
	d := setupTestDB(t)
	snapshotID := seedSnapshot(t, d)
	h := newTestHandler(t, d)
	pdf := []byte("%PDF snapshot completed")
	insertReport(t, d, snapshotID, models.ReportCompleted, pdf)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/snapshots/"+snapshotID.String()+"/report", nil)
	c.Params = gin.Params{{Key: "id", Value: snapshotID.String()}}
	h.snapshotReport(c)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "application/pdf", w.Header().Get("Content-Type"))
	require.Contains(t, w.Header().Get("Content-Disposition"), "attachment")
	require.Contains(t, w.Header().Get("Content-Disposition"), ".pdf")
	require.Equal(t, pdf, w.Body.Bytes())
}

func TestSnapshotReport_CreatesNew(t *testing.T) {
	d := setupTestDB(t)
	snapshotID := seedSnapshot(t, d)
	h := newTestHandler(t, d)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/snapshots/"+snapshotID.String()+"/report", nil)
	c.Params = gin.Params{{Key: "id", Value: snapshotID.String()}}
	h.snapshotReport(c)

	require.Equal(t, http.StatusAccepted, w.Code)
	var resp models.ReportResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, models.ReportPending, resp.Status)
	require.Equal(t, snapshotID, resp.SnapshotID)
	require.NotEqual(t, uuid.Nil, resp.ID)

	var count int
	err := d.Pool.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM reports WHERE snapshot_id = $1
	`, snapshotID).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count)
}
