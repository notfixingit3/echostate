package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/notfixingit3/echostate/internal/audit"
	"github.com/notfixingit3/echostate/internal/auth"
	"github.com/notfixingit3/echostate/internal/config"
	"github.com/notfixingit3/echostate/internal/db"
	"github.com/notfixingit3/echostate/internal/middleware"
	"github.com/notfixingit3/echostate/internal/models"
	"github.com/notfixingit3/echostate/internal/observability/health"
	"github.com/notfixingit3/echostate/internal/observability/metrics"
	"github.com/notfixingit3/echostate/internal/reports"
	"github.com/notfixingit3/echostate/internal/scanner"
	"github.com/notfixingit3/echostate/internal/scans"
	"github.com/notfixingit3/echostate/internal/scheduler"
	"github.com/notfixingit3/echostate/internal/version"
)

// scanRunner is the subset of *scanner.Scanner used by the handlers, extracted
// as an interface so tests can inject mock scanners without a live browser.
type scanRunner interface {
	Run(ctx context.Context, host string) (*models.ScanResult, error)
}

// Handler carries dependencies for HTTP handlers.
type Handler struct {
	db           *db.DB
	neo4jClient  *db.Neo4jClient
	config       *config.Config
	scanner      scanRunner
	reportWorker *reports.Worker
	scanWorker   *scans.Worker
	auth         *auth.Service
	audit        *audit.Service
}

