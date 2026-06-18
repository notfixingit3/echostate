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

func TestCreateUpdateDeleteGraphView(t *testing.T) {
	d := setupTestDB(t)
	h := &Handler{db: d}
	targetID := seedTarget(t, d, "graph-view.example.com")

	gin.SetMode(gin.TestMode)

	createBody := models.CreateSavedGraphViewRequest{
		Name:        "BGP drift lens",
		Description: "Compare latest BGP for target",
		ViewMode:    "bgp",
		TargetID:    &targetID,
		VantageFilter: "all",
		CompareMode: "latest",
		PinnedNodes: map[string]models.GraphPinnedNode{
			"node-1": {X: 12.5, Y: -4.2},
		},
		SelectedNodeID: strPtr("node-1"),
	}
	payload, err := json.Marshal(createBody)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/graph/views", bytes.NewReader(payload))
	c.Request.Header.Set("Content-Type", "application/json")
	h.createGraphView(c)
	require.Equal(t, http.StatusCreated, w.Code)

	var created models.SavedGraphView
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))
	require.Equal(t, "bgp", created.ViewMode)
	require.Equal(t, targetID, *created.TargetID)
	require.InDelta(t, 12.5, created.PinnedNodes["node-1"].X, 0.001)

	updateBody := models.UpdateSavedGraphViewRequest{
		Name:        "BGP drift lens",
		Description: "Updated description",
		ViewMode:    "traceroute",
		TargetID:    &targetID,
		VantageFilter: "external",
		CompareMode: "previous",
		PinnedNodes: map[string]models.GraphPinnedNode{},
	}
	updatePayload, err := json.Marshal(updateBody)
	require.NoError(t, err)

	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Request = httptest.NewRequest(http.MethodPut, "/api/graph/views/"+created.ID.String(), bytes.NewReader(updatePayload))
	c2.Request.Header.Set("Content-Type", "application/json")
	c2.Params = gin.Params{{Key: "id", Value: created.ID.String()}}
	h.updateGraphView(c2)
	require.Equal(t, http.StatusOK, w2.Code)

	w3 := httptest.NewRecorder()
	c3, _ := gin.CreateTestContext(w3)
	c3.Request = httptest.NewRequest(http.MethodGet, "/api/graph/views", nil)
	h.listGraphViews(c3)
	require.Equal(t, http.StatusOK, w3.Code)

	var listed []models.SavedGraphView
	require.NoError(t, json.Unmarshal(w3.Body.Bytes(), &listed))
	require.Len(t, listed, 1)

	w4 := httptest.NewRecorder()
	c4, _ := gin.CreateTestContext(w4)
	c4.Request = httptest.NewRequest(http.MethodDelete, "/api/graph/views/"+created.ID.String(), nil)
	c4.Params = gin.Params{{Key: "id", Value: created.ID.String()}}
	h.deleteGraphView(c4)
	require.Equal(t, http.StatusOK, w4.Code)
}

func strPtr(value string) *string {
	return &value
}