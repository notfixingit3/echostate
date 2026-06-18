package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/notfixingit3/echostate/internal/config"
	"github.com/notfixingit3/echostate/internal/db"
	"github.com/notfixingit3/echostate/internal/middleware"
	"github.com/notfixingit3/echostate/internal/models"
	"github.com/notfixingit3/echostate/internal/reports"
	"github.com/notfixingit3/echostate/internal/scheduler"
	"github.com/notfixingit3/echostate/internal/scanner"
	"github.com/notfixingit3/echostate/internal/scans"
	"github.com/notfixingit3/echostate/internal/version"
)

// scanRunner is the subset of *scanner.Scanner used by the handlers, extracted
// as an interface so tests can inject mock scanners without a live browser.
type scanRunner interface {
	Run(ctx context.Context, host string) (*models.ScanResult, error)
}

// Handler carries dependencies for HTTP handlers.
type Handler struct {
	db          *db.DB
	neo4jClient *db.Neo4jClient
	config      *config.Config
	scanner     scanRunner
	reportWorker *reports.Worker
	scanWorker   *scans.Worker
}

// Register wires routes into the Gin router and returns background workers.
func Register(router *gin.Engine, database *db.DB, neo4jClient *db.Neo4jClient, cfg *config.Config, rateLimiter *middleware.RateLimiter) *Workers {
	h := &Handler{
		db:          database,
		neo4jClient: neo4jClient,
		config:      cfg,
		scanner:     scanner.NewScanner(cfg.BrowserWSURL),
		reportWorker: newWorker(database),
	}

	var valBytes []byte
	err := database.Pool.QueryRow(context.Background(), `SELECT value FROM settings WHERE key = 'app_settings'`).Scan(&valBytes)
	if err == nil {
		var s config.SystemSettings
		if err := json.Unmarshal(valBytes, &s); err == nil {
			config.UpdateSettings(s)
		}
	}

	h.scanWorker = scans.New(database, h.scanner, h, config.GetSettings().ScanConcurrency)

	auth := middleware.APIKeyAuth()

	router.GET("/health", h.health)
	router.GET("/api/version", h.getVersion)
	router.POST("/api/scan", auth, rateLimiter.Middleware(), h.createScan)
	router.GET("/api/scans/:id", h.getScan)

	router.GET("/api/reports", h.listReports)
	router.POST("/api/reports", auth, h.createReport)
	router.GET("/api/reports/:id", h.getReport)
	router.GET("/api/reports/:id/download", h.downloadReport)

	router.GET("/api/targets", h.listTargets)
	router.GET("/api/targets/:id", h.getTarget)
	router.PUT("/api/targets/:id/tags", auth, h.updateTargetTags)
	router.GET("/api/targets/:id/snapshots", h.listTargetSnapshots)
	router.GET("/api/targets/:id/screenshots", h.listTargetScreenshots)

	router.GET("/api/snapshots", h.listSnapshots)
	router.GET("/api/snapshots/:id", h.getSnapshot)
	router.GET("/api/snapshots/:id/diff", h.getSnapshotDiff)
	router.GET("/api/snapshots/:id/report", h.snapshotReport)

	router.GET("/api/webhooks", h.listWebhooks)
	router.POST("/api/webhooks", auth, h.createWebhook)
	router.PUT("/api/webhooks/:id", auth, h.updateWebhook)
	router.DELETE("/api/webhooks/:id", auth, h.deleteWebhook)

	router.GET("/api/settings", h.getSettings)
	router.PUT("/api/settings", auth, h.updateSettings)

	router.GET("/api/graph", h.getGraph)

	router.GET("/api/collections", h.listCollections)
	router.POST("/api/collections", auth, h.createCollection)
	router.GET("/api/collections/:id", h.getCollection)
	router.PUT("/api/collections/:id", auth, h.updateCollection)
	router.DELETE("/api/collections/:id", auth, h.deleteCollection)
	router.POST("/api/collections/:id/targets", auth, h.addCollectionTarget)
	router.DELETE("/api/collections/:id/targets/:targetId", auth, h.removeCollectionTarget)
	router.POST("/api/collections/:id/rescan", auth, h.rescanCollection)

	router.GET("/api/notes", h.listNotes)
	router.GET("/api/notes/:id", h.getNote)
	router.POST("/api/notes", auth, h.createNote)
	router.PUT("/api/notes/:id", auth, h.updateNote)
	router.POST("/api/notes/:id/archive", auth, h.archiveNote)
	router.POST("/api/notes/:id/unarchive", auth, h.unarchiveNote)
	router.POST("/api/notes/:id/trash", auth, h.trashNote)
	router.POST("/api/notes/:id/restore", auth, h.restoreNote)
	router.DELETE("/api/notes/:id", auth, h.deleteNote)

	return &Workers{
		Reports:   h.reportWorker,
		Scans:     h.scanWorker,
		Scheduler: scheduler.New(database, h.scanWorker),
	}
}

func (h *Handler) health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"env":     h.config.Env,
		"version": version.Version,
	})
}

func (h *Handler) getVersion(c *gin.Context) {
	info := version.GetInfo(c.Request.Context(), version.Options{
		Branch:     h.config.Branch,
		GitHubRepo: h.config.GitHubRepo,
		Enabled:    h.config.UpdateCheckEnabled,
	})
	c.JSON(http.StatusOK, info)
}

func (h *Handler) createScan(c *gin.Context) {
	var req models.ScanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	jobID, err := h.scanWorker.CreateJob(ctx, req.Host, c.ClientIP())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("enqueue scan failed: %v", err)})
		return
	}

	job, err := h.scanWorker.GetJob(ctx, jobID)
	if err != nil || job == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load scan job"})
		return
	}

	c.JSON(http.StatusAccepted, job)
}

func (h *Handler) getScan(c *gin.Context) {
	jobID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid scan id"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	job, err := h.scanWorker.GetJob(ctx, jobID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load scan job"})
		return
	}
	if job == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "scan not found"})
		return
	}

	if job.Status == models.ScanJobCompleted && job.SnapshotID != nil {
		snapshot, err := h.loadSnapshot(ctx, *job.SnapshotID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load snapshot"})
			return
		}
		if snapshot != nil {
			job.Snapshot = snapshot
		}
	}

	c.JSON(http.StatusOK, job)
}

func (h *Handler) loadSnapshot(ctx context.Context, snapshotID uuid.UUID) (*models.Snapshot, error) {
	var snapshot models.Snapshot
	var rawJSON []byte
	var changeDetailsJSON []byte
	err := h.db.Pool.QueryRow(ctx, `
		SELECT id, target_id, scanned_at, last_seen, data_hash, raw_data, changes,
			COALESCE(change_details, '[]') AS change_details, client_ip
		FROM snapshots
		WHERE id = $1
	`, snapshotID).Scan(
		&snapshot.ID,
		&snapshot.TargetID,
		&snapshot.ScannedAt,
		&snapshot.LastSeen,
		&snapshot.DataHash,
		&rawJSON,
		&snapshot.Changes,
		&changeDetailsJSON,
		&snapshot.ClientIP,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(rawJSON, &snapshot.RawData); err != nil {
		return nil, err
	}
	snapshot.ChangeDetails, err = decodeChangeDetails(changeDetailsJSON)
	if err != nil {
		return nil, err
	}
	return &snapshot, nil
}

