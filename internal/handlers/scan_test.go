package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/notfixingit3/echostate/internal/config"
	"github.com/notfixingit3/echostate/internal/db"
	"github.com/notfixingit3/echostate/internal/middleware"
	"github.com/notfixingit3/echostate/internal/models"
	"github.com/notfixingit3/echostate/internal/scanner"
)

// mockScanRunner is a test double for scanner.Run that avoids live network or
// browser dependencies.
type mockScanRunner struct {
	result *models.ScanResult
	err    error
}

func (m *mockScanRunner) Run(ctx context.Context, host string) (*models.ScanResult, error) {
	return m.result, m.err
}

func testConfig() *config.Config {
	return &config.Config{
		Env:         "test",
		FrontendURL: "http://localhost:3001",
	}
}

func newScanTestHandler(t *testing.T, d *db.DB, runner scanRunner) *Handler {
	t.Helper()
	return &Handler{
		db:      d,
		config:  testConfig(),
		scanner: runner,
	}
}

func newDBTestHandler(t *testing.T, d *db.DB) *Handler {
	t.Helper()
	return &Handler{
		db:     d,
		config: testConfig(),
	}
}

func TestHealth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	h := &Handler{config: &config.Config{Env: "test"}}
	h.health(c)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), `"status":"ok"`)
	require.Contains(t, w.Body.String(), `"env":"test"`)
}

func TestCreateScan_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/scan", bytes.NewBufferString("{not json"))
	c.Request.Header.Set("Content-Type", "application/json")

	h := &Handler{}
	h.createScan(c)

	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Contains(t, w.Body.String(), "error")
}

func TestCreateScan_MissingHost(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body, _ := json.Marshal(map[string]any{})
	c.Request = httptest.NewRequest(http.MethodPost, "/api/scan", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h := &Handler{}
	h.createScan(c)

	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Contains(t, w.Body.String(), "error")
}

func TestCreateScan_EmptyHost(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body, _ := json.Marshal(map[string]any{"host": ""})
	c.Request = httptest.NewRequest(http.MethodPost, "/api/scan", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h := &Handler{}
	h.createScan(c)

	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Contains(t, w.Body.String(), "error")
}

func TestCreateScan_ScannerFailure(t *testing.T) {
	d := setupTestDB(t)
	runner := &mockScanRunner{err: errors.New("connection refused")}
	h := newScanTestHandler(t, d, runner)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body, _ := json.Marshal(map[string]any{"host": "example.com"})
	c.Request = httptest.NewRequest(http.MethodPost, "/api/scan", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.createScan(c)

	require.Equal(t, http.StatusInternalServerError, w.Code)
	require.Contains(t, w.Body.String(), "scan failed")
}

func TestCreateScan_ValidHost(t *testing.T) {
	d := setupTestDB(t)

	result := &models.ScanResult{
		Host:      "example.com",
		ScannedAt: time.Now().UTC(),
		WHOIS:     map[string]any{"domain": "example.com"},
		ASN:       map[string]any{"asn": "15169"},
	}
	runner := &mockScanRunner{result: result}
	h := newScanTestHandler(t, d, runner)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body, _ := json.Marshal(map[string]any{"host": "example.com"})
	c.Request = httptest.NewRequest(http.MethodPost, "/api/scan", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.createScan(c)

	require.Equal(t, http.StatusOK, w.Code)

	var snapshot models.Snapshot
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &snapshot))
	require.NotEqual(t, uuid.Nil, snapshot.ID)
	require.Equal(t, "example.com", snapshot.RawData["host"])

	var targetID uuid.UUID
	err := d.Pool.QueryRow(context.Background(), `
		SELECT id FROM targets WHERE normalized_host = $1
	`, "example.com").Scan(&targetID)
	require.NoError(t, err)
	require.Equal(t, targetID, snapshot.TargetID)

	var count int
	err = d.Pool.QueryRow(context.Background(), `
		SELECT COUNT(*) FROM snapshots WHERE target_id = $1
	`, targetID).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

func TestStoreSnapshot_NewTarget(t *testing.T) {
	d := setupTestDB(t)
	h := newDBTestHandler(t, d)

	ctx := context.Background()
	targetID, err := h.upsertTarget(ctx, "example.com")
	require.NoError(t, err)

	result := &models.ScanResult{
		Host:      "example.com",
		ScannedAt: time.Now().UTC(),
		WHOIS:     map[string]any{"domain": "example.com"},
	}
	hash, err := scanner.Hash(result)
	require.NoError(t, err)

	snapshot, err := h.storeSnapshot(ctx, targetID, hash, result, "")
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, snapshot.ID)
	require.Equal(t, targetID, snapshot.TargetID)
	require.Equal(t, hash, snapshot.DataHash)
	require.Empty(t, snapshot.Changes)
	require.Empty(t, snapshot.ClientIP)
}

func TestStoreSnapshot_SameHashUpdatesLastSeen(t *testing.T) {
	d := setupTestDB(t)
	h := newDBTestHandler(t, d)
	ctx := context.Background()

	targetID, err := h.upsertTarget(ctx, "example.com")
	require.NoError(t, err)

	result := &models.ScanResult{
		Host:      "example.com",
		ScannedAt: time.Now().UTC(),
		WHOIS:     map[string]any{"domain": "example.com"},
	}
	hash, err := scanner.Hash(result)
	require.NoError(t, err)

	s1, err := h.storeSnapshot(ctx, targetID, hash, result, "")
	require.NoError(t, err)

	// Small sleep to ensure last_seen moves forward.
	time.Sleep(5 * time.Millisecond)

	s2, err := h.storeSnapshot(ctx, targetID, hash, result, "")
	require.NoError(t, err)

	require.Equal(t, s1.ID, s2.ID)
	require.True(t, s2.LastSeen.After(s1.LastSeen), "last_seen should be updated")

	var count int
	err = d.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM snapshots WHERE target_id = $1
	`, targetID).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

func TestStoreSnapshot_DifferentHashInsertsWithChanges(t *testing.T) {
	d := setupTestDB(t)
	h := newDBTestHandler(t, d)
	ctx := context.Background()

	targetID, err := h.upsertTarget(ctx, "example.com")
	require.NoError(t, err)

	first := &models.ScanResult{
		Host:      "example.com",
		ScannedAt: time.Now().UTC(),
		WHOIS:     map[string]any{"domain": "example.com"},
	}
	hash1, err := scanner.Hash(first)
	require.NoError(t, err)

	s1, err := h.storeSnapshot(ctx, targetID, hash1, first, "")
	require.NoError(t, err)

	second := &models.ScanResult{
		Host:      "example.com",
		ScannedAt: time.Now().UTC(),
		WHOIS:     map[string]any{"domain": "example.com"},
		ASN:       map[string]any{"asn": "15169"},
	}
	hash2, err := scanner.Hash(second)
	require.NoError(t, err)
	require.NotEqual(t, hash1, hash2)

	s2, err := h.storeSnapshot(ctx, targetID, hash2, second, "")
	require.NoError(t, err)
	require.NotEqual(t, s1.ID, s2.ID)
	require.Contains(t, s2.Changes, "added asn")

	var count int
	err = d.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM snapshots WHERE target_id = $1
	`, targetID).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 2, count)
}

