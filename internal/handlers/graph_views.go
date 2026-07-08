package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/notfixingit3/echostate/internal/models"
)

var validGraphViewModes = map[string]struct{}{
	"infra": {}, "bgp": {}, "traceroute": {}, "peering": {},
	"ct": {}, "dns": {}, "cert": {},
}

var validGraphCompareModes = map[string]struct{}{
	"previous": {}, "latest": {},
}

func (h *Handler) listGraphViews(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	rows, err := h.db.Pool.Query(ctx, `
		SELECT v.id, v.name, v.description, v.view_mode, v.target_id, v.vantage_filter,
			v.snapshot_id, v.compare_snapshot_id, v.compare_mode, v.pinned_nodes,
			v.selected_node_id, v.created_at, v.updated_at, t.host
		FROM graph_views v
		LEFT JOIN targets t ON t.id = v.target_id
		ORDER BY v.updated_at DESC, v.name ASC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list graph views"})
		return
	}
	defer rows.Close()

	views, err := scanGraphViewRows(rows)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read graph views"})
		return
	}
	if views == nil {
		views = []models.SavedGraphView{}
	}

	c.JSON(http.StatusOK, views)
}

func (h *Handler) getGraphView(c *gin.Context) {
	viewID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid graph view id"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	view, err := h.loadGraphView(ctx, viewID)
	if err == pgx.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "graph view not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load graph view"})
		return
	}

	c.JSON(http.StatusOK, view)
}

func (h *Handler) createGraphView(c *gin.Context) {
	var req models.CreateSavedGraphViewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payload, err := normalizeGraphViewPayload(req.Name, req.Description, req.ViewMode, req.TargetID, req.VantageFilter, req.SnapshotID, req.CompareSnapshotID, req.CompareMode, req.PinnedNodes, req.SelectedNodeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var viewID uuid.UUID
	err = h.db.Pool.QueryRow(ctx, `
		INSERT INTO graph_views (
			name, description, view_mode, target_id, vantage_filter,
			snapshot_id, compare_snapshot_id, compare_mode,
			pinned_nodes, selected_node_id
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`,
		payload.name,
		payload.description,
		payload.viewMode,
		payload.targetID,
		payload.vantageFilter,
		payload.snapshotID,
		payload.compareSnapshotID,
		payload.compareMode,
		payload.pinnedJSON,
		payload.selectedNodeID,
	).Scan(&viewID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create graph view"})
		return
	}

	view, err := h.loadGraphView(ctx, viewID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load graph view"})
		return
	}

	c.JSON(http.StatusCreated, view)
}

func (h *Handler) updateGraphView(c *gin.Context) {
	viewID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid graph view id"})
		return
	}

	var req models.UpdateSavedGraphViewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payload, err := normalizeGraphViewPayload(req.Name, req.Description, req.ViewMode, req.TargetID, req.VantageFilter, req.SnapshotID, req.CompareSnapshotID, req.CompareMode, req.PinnedNodes, req.SelectedNodeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	tag, err := h.db.Pool.Exec(ctx, `
		UPDATE graph_views
		SET name = $1,
			description = $2,
			view_mode = $3,
			target_id = $4,
			vantage_filter = $5,
			snapshot_id = $6,
			compare_snapshot_id = $7,
			compare_mode = $8,
			pinned_nodes = $9,
			selected_node_id = $10,
			updated_at = NOW()
		WHERE id = $11
	`,
		payload.name,
		payload.description,
		payload.viewMode,
		payload.targetID,
		payload.vantageFilter,
		payload.snapshotID,
		payload.compareSnapshotID,
		payload.compareMode,
		payload.pinnedJSON,
		payload.selectedNodeID,
		viewID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update graph view"})
		return
	}
	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "graph view not found"})
		return
	}

	view, err := h.loadGraphView(ctx, viewID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load graph view"})
		return
	}

	c.JSON(http.StatusOK, view)
}

func (h *Handler) deleteGraphView(c *gin.Context) {
	viewID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid graph view id"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	tag, err := h.db.Pool.Exec(ctx, `DELETE FROM graph_views WHERE id = $1`, viewID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete graph view"})
		return
	}
	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "graph view not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

type graphViewPayload struct {
	name              string
	description       string
	viewMode          string
	targetID          *uuid.UUID
	vantageFilter     string
	snapshotID        *uuid.UUID
	compareSnapshotID *uuid.UUID
	compareMode       string
	pinnedJSON        []byte
	selectedNodeID    *string
}

func normalizeGraphViewPayload(
	name, description, viewMode string,
	targetID *uuid.UUID,
	vantageFilter string,
	snapshotID, compareSnapshotID *uuid.UUID,
	compareMode string,
	pinned map[string]models.GraphPinnedNode,
	selectedNodeID *string,
) (graphViewPayload, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return graphViewPayload{}, errInvalidGraphView
	}

	viewMode = strings.TrimSpace(strings.ToLower(viewMode))
	if viewMode == "" {
		viewMode = "infra"
	}
	if _, ok := validGraphViewModes[viewMode]; !ok {
		return graphViewPayload{}, errInvalidGraphView
	}

	vantageFilter = strings.TrimSpace(vantageFilter)
	if vantageFilter == "" {
		vantageFilter = "all"
	}

	compareMode = strings.TrimSpace(strings.ToLower(compareMode))
	if compareMode == "" {
		compareMode = "previous"
	}
	if _, ok := validGraphCompareModes[compareMode]; !ok {
		return graphViewPayload{}, errInvalidGraphView
	}

	if pinned == nil {
		pinned = map[string]models.GraphPinnedNode{}
	}
	pinnedJSON, err := json.Marshal(pinned)
	if err != nil {
		return graphViewPayload{}, err
	}

	return graphViewPayload{
		name:              name,
		description:       strings.TrimSpace(description),
		viewMode:          viewMode,
		targetID:          targetID,
		vantageFilter:     vantageFilter,
		snapshotID:        snapshotID,
		compareSnapshotID: compareSnapshotID,
		compareMode:       compareMode,
		pinnedJSON:        pinnedJSON,
		selectedNodeID:    trimOptionalString(selectedNodeID),
	}, nil
}

var errInvalidGraphView = &graphViewError{message: "invalid graph view payload"}

type graphViewError struct {
	message string
}

func (e *graphViewError) Error() string {
	return e.message
}

func (h *Handler) loadGraphView(ctx context.Context, viewID uuid.UUID) (*models.SavedGraphView, error) {
	row := h.db.Pool.QueryRow(ctx, `
		SELECT v.id, v.name, v.description, v.view_mode, v.target_id, v.vantage_filter,
			v.snapshot_id, v.compare_snapshot_id, v.compare_mode, v.pinned_nodes,
			v.selected_node_id, v.created_at, v.updated_at, t.host
		FROM graph_views v
		LEFT JOIN targets t ON t.id = v.target_id
		WHERE v.id = $1
	`, viewID)

	var view models.SavedGraphView
	var pinnedJSON []byte
	var targetHost *string
	err := row.Scan(
		&view.ID,
		&view.Name,
		&view.Description,
		&view.ViewMode,
		&view.TargetID,
		&view.VantageFilter,
		&view.SnapshotID,
		&view.CompareSnapshotID,
		&view.CompareMode,
		&pinnedJSON,
		&view.SelectedNodeID,
		&view.CreatedAt,
		&view.UpdatedAt,
		&targetHost,
	)
	if err != nil {
		return nil, err
	}

	view.PinnedNodes = map[string]models.GraphPinnedNode{}
	if err := json.Unmarshal(pinnedJSON, &view.PinnedNodes); err != nil {
		return nil, err
	}
	if view.PinnedNodes == nil {
		view.PinnedNodes = map[string]models.GraphPinnedNode{}
	}
	view.TargetHost = targetHost
	return &view, nil
}

func scanGraphViewRows(rows pgx.Rows) ([]models.SavedGraphView, error) {
	var views []models.SavedGraphView
	for rows.Next() {
		var view models.SavedGraphView
		var pinnedJSON []byte
		var targetHost *string
		if err := rows.Scan(
			&view.ID,
			&view.Name,
			&view.Description,
			&view.ViewMode,
			&view.TargetID,
			&view.VantageFilter,
			&view.SnapshotID,
			&view.CompareSnapshotID,
			&view.CompareMode,
			&pinnedJSON,
			&view.SelectedNodeID,
			&view.CreatedAt,
			&view.UpdatedAt,
			&targetHost,
		); err != nil {
			return nil, err
		}
		view.PinnedNodes = map[string]models.GraphPinnedNode{}
		if err := json.Unmarshal(pinnedJSON, &view.PinnedNodes); err != nil {
			return nil, err
		}
		if view.PinnedNodes == nil {
			view.PinnedNodes = map[string]models.GraphPinnedNode{}
		}
		view.TargetHost = targetHost
		views = append(views, view)
	}
	return views, rows.Err()
}
