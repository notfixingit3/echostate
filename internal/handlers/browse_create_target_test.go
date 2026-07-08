package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/notfixingit3/echostate/internal/models"
)

func TestCreateTarget(t *testing.T) {
	gin.SetMode(gin.TestMode)
	d := setupTestDB(t)
	h := &Handler{db: d}

	body, err := json.Marshal(map[string]string{"host": "new-target.example.com"})
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/targets", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.createTarget(c)
	require.Equal(t, http.StatusCreated, w.Code)

	var summary models.TargetSummary
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &summary))
	require.Equal(t, "new-target.example.com", summary.Host)
	require.Equal(t, 0, summary.SnapshotCount)

	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Request = httptest.NewRequest(http.MethodPost, "/api/targets", bytes.NewReader(body))
	c2.Request.Header.Set("Content-Type", "application/json")

	h.createTarget(c2)
	require.Equal(t, http.StatusOK, w2.Code)
}

func TestCreateTarget_InvalidHost(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &Handler{db: setupTestDB(t)}

	body, err := json.Marshal(map[string]string{"host": "notavalidhost"})
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/targets", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.createTarget(c)
	require.Equal(t, http.StatusBadRequest, w.Code)
}
