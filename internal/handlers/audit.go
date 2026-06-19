package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/notfixingit3/echostate/internal/audit"
)

func (h *Handler) listAuditEvents(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page"})
		return
	}
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if err != nil || limit < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit"})
		return
	}

	result, err := h.audit.List(c.Request.Context(), audit.ListOptions{
		Page:   page,
		Limit:  limit,
		Action: c.Query("action"),
		Query:  c.Query("q"),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list audit events"})
		return
	}

	c.JSON(http.StatusOK, result)
}