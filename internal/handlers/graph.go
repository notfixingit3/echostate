package handlers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/notfixingit3/echostate/internal/db"
	"github.com/notfixingit3/echostate/internal/observability/log"
)

func (h *Handler) getGraph(c *gin.Context) {
	if h.neo4jClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "graph database unavailable"})
		return
	}

	opts := db.GraphQueryOptsFromRequest(
		c.Query("target_id"),
		c.Query("view"),
		c.Query("snapshot_id"),
		c.Query("compare_snapshot_id"),
		c.Query("compare_mode"),
	)

	graph, err := h.neo4jClient.GetGraphWithOpts(c.Request.Context(), opts)
	if err != nil {
		log.FromContext(c.Request.Context()).Error("graph query failed",
			slog.String("action", "get_graph"),
			slog.String("target_id", opts.TargetID),
			slog.String("snapshot_id", opts.SnapshotID),
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load graph"})
		return
	}

	if opts.TargetID != "" && (opts.View == "bgp" || opts.View == "traceroute" || opts.View == "") {
		if id, err := uuid.Parse(opts.TargetID); err == nil {
			if diff, err := h.db.GetRouteDiffForSnapshots(
				c.Request.Context(),
				id,
				opts.SnapshotID,
				opts.CompareSnapshotID,
			); err == nil {
				graph.RouteDiff = diff
			}
		}
	}

	c.JSON(http.StatusOK, graph)
}
