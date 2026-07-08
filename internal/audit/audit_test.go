package audit

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/notfixingit3/echostate/internal/db"
)

func testDatabaseURL() string {
	if url := os.Getenv("DATABASE_URL"); url != "" {
		return url
	}
	return "postgres://echostate:echostate@localhost:5432/echostate?sslmode=disable"
}

func setupAuditTestDB(t *testing.T) *db.DB {
	t.Helper()

	schema := fmt.Sprintf("audit_test_%s", uuid.New().String()[:8])
	baseURL := testDatabaseURL()
	sep := "&"
	if !strings.ContainsRune(baseURL, '?') {
		sep = "?"
	}
	schemaURL := fmt.Sprintf("%s%soptions=--search_path%%3D%s", baseURL, sep, schema)

	database, err := db.Connect(schemaURL)
	if err != nil {
		t.Skipf("database not available: %v", err)
	}

	ctx := context.Background()
	_, err = database.Pool.Exec(ctx, fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schema))
	require.NoError(t, err)
	require.NoError(t, db.Migrate(database))

	t.Cleanup(func() {
		database.Close()
	})

	return database
}

func TestRecordAndList(t *testing.T) {
	database := setupAuditTestDB(t)
	svc := New(database)
	ctx := context.Background()

	require.NoError(t, svc.Record(ctx, Entry{
		ActorName:    "Admin",
		ActorRole:    "admin",
		Action:       ActionTargetDelete,
		ResourceType: "target",
		ResourceID:   uuid.New().String(),
		Detail:       map[string]any{"host": "example.com"},
		IP:           "127.0.0.1",
		UserAgent:    "test",
	}))

	result, err := svc.List(ctx, ListOptions{Page: 1, Limit: 10, Action: ActionTargetDelete})
	require.NoError(t, err)
	require.Equal(t, 1, result.Total)
	require.Len(t, result.Data, 1)
	require.Equal(t, "Admin", result.Data[0].ActorName)
	require.Equal(t, ActionTargetDelete, result.Data[0].Action)
	require.Equal(t, "example.com", result.Data[0].Detail["host"])
}

func TestListSearchMatchesIP(t *testing.T) {
	database := setupAuditTestDB(t)
	svc := New(database)
	ctx := context.Background()

	require.NoError(t, svc.Record(ctx, Entry{
		ActorName: "Admin",
		ActorRole: "admin",
		Action:    ActionSettingsUpdate,
		IP:        "203.0.113.55",
	}))

	result, err := svc.List(ctx, ListOptions{Page: 1, Limit: 10, Query: "203.0.113"})
	require.NoError(t, err)
	require.Equal(t, 1, result.Total)
	require.Equal(t, "203.0.113.55", result.Data[0].IP)
}