// Register wires routes into the Gin router and returns background workers.
func Register(router *gin.Engine, database *db.DB, neo4jClient *db.Neo4jClient, cfg *config.Config, rateLimiter *middleware.RateLimiter, _ *metrics.Registry, healthRegistry *health.Registry) *Workers {
	h := &Handler{
		db:           database,
		neo4jClient:  neo4jClient,
		config:       cfg,
		scanner:      scanner.NewScanner(cfg.BrowserWSURL),
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

	h.auth = auth.NewService(database, cfg.FrontendURL)
	h.audit = audit.New(database)
	h.scanWorker = scans.New(database, h.scanner, h, config.GetSettings().ScanConcurrency)

	requireScanner := middleware.RequireAuthRole(h.auth, auth.RoleScanner)
	requireAdmin := middleware.RequireAuthRole(h.auth, auth.RoleAdmin)

	router.GET("/health", h.health)
	router.GET("/ready", health.ReadyHandler(healthRegistry))
	router.GET("/api/version", h.getVersion)

	api := router.Group("/api")
	api.Use(middleware.SessionAuth(h.auth))

	api.GET("/auth/config", h.authConfig)
	api.GET("/auth/session", h.authSession)
	api.POST("/auth/enroll/verify", rateLimiter.EnrollVerifyMiddleware(), h.enrollVerify)
	api.POST("/auth/webauthn/register/begin", h.webauthnRegisterBegin)
	api.POST("/auth/webauthn/register/finish", h.webauthnRegisterFinish)
	api.POST("/auth/webauthn/login/begin", h.webauthnLoginBegin)
	api.POST("/auth/webauthn/login/finish", h.webauthnLoginFinish)
	api.POST("/auth/logout", h.authLogout)
	api.PATCH("/auth/profile", requireScanner, h.authUpdateProfile)
	api.POST("/auth/device-code", requireScanner, h.issueDeviceCode)

	api.POST("/scan", requireScanner, rateLimiter.Middleware(), h.createScan)
	api.GET("/scans/:id", requireScanner, h.getScan)
	api.DELETE("/scans/:id", requireScanner, h.cancelScan)

	api.GET("/reports", requireScanner, h.listReports)
	api.POST("/reports", requireScanner, h.createReport)
	api.GET("/reports/:id", requireScanner, h.getReport)
	api.GET("/reports/:id/download", requireScanner, h.downloadReport)
	api.DELETE("/reports/:id", requireScanner, h.deleteReport)

	api.GET("/targets", requireScanner, h.listTargets)
	api.POST("/targets", requireScanner, h.createTarget)
	api.GET("/targets/:id", requireScanner, h.getTarget)
	api.DELETE("/targets/:id", requireAdmin, h.deleteTarget)
	api.PUT("/targets/:id/tags", requireAdmin, h.updateTargetTags)
	api.GET("/targets/:id/snapshots", requireScanner, h.listTargetSnapshots)
	api.GET("/targets/:id/screenshots", requireScanner, h.listTargetScreenshots)

	api.GET("/snapshots", requireScanner, h.listSnapshots)
	api.GET("/snapshots/:id", requireScanner, h.getSnapshot)
	api.GET("/snapshots/:id/diff", requireScanner, h.getSnapshotDiff)
	api.GET("/snapshots/:id/report", requireScanner, h.snapshotReport)
	api.DELETE("/snapshots/:id", requireAdmin, h.deleteSnapshot)

	api.GET("/webhooks", requireAdmin, h.listWebhooks)
	api.POST("/webhooks", requireAdmin, h.createWebhook)
	api.PUT("/webhooks/:id", requireAdmin, h.updateWebhook)
	api.DELETE("/webhooks/:id", requireAdmin, h.deleteWebhook)

	api.GET("/settings", requireAdmin, h.getSettings)
	api.PUT("/settings", requireAdmin, h.updateSettings)

	api.GET("/export", requireAdmin, h.exportData)
	api.POST("/import", requireAdmin, h.importData)

	api.GET("/audit", requireAdmin, h.listAuditEvents)

	api.GET("/graph", requireScanner, h.getGraph)

	api.GET("/collections", requireScanner, h.listCollections)
	api.POST("/collections", requireAdmin, h.createCollection)
	api.GET("/collections/:id", requireScanner, h.getCollection)
	api.PUT("/collections/:id", requireAdmin, h.updateCollection)
	api.DELETE("/collections/:id", requireAdmin, h.deleteCollection)
	api.POST("/collections/:id/targets", requireAdmin, h.addCollectionTarget)
	api.DELETE("/collections/:id/targets/:targetId", requireAdmin, h.removeCollectionTarget)
	api.POST("/collections/:id/rescan", requireAdmin, h.rescanCollection)

	api.GET("/notes", requireScanner, h.listNotes)
	api.GET("/notes/:id", requireScanner, h.getNote)
	api.POST("/notes", requireAdmin, h.createNote)
	api.PUT("/notes/:id", requireAdmin, h.updateNote)
	api.POST("/notes/:id/archive", requireAdmin, h.archiveNote)
	api.POST("/notes/:id/unarchive", requireAdmin, h.unarchiveNote)
	api.POST("/notes/:id/trash", requireAdmin, h.trashNote)
	api.POST("/notes/:id/restore", requireAdmin, h.restoreNote)
	api.DELETE("/notes/:id", requireAdmin, h.deleteNote)

	api.GET("/graph/views", requireScanner, h.listGraphViews)
	api.POST("/graph/views", requireAdmin, h.createGraphView)
	api.GET("/graph/views/:id", requireScanner, h.getGraphView)
	api.PUT("/graph/views/:id", requireAdmin, h.updateGraphView)
	api.DELETE("/graph/views/:id", requireAdmin, h.deleteGraphView)

	api.GET("/users", requireAdmin, h.listUsers)
	api.POST("/users", requireAdmin, h.createUser)
	api.POST("/users/:id/codes", requireAdmin, h.issueUserCode)
	api.GET("/users/:id/credentials", requireScanner, h.listUserCredentials)
	api.DELETE("/users/:id/credentials/:credId", requireScanner, h.deleteUserCredential)
	api.PATCH("/users/:id/credentials/:credId", requireScanner, h.renameUserCredential)
	api.PUT("/users/:id/credentials/:credId", requireScanner, h.renameUserCredential)

	return &Workers{
		Reports:     h.reportWorker,
		Scans:       h.scanWorker,
		Scheduler:   scheduler.New(database, h.scanWorker),
		AuditPurger: audit.NewPurger(database),
	}
}

func (h *Handler) health(c *gin.Context) {
	health.HealthHandler(h.config.Env, version.Version)(c)
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

	normalized, err := scanner.ValidateScanTarget(req.Host)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	jobID, err := h.scanWorker.CreateJob(ctx, normalized, c.ClientIP())
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

func (h *Handler) cancelScan(c *gin.Context) {
	jobID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid scan id"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	job, err := h.scanWorker.CancelJob(ctx, jobID)
	if err != nil {
		if errors.Is(err, scans.ErrScanNotCancellable) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to cancel scan"})
		return
	}
	if job == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "scan not found"})
		return
	}

	h.recordAudit(c, audit.ActionScanCancel, "scan", jobID.String(), map[string]any{
		"host": job.Host,
	})
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
