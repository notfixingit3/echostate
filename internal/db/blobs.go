package db

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const BlobKindScreenshotThumbnail = "screenshot_thumbnail"

// StoreSnapshotBlob persists binary snapshot data keyed by snapshot and kind.
func (db *DB) StoreSnapshotBlob(ctx context.Context, snapshotID uuid.UUID, kind, contentType string, data []byte) (uuid.UUID, error) {
	var blobID uuid.UUID
	err := db.Pool.QueryRow(ctx, `
		INSERT INTO snapshot_blobs (snapshot_id, kind, content_type, data)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (snapshot_id, kind) DO UPDATE
		SET content_type = EXCLUDED.content_type,
		    data = EXCLUDED.data
		RETURNING id
	`, snapshotID, kind, contentType, data).Scan(&blobID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("store snapshot blob: %w", err)
	}
	return blobID, nil
}

// GetSnapshotBlob returns blob bytes for a snapshot/kind pair.
func (db *DB) GetSnapshotBlob(ctx context.Context, snapshotID uuid.UUID, kind string) ([]byte, string, error) {
	var data []byte
	var contentType string
	err := db.Pool.QueryRow(ctx, `
		SELECT data, content_type
		FROM snapshot_blobs
		WHERE snapshot_id = $1 AND kind = $2
	`, snapshotID, kind).Scan(&data, &contentType)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, "", nil
		}
		return nil, "", fmt.Errorf("get snapshot blob: %w", err)
	}
	return data, contentType, nil
}

// ThumbnailBase64ForSnapshot returns a base64 thumbnail, preferring blob storage.
func (db *DB) ThumbnailBase64ForSnapshot(ctx context.Context, snapshotID uuid.UUID, legacy string) (string, error) {
	data, _, err := db.GetSnapshotBlob(ctx, snapshotID, BlobKindScreenshotThumbnail)
	if err != nil {
		return "", err
	}
	if len(data) > 0 {
		return base64.StdEncoding.EncodeToString(data), nil
	}
	return legacy, nil
}
