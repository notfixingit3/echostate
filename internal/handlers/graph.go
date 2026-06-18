package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *Handler) getGraph(c *gin.Context) {
	if h.neo4jClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "graph database unavailable"})
		return
	}

	targetID := c.Query("target_id")
	view := c.Query("view")
	graph, err := h.neo4jClient.GetGraph(c.Request.Context(), targetID, view)
	if err != nil {
		log.Printf("graph query failed (target_id=%q): %v", targetID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load graph"})
		return
	}

	if targetID != "" && (view == "bgp" || view == "traceroute") {
		if id, err := uuid.Parse(targetID); err == nil {
			if diff, err := h.db.GetRouteDiffForTarget(c.Request.Context(), id); err == nil {
				graph.RouteDiff = diff
			}
		}
	}

	c.JSON(http.StatusOK, graph)
}