package retention

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/notfixingit3/echostate/internal/config"
	"github.com/notfixingit3/echostate/internal/db"
)

// PruneTargetSnapshots deletes oldest snapshots beyond the configured retention limit.
func PruneTargetSnapshots(ctx context.Context, database *db.DB, targetID uuid.UUID) error {
	limit := config.GetSettings().RetentionMaxSnapshots
	if limit <= 0 {
		return nil
	}

	_, err := database.Pool.Exec(ctx, `
		DELETE FROM snapshots
		WHERE id IN (
			SELECT id
			FROM snapshots
			WHERE target_id = $1
			ORDER BY scanned_at DESC
			OFFSET $2
		)
	`, targetID, limit)
	if err != nil {
		return fmt.Errorf("prune snapshots: %w", err)
	}
	return nil
}
