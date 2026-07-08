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

	"github.com/notfixingit3/echostate/internal/db"
	"github.com/notfixingit3/echostate/internal/models"
	"github.com/notfixingit3/echostate/internal/notes"
)

func TestCreateAndUpdateNote(t *testing.T) {
	d := setupTestDB(t)
	h := &Handler{db: d}
	targetID := seedTarget(t, d, "notes.example.com")

	gin.SetMode(gin.TestMode)

	createBody := models.CreateInvestigationNoteRequest{
		Title:    "Shared transit",
		Body:     `Watch <a href="https://example.com">RIPE</a> and https://bgp.tools`,
		TargetID: &targetID,
		References: []models.NoteReference{
			{Type: "url", Label: "BGP Tools", URL: "https://bgp.tools"},
			{Type: "target", Label: "notes.example.com", ID: targetID.String()},
		},
	}
	payload, err := json.Marshal(createBody)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/notes", bytes.NewReader(payload))
	c.Request.Header.Set("Content-Type", "application/json")
	h.createNote(c)

	require.Equal(t, http.StatusCreated, w.Code)
	var created models.InvestigationNote
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))
	require.Equal(t, "Shared transit", created.Title)
	require.Contains(t, created.Body, `href="https://example.com"`)
	require.Len(t, created.References, 2)

	updateBody := models.UpdateInvestigationNoteRequest{
		Title:    "Updated title",
		Body:     "Still watching this ASN",
		TargetID: &targetID,
	}
	updatePayload, err := json.Marshal(updateBody)
	require.NoError(t, err)

	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Request = httptest.NewRequest(http.MethodPut, "/api/notes/"+created.ID.String(), bytes.NewReader(updatePayload))
	c2.Request.Header.Set("Content-Type", "application/json")
	c2.Params = gin.Params{{Key: "id", Value: created.ID.String()}}
	h.updateNote(c2)

	require.Equal(t, http.StatusOK, w2.Code)
	var updated models.InvestigationNote
	require.NoError(t, json.Unmarshal(w2.Body.Bytes(), &updated))
	require.Equal(t, "Updated title", updated.Title)
}

func TestNoteArchiveTrashRestoreDelete(t *testing.T) {
	d := setupTestDB(t)
	h := &Handler{db: d}
	targetID := seedTarget(t, d, "lifecycle.example.com")
	noteID := seedNote(t, d, targetID, "Lifecycle note")

	gin.SetMode(gin.TestMode)

	post := func(path string, handler func(*gin.Context)) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, path, nil)
		c.Params = gin.Params{{Key: "id", Value: noteID.String()}}
		handler(c)
		return w
	}

	require.Equal(t, http.StatusOK, post("/api/notes/"+noteID.String()+"/archive", h.archiveNote).Code)
	require.Equal(t, http.StatusOK, post("/api/notes/"+noteID.String()+"/unarchive", h.unarchiveNote).Code)
	require.Equal(t, http.StatusOK, post("/api/notes/"+noteID.String()+"/trash", h.trashNote).Code)

	wRestore := post("/api/notes/"+noteID.String()+"/restore", h.restoreNote)
	require.Equal(t, http.StatusOK, wRestore.Code)

	updatePayload, err := json.Marshal(models.UpdateInvestigationNoteRequest{
		Title:    "Blocked in trash",
		Body:     "Should not update",
		TargetID: &targetID,
	})
	require.NoError(t, err)
	wBlocked := httptest.NewRecorder()
	cBlocked, _ := gin.CreateTestContext(wBlocked)
	cBlocked.Request = httptest.NewRequest(http.MethodPut, "/api/notes/"+noteID.String(), bytes.NewReader(updatePayload))
	cBlocked.Request.Header.Set("Content-Type", "application/json")
	cBlocked.Params = gin.Params{{Key: "id", Value: noteID.String()}}
	h.trashNote(cBlocked)
	require.Equal(t, http.StatusOK, wBlocked.Code)
	wEdit := httptest.NewRecorder()
	cEdit, _ := gin.CreateTestContext(wEdit)
	cEdit.Request = httptest.NewRequest(http.MethodPut, "/api/notes/"+noteID.String(), bytes.NewReader(updatePayload))
	cEdit.Request.Header.Set("Content-Type", "application/json")
	cEdit.Params = gin.Params{{Key: "id", Value: noteID.String()}}
	h.updateNote(cEdit)
	require.Equal(t, http.StatusNotFound, wEdit.Code)

	w8 := httptest.NewRecorder()
	c8, _ := gin.CreateTestContext(w8)
	c8.Request = httptest.NewRequest(http.MethodDelete, "/api/notes/"+noteID.String(), nil)
	c8.Params = gin.Params{{Key: "id", Value: noteID.String()}}
	h.deleteNote(c8)
	require.Equal(t, http.StatusOK, w8.Code)
}

func TestPurgeExpiredTrashedNotes(t *testing.T) {
	d := setupTestDB(t)
	targetID := seedTarget(t, d, "purge.example.com")
	noteID := seedNote(t, d, targetID, "Old trash")

	ctx := context.Background()
	_, err := d.Pool.Exec(ctx, `
		UPDATE investigation_notes
		SET status = 'trashed', trashed_at = NOW() - INTERVAL '31 days'
		WHERE id = $1
	`, noteID)
	require.NoError(t, err)

	require.NoError(t, notes.PurgeExpiredTrashed(ctx, d))

	var count int
	err = d.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM investigation_notes WHERE id = $1`, noteID).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 0, count)
}

func seedNote(t *testing.T, d *db.DB, targetID uuid.UUID, body string) uuid.UUID {
	t.Helper()
	ctx := context.Background()
	var id uuid.UUID
	err := d.Pool.QueryRow(ctx, `
		INSERT INTO investigation_notes (title, body, target_id)
		VALUES ($1, $2, $3)
		RETURNING id
	`, "seed", body, targetID).Scan(&id)
	require.NoError(t, err)
	return id
}
