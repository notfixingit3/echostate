package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/notfixingit3/echostate/internal/models"
)

func (h *Handler) listWebhooks(c *gin.Context) {
	rows, err := h.db.Pool.Query(c.Request.Context(), `
		SELECT id, name, type, url, enabled, created_at, updated_at
		FROM webhooks
		ORDER BY created_at DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var webhooks []models.Webhook
	for rows.Next() {
		var w models.Webhook
		if err := rows.Scan(&w.ID, &w.Name, &w.Type, &w.URL, &w.Enabled, &w.CreatedAt, &w.UpdatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		webhooks = append(webhooks, w)
	}

	if webhooks == nil {
		webhooks = []models.Webhook{}
	}

	c.JSON(http.StatusOK, webhooks)
}

func (h *Handler) createWebhook(c *gin.Context) {
	var req models.Webhook
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.ID = uuid.New()
	req.CreatedAt = time.Now().UTC()
	req.UpdatedAt = req.CreatedAt

	_, err := h.db.Pool.Exec(c.Request.Context(), `
		INSERT INTO webhooks (id, name, type, url, enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, req.ID, req.Name, req.Type, req.URL, req.Enabled, req.CreatedAt, req.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, req)
}

func (h *Handler) updateWebhook(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid webhook id"})
		return
	}

	var req models.Webhook
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now().UTC()
	_, err = h.db.Pool.Exec(c.Request.Context(), `
		UPDATE webhooks
		SET name = $1, type = $2, url = $3, enabled = $4, updated_at = $5
		WHERE id = $6
	`, req.Name, req.Type, req.URL, req.Enabled, now, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req.ID = id
	req.UpdatedAt = now
	c.JSON(http.StatusOK, req)
}

func (h *Handler) deleteWebhook(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid webhook id"})
		return
	}

	_, err = h.db.Pool.Exec(c.Request.Context(), `DELETE FROM webhooks WHERE id = $1`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
