package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/notfixingit3/echostate/internal/models"
	"github.com/notfixingit3/echostate/internal/webhooks"
)

func scanWebhook(row interface {
	Scan(dest ...any) error
}) (models.Webhook, error) {
	var w models.Webhook
	var configBytes []byte
	if err := row.Scan(&w.ID, &w.Name, &w.Type, &w.URL, &configBytes, &w.Enabled, &w.CreatedAt, &w.UpdatedAt); err != nil {
		return w, err
	}
	if len(configBytes) > 0 {
		_ = json.Unmarshal(configBytes, &w.Config)
	}
	if w.Config == nil {
		w.Config = map[string]any{}
	}
	return w, nil
}

func normalizeWebhook(req *models.Webhook) error {
	req.Name = strings.TrimSpace(req.Name)
	req.Type = strings.TrimSpace(strings.ToLower(req.Type))
	req.URL = strings.TrimSpace(req.URL)

	switch req.Type {
	case "slack", "discord", "teams":
		if req.URL == "" {
			return errBadRequest("webhook url is required")
		}
		req.Config = map[string]any{}
	case "pushover":
		req.URL = "https://api.pushover.net/1/messages.json"
		if _, err := webhooks.ParsePushoverConfig(req.Config); err != nil {
			return errBadRequest(err.Error())
		}
	default:
		return errBadRequest("unsupported webhook type")
	}

	return nil
}

type badRequestError struct {
	msg string
}

func (e badRequestError) Error() string { return e.msg }

func errBadRequest(msg string) error { return badRequestError{msg: msg} }

func (h *Handler) listWebhooks(c *gin.Context) {
	rows, err := h.db.Pool.Query(c.Request.Context(), `
		SELECT id, name, type, url, config, enabled, created_at, updated_at
		FROM webhooks
		ORDER BY created_at DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var hooks []models.Webhook
	for rows.Next() {
		w, err := scanWebhook(rows)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		hooks = append(hooks, w)
	}

	if hooks == nil {
		hooks = []models.Webhook{}
	}

	c.JSON(http.StatusOK, hooks)
}

func (h *Handler) createWebhook(c *gin.Context) {
	var req models.Webhook
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := normalizeWebhook(&req); err != nil {
		if br, ok := err.(badRequestError); ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": br.msg})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.ID = uuid.New()
	req.CreatedAt = time.Now().UTC()
	req.UpdatedAt = req.CreatedAt

	configBytes, _ := json.Marshal(req.Config)

	_, err := h.db.Pool.Exec(c.Request.Context(), `
		INSERT INTO webhooks (id, name, type, url, config, enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, req.ID, req.Name, req.Type, req.URL, configBytes, req.Enabled, req.CreatedAt, req.UpdatedAt)
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

	if err := normalizeWebhook(&req); err != nil {
		if br, ok := err.(badRequestError); ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": br.msg})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now().UTC()
	configBytes, _ := json.Marshal(req.Config)

	_, err = h.db.Pool.Exec(c.Request.Context(), `
		UPDATE webhooks
		SET name = $1, type = $2, url = $3, config = $4, enabled = $5, updated_at = $6
		WHERE id = $7
	`, req.Name, req.Type, req.URL, configBytes, req.Enabled, now, id)
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