package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/notfixingit3/echostate/internal/audit"
	"github.com/notfixingit3/echostate/internal/export"
	"github.com/notfixingit3/echostate/internal/models"
	"github.com/notfixingit3/echostate/internal/version"
)

func (h *Handler) exportData(c *gin.Context) {
	includeBlobs := parseBoolQuery(c, "include_blobs", true)
	includeConfig := parseBoolQuery(c, "include_config", false)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Minute)
	defer cancel()

	bundle, err := export.BuildBundle(ctx, h.db, export.Options{
		IncludeBlobs:  includeBlobs,
		IncludeConfig: includeConfig,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("export failed: %v", err)})
		return
	}

	filename := fmt.Sprintf("echostate-export-%s.json", time.Now().UTC().Format("20060102-150405"))
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Header("X-EchoState-Version", version.Version)
	h.recordAudit(c, audit.ActionDataExport, "export", filename, map[string]any{
		"include_blobs":  includeBlobs,
		"include_config": includeConfig,
	})
	c.JSON(http.StatusOK, bundle)
}

func (h *Handler) importData(c *gin.Context) {
	var req models.ImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Bundle.FormatVersion == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bundle is required"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Minute)
	defer cancel()

	var neo4j export.GraphSyncer
	if h.neo4jClient != nil {
		neo4j = h.neo4jClient
	}

	result, err := export.ImportBundle(ctx, h.db, neo4j, &req.Bundle, export.ImportOptions{
		Conflict:     req.Conflict,
		RebuildGraph: req.RebuildGraph,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.recordAudit(c, audit.ActionDataImport, "import", "bundle", map[string]any{
		"conflict":      req.Conflict,
		"rebuild_graph": req.RebuildGraph,
		"targets_imported":   result.TargetsImported,
		"snapshots_imported": result.SnapshotsImported,
		"targets_skipped":    result.TargetsSkipped,
		"snapshots_skipped":  result.SnapshotsSkipped,
	})
	c.JSON(http.StatusOK, result)
}

func parseBoolQuery(c *gin.Context, key string, defaultValue bool) bool {
	raw := c.Query(key)
	if raw == "" {
		return defaultValue
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return defaultValue
	}
	return value
}

