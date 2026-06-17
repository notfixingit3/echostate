package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/notfixingit3/echostate/internal/config"
)

func (h *Handler) getSettings(c *gin.Context) {
	var valBytes []byte
	err := h.db.Pool.QueryRow(c.Request.Context(), `SELECT value FROM settings WHERE key = 'app_settings'`).Scan(&valBytes)
	
	settings := config.GetSettings()
	if err == nil {
		json.Unmarshal(valBytes, &settings)
		config.UpdateSettings(settings)
	}

	c.JSON(http.StatusOK, settings)
}

func (h *Handler) updateSettings(c *gin.Context) {
	var req config.SystemSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	valBytes, _ := json.Marshal(req)
	_, err := h.db.Pool.Exec(c.Request.Context(), `
		INSERT INTO settings (key, value) VALUES ('app_settings', $1)
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()
	`, valBytes)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save settings"})
		return
	}

	config.UpdateSettings(req)
	c.JSON(http.StatusOK, req)
}
