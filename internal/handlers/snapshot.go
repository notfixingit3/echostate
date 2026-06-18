package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/notfixingit3/echostate/internal/db"
	"github.com/notfixingit3/echostate/internal/diff"
	"github.com/notfixingit3/echostate/internal/enrichment"
	"github.com/notfixingit3/echostate/internal/models"
	"github.com/notfixingit3/echostate/internal/retention"
	"github.com/notfixingit3/echostate/internal/scanner"
	"github.com/notfixingit3/echostate/internal/webhooks"
)

// PersistScan upserts the target, stores the snapshot, offloads blobs, and triggers side effects.
func (h *Handler) PersistScan(ctx context.Context, clientIP string, result *models.ScanResult) (*models.Snapshot, error) {
	targetID, err := h.upsertTarget(ctx, result.Host)
	if err != nil {
		return nil, fmt.Errorf("target upsert failed: %w", err)
	}

	thumbnailBytes, cleanedScreenshot := extractScreenshotThumbnail(result.Screenshot)
	if cleanedScreenshot != nil {
		result.Screenshot = cleanedScreenshot
	}

	dataHash, err := scanner.HashForDedup(result)
	if err != nil {
		return nil, fmt.Errorf("hash failed: %w", err)
	}

	snapshot, err := h.storeSnapshot(ctx, targetID, dataHash, result, clientIP)
	if err != nil {
		return nil, err
	}

	if len(thumbnailBytes) > 0 {
		blobID, err := h.db.StoreSnapshotBlob(ctx, snapshot.ID, db.BlobKindScreenshotThumbnail, "image/jpeg", thumbnailBytes)
		if err != nil {
			return nil, fmt.Errorf("store screenshot blob: %w", err)
		}
		if snapshot.RawData != nil {
			if shot, ok := snapshot.RawData["screenshot"].(map[string]any); ok {
				shot["blob_id"] = blobID.String()
				delete(shot, "thumbnail")
			}
		}
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
	go func() {
		_ = enrichment.EnrichSnapshot(context.Background(), h.db, snapshot.ID)
	}()
	go func() {
		_ = retention.PruneTargetSnapshots(context.Background(), h.db, targetID)
	}()

	return snapshot, nil
}

func (h *Handler) frontendURL() string {
	if h.config == nil {
		return ""
	}
	return h.config.FrontendURL
}

func extractScreenshotThumbnail(screenshot map[string]any) ([]byte, map[string]any) {
	if len(screenshot) == 0 {
		return nil, screenshot
	}

	rawThumb, ok := screenshot["thumbnail"].(string)
	if !ok || rawThumb == "" {
		return nil, screenshot
	}

	thumbBytes, err := base64.StdEncoding.DecodeString(rawThumb)
	if err != nil || len(thumbBytes) == 0 {
		return nil, screenshot
	}

	cleaned := make(map[string]any, len(screenshot))
	for key, value := range screenshot {
		if key == "thumbnail" {
			continue
		}
		cleaned[key] = value
	}
	return thumbBytes, cleaned
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
			ClientIP:  clientIP,
		}, nil
	}

	var changes []string
	var changeDetails []models.ChangeDetail
	if hasPrevious {
		entries := diff.Compute(previous.RawData, result)
		changes = diff.Summaries(entries)
		changeDetails = toChangeDetails(entries)
	}

	snapshot := &models.Snapshot{
		ID:            uuid.New(),
		TargetID:      targetID,
		ScannedAt:     now,
		LastSeen:      now,
		DataHash:      dataHash,
		RawData:       map[string]any{},
		Changes:       changes,
		ChangeDetails: changeDetails,
		ClientIP:      clientIP,
	}

	if err := json.Unmarshal(rawJSON, &snapshot.RawData); err != nil {
		return nil, fmt.Errorf("unmarshal raw data: %w", err)
	}

	detailsJSON, err := json.Marshal(changeDetails)
	if err != nil {
		return nil, fmt.Errorf("marshal change details: %w", err)
	}

	_, err = h.db.Pool.Exec(ctx, `
		INSERT INTO snapshots (id, target_id, scanned_at, last_seen, data_hash, raw_data, changes, change_details, client_ip)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, snapshot.ID, snapshot.TargetID, snapshot.ScannedAt, snapshot.LastSeen, snapshot.DataHash, rawJSON, snapshot.Changes, detailsJSON, clientIP)
	if err != nil {
		return nil, fmt.Errorf("insert snapshot: %w", err)
	}

	return snapshot, nil
}

func decodeChangeDetails(data []byte) ([]models.ChangeDetail, error) {
	if len(data) == 0 {
		return nil, nil
	}
	var details []models.ChangeDetail
	if err := json.Unmarshal(data, &details); err != nil {
		return nil, err
	}
	return details, nil
}

func toChangeDetails(entries []diff.Entry) []models.ChangeDetail {
	out := make([]models.ChangeDetail, 0, len(entries))
	for _, entry := range entries {
		out = append(out, models.ChangeDetail{
			Type:     entry.Type,
			Severity: entry.Severity,
			Summary:  entry.Summary,
			Field:    entry.Field,
			Detail:   entry.Detail,
		})
	}
	return out
}