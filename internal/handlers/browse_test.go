package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/notfixingit3/echostate/internal/db"
	"github.com/notfixingit3/echostate/internal/models"
)

func seedTarget(t *testing.T, d *db.DB, host string) uuid.UUID {
	t.Helper()
	ctx := context.Background()
	var id uuid.UUID
	err := d.Pool.QueryRow(ctx, `
		INSERT INTO targets (host, normalized_host)
		VALUES ($1, $2)
		RETURNING id
	`, host, host).Scan(&id)
	require.NoError(t, err)
	return id
}

func seedSnapshotForTarget(t *testing.T, d *db.DB, targetID uuid.UUID, rawData map[string]any, clientIP string) uuid.UUID {
	t.Helper()
	ctx := context.Background()
	raw, err := json.Marshal(rawData)
	require.NoError(t, err)

	var id uuid.UUID
	err = d.Pool.QueryRow(ctx, `
		INSERT INTO snapshots (target_id, data_hash, raw_data, client_ip, pwhois_data)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, targetID, "hash-"+uuid.New().String()[:8], raw, clientIP, `{"asn": "15169", "org": "Google LLC"}`).Scan(&id)
	require.NoError(t, err)
	return id
}

func TestListTargets_Pagination(t *testing.T) {
	d := setupTestDB(t)
	h := &Handler{db: d}

	for i := 0; i < 3; i++ {
		seedTarget(t, d, fmt.Sprintf("target-%d.example.com", i))
	}

	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/targets?page=1&limit=2", nil)
	h.listTargets(c)

	require.Equal(t, http.StatusOK, w.Code)
	var resp models.PaginatedResponse[models.TargetSummary]
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Len(t, resp.Data, 2)
	require.Equal(t, 3, resp.Total)
	require.Equal(t, 1, resp.Page)
	require.Equal(t, 2, resp.Limit)

	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Request = httptest.NewRequest(http.MethodGet, "/api/targets?page=2&limit=2", nil)
	h.listTargets(c2)

	require.Equal(t, http.StatusOK, w2.Code)
	var resp2 models.PaginatedResponse[models.TargetSummary]
	require.NoError(t, json.Unmarshal(w2.Body.Bytes(), &resp2))
	require.Len(t, resp2.Data, 1)
	require.Equal(t, 3, resp2.Total)
}

func TestListTargets_Search(t *testing.T) {
	d := setupTestDB(t)
	h := &Handler{db: d}

	seedTarget(t, d, "alpha.example.com")
	seedTarget(t, d, "beta.example.com")
	seedTarget(t, d, "gamma.test.net")

	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/targets?q=example", nil)
	h.listTargets(c)

	require.Equal(t, http.StatusOK, w.Code)
	var resp models.PaginatedResponse[models.TargetSummary]
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Len(t, resp.Data, 2)
	require.Equal(t, 2, resp.Total)
}

func TestListTargets_Empty(t *testing.T) {
	d := setupTestDB(t)
	h := &Handler{db: d}

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/targets", nil)
	h.listTargets(c)

	require.Equal(t, http.StatusOK, w.Code)
	var resp models.PaginatedResponse[models.TargetSummary]
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Empty(t, resp.Data)
	require.Equal(t, 0, resp.Total)
}

func TestGetTarget_WithLatestSnapshot(t *testing.T) {
	d := setupTestDB(t)
	h := &Handler{db: d}

	targetID := seedTarget(t, d, "example.com")
	snapshotID := seedSnapshotForTarget(t, d, targetID, map[string]any{
		"host": "example.com",
		"asn": map[string]any{
			"asn":     "15169",
			"as_name": "GOOGLE",
			"country": "US",
			"ip":      "8.8.8.8",
		},
		"web": map[string]any{"title": "Example Domain"},
	}, "10.0.0.1")

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/targets/"+targetID.String(), nil)
	c.Params = gin.Params{{Key: "id", Value: targetID.String()}}
	h.getTarget(c)

	require.Equal(t, http.StatusOK, w.Code)
	var detail models.TargetDetail
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &detail))
	require.Equal(t, targetID, detail.ID)
	require.Equal(t, "example.com", detail.Host)
	require.Equal(t, 1, detail.SnapshotCount)
	require.NotNil(t, detail.LatestSnapshot)
	require.Equal(t, snapshotID, detail.LatestSnapshot.ID)
	require.NotNil(t, detail.LatestAsn)
	require.Equal(t, "15169", *detail.LatestAsn)
	require.NotNil(t, detail.LatestWebTitle)
	require.Equal(t, "Example Domain", *detail.LatestWebTitle)
}

func TestGetTarget_NotFound(t *testing.T) {
	d := setupTestDB(t)
	h := &Handler{db: d}

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	missingID := uuid.New()
	c.Request = httptest.NewRequest(http.MethodGet, "/api/targets/"+missingID.String(), nil)
	c.Params = gin.Params{{Key: "id", Value: missingID.String()}}
	h.getTarget(c)

	require.Equal(t, http.StatusNotFound, w.Code)
	require.Contains(t, w.Body.String(), "target not found")
}

func TestGetTarget_NoSnapshots(t *testing.T) {
	d := setupTestDB(t)
	h := &Handler{db: d}

	targetID := seedTarget(t, d, "empty.example.com")

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/targets/"+targetID.String(), nil)
	c.Params = gin.Params{{Key: "id", Value: targetID.String()}}
	h.getTarget(c)

	require.Equal(t, http.StatusOK, w.Code)
	var detail models.TargetDetail
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &detail))
	require.Equal(t, 0, detail.SnapshotCount)
	require.Nil(t, detail.LatestSnapshot)
}

func TestListTargetSnapshots(t *testing.T) {
	d := setupTestDB(t)
	h := &Handler{db: d}

	targetID := seedTarget(t, d, "example.com")
	s1 := seedSnapshotForTarget(t, d, targetID, map[string]any{"host": "example.com", "scanned_at": time.Now().UTC()}, "10.0.0.1")
	s2 := seedSnapshotForTarget(t, d, targetID, map[string]any{"host": "example.com", "scanned_at": time.Now().UTC()}, "10.0.0.2")

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/targets/"+targetID.String()+"/snapshots", nil)
	c.Params = gin.Params{{Key: "id", Value: targetID.String()}}
	h.listTargetSnapshots(c)

	require.Equal(t, http.StatusOK, w.Code)
	var resp models.PaginatedResponse[models.SnapshotSummary]
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Len(t, resp.Data, 2)

	ids := map[uuid.UUID]bool{s1: true, s2: true}
	for _, s := range resp.Data {
		require.True(t, ids[s.ID], "unexpected snapshot id %v", s.ID)
		require.Equal(t, targetID, s.TargetID)
	}
}

func TestListTargetSnapshots_TargetNotFound(t *testing.T) {
	d := setupTestDB(t)
	h := &Handler{db: d}

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	missingID := uuid.New()
	c.Request = httptest.NewRequest(http.MethodGet, "/api/targets/"+missingID.String()+"/snapshots", nil)
	c.Params = gin.Params{{Key: "id", Value: missingID.String()}}
	h.listTargetSnapshots(c)

	require.Equal(t, http.StatusNotFound, w.Code)
	require.Contains(t, w.Body.String(), "target not found")
}

func TestListSnapshots_FilterByTargetID(t *testing.T) {
	d := setupTestDB(t)
	h := &Handler{db: d}

	t1 := seedTarget(t, d, "alpha.example.com")
	t2 := seedTarget(t, d, "beta.example.com")

	seedSnapshotForTarget(t, d, t1, map[string]any{"host": "alpha.example.com"}, "10.0.0.1")
	seedSnapshotForTarget(t, d, t2, map[string]any{"host": "beta.example.com"}, "10.0.0.2")
	seedSnapshotForTarget(t, d, t2, map[string]any{"host": "beta.example.com"}, "10.0.0.3")

	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/snapshots?target_id="+t2.String(), nil)
	h.listSnapshots(c)

	require.Equal(t, http.StatusOK, w.Code)
	var resp models.PaginatedResponse[models.SnapshotSummary]
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Len(t, resp.Data, 2)
	for _, s := range resp.Data {
		require.Equal(t, t2, s.TargetID)
	}
}

func TestListSnapshots_NoFilter(t *testing.T) {
	d := setupTestDB(t)
	h := &Handler{db: d}

	t1 := seedTarget(t, d, "alpha.example.com")
	t2 := seedTarget(t, d, "beta.example.com")
	seedSnapshotForTarget(t, d, t1, map[string]any{"host": "alpha.example.com"}, "")
	seedSnapshotForTarget(t, d, t2, map[string]any{"host": "beta.example.com"}, "")

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/snapshots", nil)
	h.listSnapshots(c)

	require.Equal(t, http.StatusOK, w.Code)
	var resp models.PaginatedResponse[models.SnapshotSummary]
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Len(t, resp.Data, 2)
}

func TestGetSnapshot_IncludesClientIPAndPwhois(t *testing.T) {
	d := setupTestDB(t)
	h := &Handler{db: d}

	targetID := seedTarget(t, d, "example.com")
	snapshotID := seedSnapshotForTarget(t, d, targetID, map[string]any{"host": "example.com"}, "192.168.1.42")

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/snapshots/"+snapshotID.String(), nil)
	c.Params = gin.Params{{Key: "id", Value: snapshotID.String()}}
	h.getSnapshot(c)

	require.Equal(t, http.StatusOK, w.Code)
	var s models.Snapshot
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &s))
	require.Equal(t, snapshotID, s.ID)
	require.Equal(t, targetID, s.TargetID)
	require.Equal(t, "192.168.1.42", s.ClientIP)
	require.NotNil(t, s.PwhoisData)
	require.Equal(t, "15169", s.PwhoisData["asn"])
}

func TestGetSnapshot_NotFound(t *testing.T) {
	d := setupTestDB(t)
	h := &Handler{db: d}

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	missingID := uuid.New()
	c.Request = httptest.NewRequest(http.MethodGet, "/api/snapshots/"+missingID.String(), nil)
	c.Params = gin.Params{{Key: "id", Value: missingID.String()}}
	h.getSnapshot(c)

	require.Equal(t, http.StatusNotFound, w.Code)
	require.Contains(t, w.Body.String(), "snapshot not found")
}

func TestGetSnapshot_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/snapshots/bad-id", nil)
	c.Params = gin.Params{{Key: "id", Value: "bad-id"}}

	h := &Handler{}
	h.getSnapshot(c)

	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Contains(t, w.Body.String(), "invalid id")
}

func TestListReports_FilterByStatus(t *testing.T) {
	d := setupTestDB(t)
	h := &Handler{db: d}

	targetID := seedTarget(t, d, "example.com")
	snapshotID := seedSnapshotForTarget(t, d, targetID, map[string]any{"host": "example.com"}, "")

	insertReport(t, d, snapshotID, models.ReportCompleted, []byte("%PDF completed"))
	insertReport(t, d, snapshotID, models.ReportPending, nil)

	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/reports?status=completed", nil)
	h.listReports(c)

	require.Equal(t, http.StatusOK, w.Code)
	var resp models.PaginatedResponse[models.ReportSummary]
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Len(t, resp.Data, 1)
	require.Equal(t, models.ReportCompleted, resp.Data[0].Status)
}

func TestListReports_NoFilter(t *testing.T) {
	d := setupTestDB(t)
	h := &Handler{db: d}

	targetID := seedTarget(t, d, "example.com")
	snapshotID := seedSnapshotForTarget(t, d, targetID, map[string]any{"host": "example.com"}, "")

	insertReport(t, d, snapshotID, models.ReportCompleted, []byte("%PDF"))
	insertReport(t, d, snapshotID, models.ReportPending, nil)

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/reports", nil)
	h.listReports(c)

	require.Equal(t, http.StatusOK, w.Code)
	var resp models.PaginatedResponse[models.ReportSummary]
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Len(t, resp.Data, 2)
}

func TestDeleteSnapshot_RemovesRow(t *testing.T) {
	d := setupTestDB(t)
	targetID := seedTarget(t, d, "example.com")
	snapshotID := seedSnapshotForTarget(t, d, targetID, map[string]any{"host": "example.com"}, "")

	h := &Handler{db: d}
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/snapshots/"+snapshotID.String(), nil)
	c.Params = gin.Params{{Key: "id", Value: snapshotID.String()}}
	h.deleteSnapshot(c)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), `"deleted":true`)

	var exists bool
	err := d.Pool.QueryRow(context.Background(), `
		SELECT EXISTS(SELECT 1 FROM snapshots WHERE id = $1)
	`, snapshotID).Scan(&exists)
	require.NoError(t, err)
	require.False(t, exists)
}

func TestDeleteReport_RemovesRow(t *testing.T) {
	d := setupTestDB(t)
	h := newTestHandler(t, d)
	targetID := seedTarget(t, d, "example.com")
	snapshotID := seedSnapshotForTarget(t, d, targetID, map[string]any{"host": "example.com"}, "")
	reportID := insertReport(t, d, snapshotID, models.ReportCompleted, []byte("%PDF"))

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/reports/"+reportID.String(), nil)
	c.Params = gin.Params{{Key: "id", Value: reportID.String()}}
	h.deleteReport(c)

	require.Equal(t, http.StatusOK, w.Code)

	var exists bool
	err := d.Pool.QueryRow(context.Background(), `
		SELECT EXISTS(SELECT 1 FROM reports WHERE id = $1)
	`, reportID).Scan(&exists)
	require.NoError(t, err)
	require.False(t, exists)
}

func TestListTargets_InvalidTargetID(t *testing.T) {
	d := setupTestDB(t)
	h := &Handler{db: d}

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/targets/bad-id/snapshots", nil)
	c.Params = gin.Params{{Key: "id", Value: "bad-id"}}
	h.listTargetSnapshots(c)

	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Contains(t, w.Body.String(), "invalid id")
}
