package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/notfixingit3/echostate/internal/export"
	"github.com/notfixingit3/echostate/internal/models"
)

func TestExportImportRoundTrip(t *testing.T) {
	source := setupTestDB(t)
	target := setupTestDB(t)

	targetID := seedTarget(t, source, "transfer.example.com")
	snapshotID := seedSnapshotForTarget(t, source, targetID, map[string]any{
		"host": "transfer.example.com",
		"web":  map[string]any{"title": "Transfer Test"},
	}, "203.0.113.10")

	ctx := context.Background()
	collectionID := uuid.New()
	_, err := source.Pool.Exec(ctx, `
		INSERT INTO target_collections (id, name, description)
		VALUES ($1, 'Transfer set', 'For export test')
	`, collectionID)
	require.NoError(t, err)
	_, err = source.Pool.Exec(ctx, `
		INSERT INTO target_collection_members (collection_id, target_id)
		VALUES ($1, $2)
	`, collectionID, targetID)
	require.NoError(t, err)

	noteID := seedNote(t, source, targetID, "Transfer note")

	bundle, err := export.BuildBundle(ctx, source, export.Options{IncludeBlobs: true})
	require.NoError(t, err)
	require.NotEmpty(t, bundle.Targets)
	require.NotEmpty(t, bundle.Snapshots)
	require.NotEmpty(t, bundle.Collections)
	require.NotEmpty(t, bundle.Notes)

	result, err := export.ImportBundle(ctx, target, nil, bundle, export.ImportOptions{
		Conflict: "overwrite",
	})
	require.NoError(t, err)
	require.Equal(t, 1, result.TargetsImported)
	require.Equal(t, 1, result.SnapshotsImported)
	require.Equal(t, 1, result.CollectionsImported)
	require.Equal(t, 1, result.CollectionMembersImported)
	require.Equal(t, 1, result.NotesImported)

	var importedHost string
	err = target.Pool.QueryRow(ctx, `SELECT host FROM targets WHERE id = $1`, targetID).Scan(&importedHost)
	require.NoError(t, err)
	require.Equal(t, "transfer.example.com", importedHost)

	var importedSnapshot uuid.UUID
	err = target.Pool.QueryRow(ctx, `SELECT id FROM snapshots WHERE id = $1`, snapshotID).Scan(&importedSnapshot)
	require.NoError(t, err)

	var importedNoteBody string
	err = target.Pool.QueryRow(ctx, `SELECT body FROM investigation_notes WHERE id = $1`, noteID).Scan(&importedNoteBody)
	require.NoError(t, err)
	require.Equal(t, "Transfer note", importedNoteBody)
}

func TestExportImportHTTPHandlers(t *testing.T) {
	d := setupTestDB(t)
	h := &Handler{db: d}
	seedTarget(t, d, "http-export.example.com")

	gin.SetMode(gin.TestMode)

	exportReq := httptest.NewRequest(http.MethodGet, "/api/export?include_blobs=false", nil)
	exportRes := httptest.NewRecorder()
	exportCtx, _ := gin.CreateTestContext(exportRes)
	exportCtx.Request = exportReq
	h.exportData(exportCtx)
	require.Equal(t, http.StatusOK, exportRes.Code)

	var bundle models.ExportBundle
	require.NoError(t, json.Unmarshal(exportRes.Body.Bytes(), &bundle))
	require.NotEmpty(t, bundle.Targets)

	importPayload, err := json.Marshal(models.ImportRequest{
		Conflict:     "skip",
		RebuildGraph: false,
		Bundle:       bundle,
	})
	require.NoError(t, err)

	importReq := httptest.NewRequest(http.MethodPost, "/api/import", bytes.NewReader(importPayload))
	importReq.Header.Set("Content-Type", "application/json")
	importRes := httptest.NewRecorder()
	importCtx, _ := gin.CreateTestContext(importRes)
	importCtx.Request = importReq
	h.importData(importCtx)
	require.Equal(t, http.StatusOK, importRes.Code)

	var result models.ImportResult
	require.NoError(t, json.Unmarshal(importRes.Body.Bytes(), &result))
	require.Equal(t, 0, result.TargetsImported)
	require.Equal(t, 1, result.TargetsSkipped)
}

func TestImportBundleRejectsUnknownFormat(t *testing.T) {
	d := setupTestDB(t)
	ctx := context.Background()

	_, err := export.ImportBundle(ctx, d, nil, &models.ExportBundle{FormatVersion: 99}, export.ImportOptions{})
	require.Error(t, err)
}
