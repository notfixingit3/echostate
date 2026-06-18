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
	"github.com/notfixingit3/echostate/internal/scanner"
	"github.com/notfixingit3/echostate/internal/version"
	"github.com/notfixingit3/echostate/internal/webhooks"
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
	worker      *reports.Worker
}

// Register wires routes into the Gin router and returns the report worker so
// the caller can manage its lifecycle.
func Register(router *gin.Engine, database *db.DB, neo4jClient *db.Neo4jClient, cfg *config.Config, rateLimiter *middleware.RateLimiter) *reports.Worker {
	h := &Handler{
		db:          database,
		neo4jClient: neo4jClient,
		config:      cfg,
		scanner:     scanner.NewScanner(cfg.BrowserWSURL),
		worker:      newWorker(database),
	}

	var valBytes []byte
	err := database.Pool.QueryRow(context.Background(), `SELECT value FROM settings WHERE key = 'app_settings'`).Scan(&valBytes)
	if err == nil {
		var s config.SystemSettings
		if err := json.Unmarshal(valBytes, &s); err == nil {
			config.UpdateSettings(s)
		}
	}

	router.GET("/health", h.health)
	router.POST("/api/scan", rateLimiter.Middleware(), h.createScan)

	router.GET("/api/reports", h.listReports)
	router.POST("/api/reports", h.createReport)
	router.GET("/api/reports/:id", h.getReport)
	router.GET("/api/reports/:id/download", h.downloadReport)

	router.GET("/api/targets", h.listTargets)
	router.GET("/api/targets/:id", h.getTarget)
	router.PUT("/api/targets/:id/tags", h.updateTargetTags)
	router.GET("/api/targets/:id/snapshots", h.listTargetSnapshots)
	router.GET("/api/targets/:id/screenshots", h.listTargetScreenshots)

	router.GET("/api/snapshots", h.listSnapshots)
	router.GET("/api/snapshots/:id", h.getSnapshot)
	router.GET("/api/snapshots/:id/diff", h.getSnapshotDiff)
	router.GET("/api/snapshots/:id/report", h.snapshotReport)

	router.GET("/api/webhooks", h.listWebhooks)
	router.POST("/api/webhooks", h.createWebhook)
	router.PUT("/api/webhooks/:id", h.updateWebhook)
	router.DELETE("/api/webhooks/:id", h.deleteWebhook)

	router.GET("/api/settings", h.getSettings)
	router.PUT("/api/settings", h.updateSettings)

	router.GET("/api/graph", h.getGraph)

	return h.worker
}

func (h *Handler) health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"env":     h.config.Env,
		"version": version.Version,
	})
}

func (h *Handler) createScan(c *gin.Context) {
	var req models.ScanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 100*time.Second)
	defer cancel()

	result, err := h.scanner.Run(ctx, req.Host)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("scan failed: %v", err)})
		return
	}

	dataHash, err := scanner.Hash(result)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("hash failed: %v", err)})
		return
	}

	targetID, err := h.upsertTarget(ctx, req.Host)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("target upsert failed: %v", err)})
		return
	}

	clientIP := c.ClientIP()

	snapshot, err := h.storeSnapshot(ctx, targetID, dataHash, result, clientIP)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("snapshot failed: %v", err)})
		return
	}

	c.JSON(http.StatusOK, snapshot)
}

func (h *Handler) upsertTarget(ctx context.Context, host string) (uuid.UUID, error) {
	var id uuid.UUID

	normalized := scanner.NormalizeHost(host)

	err := h.db.Pool.QueryRow(ctx, `
		INSERT INTO targets (host, normalized_host)
		VALUES ($1, $2)
		ON CONFLICT (host) DO UPDATE SET normalized_host = EXCLUDED.normalized_host
		RETURNING id
	`, host, normalized).Scan(&id)

	return id, err
}