func TestStoreSnapshot_DatabaseError(t *testing.T) {
	d := setupTestDB(t)
	h := newDBTestHandler(t, d)
	ctx := context.Background()

	targetID := uuid.New()
	result := &models.ScanResult{Host: "example.com"}
	hash, err := scanner.Hash(result)
	require.NoError(t, err)

	_, err = h.storeSnapshot(ctx, targetID, hash, result, "")
	require.Error(t, err)
}

func TestComputeChanges_NoPrevious(t *testing.T) {
	current := &models.ScanResult{
		Host:      "example.com",
		ScannedAt: time.Now().UTC(),
		WHOIS:     map[string]any{"domain": "example.com"},
	}

	changes := computeChanges(map[string]any{}, current)
	require.Empty(t, changes)
}

func TestComputeChanges_Identical(t *testing.T) {
	current := &models.ScanResult{
		Host:      "example.com",
		ScannedAt: time.Now().UTC(),
		WHOIS:     map[string]any{"domain": "example.com"},
	}
	data, err := json.Marshal(current)
	require.NoError(t, err)

	previous := map[string]any{}
	require.NoError(t, json.Unmarshal(data, &previous))

	changes := computeChanges(previous, current)
	require.Empty(t, changes)
}

func TestComputeChanges_DetectsAddedRemovedChanged(t *testing.T) {
	previous := map[string]any{
		"host":  "example.com",
		"whois": map[string]any{"domain": "example.com"},
		"asn":   map[string]any{"asn": "15169"},
	}
	current := &models.ScanResult{
		Host:      "example.com",
		ScannedAt: time.Now().UTC(),
		WHOIS:     map[string]any{"domain": "example.org"},
		Web:       map[string]any{"title": "Example"},
	}

	changes := computeChanges(previous, current)

	require.Contains(t, changes, "changed whois")
	require.Contains(t, changes, "removed asn")
	require.Contains(t, changes, "added web")
}

func TestComputeChanges_MarshalCurrentError(t *testing.T) {
	previous := map[string]any{"host": "example.com"}
	current := &models.ScanResult{
		Host:      "example.com",
		ScannedAt: time.Now().UTC(),
		WHOIS:     map[string]any{"bad": make(chan int)},
	}

	changes := computeChanges(previous, current)
	require.Len(t, changes, 1)
	require.Contains(t, changes[0], "marshal current:")
}

func TestHandler_FrontendURL_NilConfig(t *testing.T) {
	h := &Handler{}
	require.Equal(t, "", h.frontendURL())
}

func TestComputeChanges_MarshalPreviousValueError(t *testing.T) {
	previous := map[string]any{
		"whois": make(chan int),
	}
	current := &models.ScanResult{
		Host:      "example.com",
		ScannedAt: time.Now().UTC(),
		WHOIS:     map[string]any{"domain": "example.com"},
	}

	changes := computeChanges(previous, current)

	found := false
	for _, change := range changes {
		if strings.Contains(change, "marshal previous whois:") {
			found = true
			break
		}
	}
	require.True(t, found, "expected a marshal-previous-whois error, got %v", changes)
}

func TestCreateScan_RateLimitExceeded(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rl := middleware.NewRateLimiter()
	defer rl.Stop()

	router := gin.New()
	router.POST("/api/scan", rl.Middleware(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	body, _ := json.Marshal(map[string]any{"host": "example.com"})

	var lastRecorder *httptest.ResponseRecorder
	for i := 0; i < 31; i++ {
		lastRecorder = httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/scan", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.RemoteAddr = "10.0.0.1:12345"
		router.ServeHTTP(lastRecorder, req)
	}

	require.Equal(t, http.StatusTooManyRequests, lastRecorder.Code)
	require.Contains(t, lastRecorder.Body.String(), "rate limit exceeded")
}
