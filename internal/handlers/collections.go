package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/notfixingit3/echostate/internal/models"
)

func (h *Handler) listCollections(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	rows, err := h.db.Pool.Query(ctx, `
		SELECT c.id, c.name, c.description, c.created_at, c.updated_at,
		       COUNT(m.target_id) AS target_count
		FROM target_collections c
		LEFT JOIN target_collection_members m ON m.collection_id = c.id
		GROUP BY c.id
		ORDER BY c.updated_at DESC, c.name ASC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list collections"})
		return
	}
	defer rows.Close()

	var collections []models.TargetCollectionSummary
	for rows.Next() {
		var item models.TargetCollectionSummary
		if err := rows.Scan(
			&item.ID, &item.Name, &item.Description, &item.CreatedAt, &item.UpdatedAt, &item.TargetCount,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read collection"})
			return
		}
		collections = append(collections, item)
	}
	if collections == nil {
		collections = []models.TargetCollectionSummary{}
	}

	c.JSON(http.StatusOK, collections)
}

func (h *Handler) createCollection(c *gin.Context) {
	var req models.CreateTargetCollectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var collection models.TargetCollectionDetail
	err := h.db.Pool.QueryRow(ctx, `
		INSERT INTO target_collections (name, description)
		VALUES ($1, $2)
		RETURNING id, name, description, created_at, updated_at
	`, name, strings.TrimSpace(req.Description)).Scan(
		&collection.ID, &collection.Name, &collection.Description, &collection.CreatedAt, &collection.UpdatedAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create collection"})
		return
	}
	collection.Targets = []models.TargetSummary{}

	c.JSON(http.StatusCreated, collection)
}

func (h *Handler) getCollection(c *gin.Context) {
	collectionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid collection id"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	detail, err := h.loadCollectionDetail(ctx, collectionID)
	if err == pgx.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "collection not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load collection"})
		return
	}

	c.JSON(http.StatusOK, detail)
}

func (h *Handler) updateCollection(c *gin.Context) {
	collectionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid collection id"})
		return
	}

	var req models.UpdateTargetCollectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	tag, err := h.db.Pool.Exec(ctx, `
		UPDATE target_collections
		SET name = $1, description = $2, updated_at = NOW()
		WHERE id = $3
	`, name, strings.TrimSpace(req.Description), collectionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update collection"})
		return
	}
	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "collection not found"})
		return
	}

	detail, err := h.loadCollectionDetail(ctx, collectionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load collection"})
		return
	}

	c.JSON(http.StatusOK, detail)
}

func (h *Handler) deleteCollection(c *gin.Context) {
	collectionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid collection id"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	tag, err := h.db.Pool.Exec(ctx, `DELETE FROM target_collections WHERE id = $1`, collectionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete collection"})
		return
	}
	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "collection not found"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) addCollectionTarget(c *gin.Context) {
	collectionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid collection id"})
		return
	}

	var req models.CollectionTargetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var exists bool
	if err := h.db.Pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM targets WHERE id = $1)`, req.TargetID).Scan(&exists); err != nil || !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "target not found"})
		return
	}

	_, err = h.db.Pool.Exec(ctx, `
		INSERT INTO target_collection_members (collection_id, target_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, collectionID, req.TargetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add target"})
		return
	}

	_, _ = h.db.Pool.Exec(ctx, `UPDATE target_collections SET updated_at = NOW() WHERE id = $1`, collectionID)

	detail, err := h.loadCollectionDetail(ctx, collectionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load collection"})
		return
	}

	c.JSON(http.StatusOK, detail)
}

func (h *Handler) removeCollectionTarget(c *gin.Context) {
	collectionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid collection id"})
		return
	}
	targetID, err := uuid.Parse(c.Param("targetId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid target id"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	tag, err := h.db.Pool.Exec(ctx, `
		DELETE FROM target_collection_members
		WHERE collection_id = $1 AND target_id = $2
	`, collectionID, targetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove target"})
		return
	}
	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "membership not found"})
		return
	}

	_, _ = h.db.Pool.Exec(ctx, `UPDATE target_collections SET updated_at = NOW() WHERE id = $1`, collectionID)
	c.Status(http.StatusNoContent)
}

func (h *Handler) rescanCollection(c *gin.Context) {
	collectionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid collection id"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	rows, err := h.db.Pool.Query(ctx, `
		SELECT t.id, t.host
		FROM target_collection_members m
		JOIN targets t ON t.id = m.target_id
		WHERE m.collection_id = $1
		ORDER BY t.host ASC
	`, collectionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load collection targets"})
		return
	}
	defer rows.Close()

	var jobs []models.CollectionRescanJob
	for rows.Next() {
		var targetID uuid.UUID
		var host string
		if err := rows.Scan(&targetID, &host); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read target"})
			return
		}

		jobID, err := h.scanWorker.CreateJob(ctx, host, c.ClientIP())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to enqueue rescan"})
			return
		}
		jobs = append(jobs, models.CollectionRescanJob{
			TargetID: targetID,
			Host:     host,
			JobID:    jobID,
		})
	}
	if jobs == nil {
		jobs = []models.CollectionRescanJob{}
	}

	c.JSON(http.StatusAccepted, models.CollectionRescanResponse{
		CollectionID: collectionID,
		Jobs:         jobs,
	})
}

func (h *Handler) loadCollectionDetail(ctx context.Context, collectionID uuid.UUID) (*models.TargetCollectionDetail, error) {
	var detail models.TargetCollectionDetail
	err := h.db.Pool.QueryRow(ctx, `
		SELECT id, name, description, created_at, updated_at
		FROM target_collections
		WHERE id = $1
	`, collectionID).Scan(
		&detail.ID, &detail.Name, &detail.Description, &detail.CreatedAt, &detail.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	rows, err := h.db.Pool.Query(ctx, `
		SELECT t.id, t.host, t.tags, t.created_at,
		       COUNT(s.id) AS snapshot_count,
		       MAX(s.scanned_at) AS latest_snapshot_at
		FROM target_collection_members m
		JOIN targets t ON t.id = m.target_id
		LEFT JOIN snapshots s ON s.target_id = t.id
		WHERE m.collection_id = $1
		GROUP BY t.id
		ORDER BY t.host ASC
	`, collectionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var target models.TargetSummary
		if err := rows.Scan(
			&target.ID, &target.Host, &target.Tags, &target.CreatedAt,
			&target.SnapshotCount, &target.LatestSnapshotAt,
		); err != nil {
			return nil, err
		}
		detail.Targets = append(detail.Targets, target)
	}
	if detail.Targets == nil {
		detail.Targets = []models.TargetSummary{}
	}
	detail.TargetCount = len(detail.Targets)

	return &detail, nil
}
