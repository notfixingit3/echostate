package audit

import (
	"context"
	"fmt"
	"time"

	"github.com/notfixingit3/echostate/internal/db"
)

// PurgeExpired deletes audit events older than retentionDays. Zero or negative keeps all rows.
func PurgeExpired(ctx context.Context, database *db.DB, retentionDays int) (int64, error) {
	if retentionDays <= 0 {
		return 0, nil
	}

	cutoff := time.Now().UTC().AddDate(0, 0, -retentionDays)
	tag, err := database.Pool.Exec(ctx, `
		DELETE FROM audit_events
		WHERE created_at < $1
	`, cutoff)
	if err != nil {
		return 0, fmt.Errorf("purge audit events: %w", err)
	}
	return tag.RowsAffected(), nil
}