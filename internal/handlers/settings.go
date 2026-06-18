package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/notfixingit3/echostate/internal/config"
)

func (h *Handler) getSettings(c *gin.Context) {
	var valBytes []byte
	err := h.db.Pool.QueryRow(c.Request.Context(), `SELECT value FROM settings WHERE key = 'app_settings'`).Scan(&valBytes)

	settings := config.GetSettings()
	if err == nil {
		_ = json.Unmarshal(valBytes, &settings)
		config.UpdateSettings(settings)
	}

	c.JSON(http.StatusOK, config.PublicSettings(settings))
}

func (h *Handler) updateSettings(c *gin.Context) {
	var req config.SystemSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	current := config.GetSettings()
	req.APIKey = mergeSecret(req.APIKey, current.APIKey)
	req.ShodanAPIKey = mergeSecret(req.ShodanAPIKey, current.ShodanAPIKey)
	req.CensysAPISecret = mergeSecret(req.CensysAPISecret, current.CensysAPISecret)
	config.UpdateSettings(req)
	req = config.GetSettings()

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
	c.JSON(http.StatusOK, config.PublicSettings(req))
}

func mergeSecret(incoming, current string) string {
	if strings.Contains(incoming, "****") {
		return current
	}
	return strings.TrimSpace(incoming)
}