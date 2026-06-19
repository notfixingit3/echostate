package audit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPurgeExpired(t *testing.T) {
	database := setupAuditTestDB(t)
	ctx := context.Background()
	svc := New(database)

	require.NoError(t, svc.Record(ctx, Entry{
		Action:    ActionSettingsUpdate,
		ActorName: "old",
	}))
	_, err := database.Pool.Exec(ctx, `
		UPDATE audit_events SET created_at = NOW() - INTERVAL '120 days'
		WHERE actor_name = 'old'
	`)
	require.NoError(t, err)

	require.NoError(t, svc.Record(ctx, Entry{
		Action:    ActionSettingsUpdate,
		ActorName: "recent",
	}))

	deleted, err := PurgeExpired(ctx, database, 90)
	require.NoError(t, err)
	require.Equal(t, int64(1), deleted)

	result, err := svc.List(ctx, ListOptions{Page: 1, Limit: 10})
	require.NoError(t, err)
	require.Equal(t, 1, result.Total)
	require.Equal(t, "recent", result.Data[0].ActorName)
}

func TestPurgeExpired_DisabledWhenZero(t *testing.T) {
	database := setupAuditTestDB(t)
	ctx := context.Background()
	svc := New(database)

	require.NoError(t, svc.Record(ctx, Entry{Action: ActionLogin, ActorName: "stay"}))
	_, err := database.Pool.Exec(ctx, `
		UPDATE audit_events SET created_at = $1
	`, time.Now().Add(-200*24*time.Hour))
	require.NoError(t, err)

	deleted, err := PurgeExpired(ctx, database, 0)
	require.NoError(t, err)
	require.Equal(t, int64(0), deleted)

	result, err := svc.List(ctx, ListOptions{Page: 1, Limit: 10})
	require.NoError(t, err)
	require.Equal(t, 1, result.Total)
}