func (h *Handler) storeSnapshot(ctx context.Context, targetID uuid.UUID, dataHash string, result *models.ScanResult, clientIP string) (*models.Snapshot, error) {
	var previous models.Snapshot
	errPrev := h.db.Pool.QueryRow(ctx, `
		SELECT id, data_hash, raw_data
		FROM snapshots
		WHERE target_id = $1
		ORDER BY scanned_at DESC
		LIMIT 1
	`, targetID).Scan(&previous.ID, &previous.DataHash, &previous.RawData)
	if errPrev != nil && errPrev != pgx.ErrNoRows {
		return nil, fmt.Errorf("fetch previous snapshot: %w", errPrev)
	}

	hasPrevious := errPrev == nil

	rawJSON, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("marshal raw data: %w", err)
	}

	now := time.Now().UTC()

	if hasPrevious && previous.DataHash == dataHash {
		_, err := h.db.Pool.Exec(ctx, `
			UPDATE snapshots
			SET last_seen = $1, client_ip = $2
			WHERE id = $3
		`, now, clientIP, previous.ID)
		if err != nil {
			return nil, fmt.Errorf("update last_seen: %w", err)
		}

		return &models.Snapshot{
			ID:        previous.ID,
			TargetID:  targetID,
			ScannedAt: previous.ScannedAt,
			LastSeen:  now,
			DataHash:  dataHash,
			RawData:   previous.RawData,
		}, nil
	}

	var changes []string
	if hasPrevious {
		changes = computeChanges(previous.RawData, result)
	}

	snapshot := &models.Snapshot{
		ID:        uuid.New(),
		TargetID:  targetID,
		ScannedAt: now,
		LastSeen:  now,
		DataHash:  dataHash,
		RawData:   map[string]any{},
		Changes:   changes,
	}

	if err := json.Unmarshal(rawJSON, &snapshot.RawData); err != nil {
		return nil, fmt.Errorf("unmarshal raw data: %w", err)
	}

	_, err = h.db.Pool.Exec(ctx, `
		INSERT INTO snapshots (id, target_id, scanned_at, last_seen, data_hash, raw_data, changes, client_ip)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, snapshot.ID, snapshot.TargetID, snapshot.ScannedAt, snapshot.LastSeen, snapshot.DataHash, rawJSON, snapshot.Changes, clientIP)
	if err != nil {
		return nil, fmt.Errorf("insert snapshot: %w", err)
	}

	if h.neo4jClient != nil {
		target := &models.Target{
			ID:        targetID,
			Host:      result.Host,
			CreatedAt: time.Now().UTC(),
		}
		go func() {
			_ = h.neo4jClient.SyncSnapshot(context.Background(), target, snapshot)
		}()
	}

	go webhooks.Dispatch(context.Background(), h.db, result.Host, snapshot, h.frontendURL())

	return snapshot, nil
}

func (h *Handler) frontendURL() string {
	if h.config == nil {
		return ""
	}
	return h.config.FrontendURL
}

func computeChanges(previous map[string]any, current *models.ScanResult) []string {
	// Empty previous map means no prior snapshot exists, so no changes can be
	// detected. storeSnapshot already guards this path (it only calls
	// computeChanges when a previous snapshot is present), but direct callers
	// and unit tests expect empty changes when there is no previous data.
	if len(previous) == 0 {
		return []string{}
	}

	currentMap := map[string]any{}
	data, err := json.Marshal(current)
	if err != nil {
		return []string{fmt.Sprintf("marshal current: %v", err)}
	}
	if err := json.Unmarshal(data, &currentMap); err != nil {
		return []string{fmt.Sprintf("unmarshal current: %v", err)}
	}

	changes := []string{}
	for key, curVal := range currentMap {
		prevVal, ok := previous[key]
		if !ok {
			changes = append(changes, fmt.Sprintf("added %s", key))
			continue
		}

		prevJSON, err := json.Marshal(prevVal)
		if err != nil {
			changes = append(changes, fmt.Sprintf("marshal previous %s: %v", key, err))
			continue
		}
		curJSON, err := json.Marshal(curVal)
		if err != nil {
			changes = append(changes, fmt.Sprintf("marshal current %s: %v", key, err))
			continue
		}
		if string(prevJSON) != string(curJSON) {
			changes = append(changes, fmt.Sprintf("changed %s", key))
		}
	}

	for key := range previous {
		if _, ok := currentMap[key]; !ok {
			changes = append(changes, fmt.Sprintf("removed %s", key))
		}
	}

	return changes
}
