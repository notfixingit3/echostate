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
	"github.com/notfixingit3/echostate/internal/models"
	"github.com/notfixingit3/echostate/internal/scanner"
)

// Handler carries dependencies for HTTP handlers.
type Handler struct {
	db      *db.DB
	config  *config.Config
	scanner *scanner.Scanner
}

// Register wires routes into the Gin router.
func Register(router *gin.Engine, database *db.DB, cfg *config.Config) {
	h := &Handler{
		db:      database,
		config:  cfg,
		scanner: scanner.NewScanner(),
	}

	router.GET("/health", h.health)
	router.POST("/api/scan", h.createScan)
}

func (h *Handler) health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok", "env": h.config.Env})
}

func (h *Handler) createScan(c *gin.Context) {
	var req models.ScanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
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

	snapshot, err := h.storeSnapshot(ctx, targetID, dataHash, result)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("snapshot failed: %v", err)})
		return
	}

	c.JSON(http.StatusOK, snapshot)
}

func (h *Handler) upsertTarget(ctx context.Context, host string) (uuid.UUID, error) {
	var id uuid.UUID

	err := h.db.Pool.QueryRow(ctx, `
		INSERT INTO targets (host)
		VALUES ($1)
		ON CONFLICT (host) DO UPDATE SET host = EXCLUDED.host
		RETURNING id
	`, host).Scan(&id)

	return id, err
}

func (h *Handler) storeSnapshot(ctx context.Context, targetID uuid.UUID, dataHash string, result *models.ScanResult) (*models.Snapshot, error) {
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
			SET last_seen = $1
			WHERE id = $2
		`, now, previous.ID)
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
		INSERT INTO snapshots (id, target_id, scanned_at, last_seen, data_hash, raw_data, changes)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, snapshot.ID, snapshot.TargetID, snapshot.ScannedAt, snapshot.LastSeen, snapshot.DataHash, rawJSON, snapshot.Changes)
	if err != nil {
		return nil, fmt.Errorf("insert snapshot: %w", err)
	}

	return snapshot, nil
}

func computeChanges(previous map[string]any, current *models.ScanResult) []string {
	currentMap := map[string]any{}
	data, _ := json.Marshal(current)
	_ = json.Unmarshal(data, &currentMap)

	changes := []string{}
	for key, curVal := range currentMap {
		prevVal, ok := previous[key]
		if !ok {
			changes = append(changes, fmt.Sprintf("added %s", key))
			continue
		}

		prevJSON, _ := json.Marshal(prevVal)
		curJSON, _ := json.Marshal(curVal)
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
