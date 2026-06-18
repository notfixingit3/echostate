package notes

import (
	"context"
	"time"

	"github.com/notfixingit3/echostate/internal/db"
)

const TrashRetention = 30 * 24 * time.Hour

// PurgeExpiredTrashed permanently deletes notes in trash older than TrashRetention.
func PurgeExpiredTrashed(ctx context.Context, database *db.DB) error {
	_, err := database.Pool.Exec(ctx, `
		DELETE FROM investigation_notes
		WHERE status = 'trashed'
		  AND trashed_at IS NOT NULL
		  AND trashed_at < NOW() - ($1::interval)
	`, TrashRetention.String())
	return err
}