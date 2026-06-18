package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/notfixingit3/echostate/internal/models"
	"github.com/notfixingit3/echostate/internal/notes"
)

func (h *Handler) listNotes(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	_ = notes.PurgeExpiredTrashed(ctx, h.db)

	status := strings.TrimSpace(c.Query("status"))
	if status == "" {
		status = string(models.NoteStatusActive)
	}
	if !models.NoteStatus(status).IsValid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
		return
	}

	page, limit, ok := parseBrowsePagination(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pagination"})
		return
	}
	if limit > 200 {
		limit = 200
	}
	offset := (page - 1) * limit

	args := []any{status}
	where := "n.status = $1"
	argPos := 2

	if targetID := strings.TrimSpace(c.Query("target_id")); targetID != "" {
		parsed, err := uuid.Parse(targetID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid target_id"})
			return
		}
		where += " AND n.target_id = $" + itoa(argPos)
		args = append(args, parsed)
		argPos++
	}
	if snapshotID := strings.TrimSpace(c.Query("snapshot_id")); snapshotID != "" {
		parsed, err := uuid.Parse(snapshotID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid snapshot_id"})
			return
		}
		where += " AND n.snapshot_id = $" + itoa(argPos)
		args = append(args, parsed)
		argPos++
	}
	if collectionID := strings.TrimSpace(c.Query("collection_id")); collectionID != "" {
		parsed, err := uuid.Parse(collectionID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid collection_id"})
			return
		}
		where += " AND n.collection_id = $" + itoa(argPos)
		args = append(args, parsed)
		argPos++
	}
	if graphNodeID := strings.TrimSpace(c.Query("graph_node_id")); graphNodeID != "" {
		where += " AND n.graph_node_id = $" + itoa(argPos)
		args = append(args, graphNodeID)
		argPos++
	}

	var total int
	countQuery := "SELECT COUNT(*) FROM investigation_notes n WHERE " + where
	if err := h.db.Pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to count notes"})
		return
	}

	listArgs := append(args, limit, offset)
	query := `
		SELECT n.id, n.title, n.body, n.status, n.target_id, n.snapshot_id, n.collection_id,
			n.graph_node_id, n.graph_node_label, n.graph_node_type, n.refs,
			n.trashed_at, n.archived_at, n.created_at, n.updated_at,
			t.host, c.name
		FROM investigation_notes n
		LEFT JOIN targets t ON t.id = n.target_id
		LEFT JOIN target_collections c ON c.id = n.collection_id
		WHERE ` + where + `
		ORDER BY n.updated_at DESC
		LIMIT $` + itoa(argPos) + ` OFFSET $` + itoa(argPos+1)

	rows, err := h.db.Pool.Query(ctx, query, listArgs...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list notes"})
		return
	}
	defer rows.Close()

	items, err := scanNoteRows(rows)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read notes"})
		return
	}
	if items == nil {
		items = []models.InvestigationNote{}
	}

	c.JSON(http.StatusOK, models.PaginatedResponse[models.InvestigationNote]{
		Data:  items,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

func (h *Handler) getNote(c *gin.Context) {
	noteID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid note id"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	_ = notes.PurgeExpiredTrashed(ctx, h.db)

	item, err := h.loadNote(ctx, noteID)
	if err == pgx.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "note not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load note"})
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *Handler) createNote(c *gin.Context) {
	var req models.CreateInvestigationNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	body := notes.SanitizeBody(req.Body)
	if body == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body is required"})
		return
	}
	if !hasNoteAnchor(req.TargetID, req.SnapshotID, req.CollectionID, req.GraphNodeID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one anchor is required"})
		return
	}

	refs, err := normalizeNoteReferences(req.References)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	refsJSON, err := json.Marshal(refs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to encode references"})
		return
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		title = deriveNoteTitle(body)
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var noteID uuid.UUID
	err = h.db.Pool.QueryRow(ctx, `
		INSERT INTO investigation_notes (
			title, body, status, target_id, snapshot_id, collection_id,
			graph_node_id, graph_node_label, graph_node_type, refs
		)
		VALUES ($1, $2, 'active', $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`,
		title,
		body,
		req.TargetID,
		req.SnapshotID,
		req.CollectionID,
		trimOptionalString(req.GraphNodeID),
		trimOptionalString(req.GraphNodeLabel),
		trimOptionalString(req.GraphNodeType),
		refsJSON,
	).Scan(&noteID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create note"})
		return
	}

	item, err := h.loadNote(ctx, noteID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load note"})
		return
	}

	c.JSON(http.StatusCreated, item)
}

func (h *Handler) updateNote(c *gin.Context) {
	noteID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid note id"})
		return
	}

	var req models.UpdateInvestigationNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	body := notes.SanitizeBody(req.Body)
	if body == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body is required"})
		return
	}
	if !hasNoteAnchor(req.TargetID, req.SnapshotID, req.CollectionID, req.GraphNodeID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one anchor is required"})
		return
	}

	refs, err := normalizeNoteReferences(req.References)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	refsJSON, err := json.Marshal(refs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to encode references"})
		return
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		title = deriveNoteTitle(body)
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	tag, err := h.db.Pool.Exec(ctx, `
		UPDATE investigation_notes
		SET title = $1,
			body = $2,
			target_id = $3,
			snapshot_id = $4,
			collection_id = $5,
			graph_node_id = $6,
			graph_node_label = $7,
			graph_node_type = $8,
			refs = $9,
			updated_at = NOW()
		WHERE id = $10 AND status <> 'trashed'
	`,
		title,
		body,
		req.TargetID,
		req.SnapshotID,
		req.CollectionID,
		trimOptionalString(req.GraphNodeID),
		trimOptionalString(req.GraphNodeLabel),
		trimOptionalString(req.GraphNodeType),
		refsJSON,
		noteID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update note"})
		return
	}
	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "note not found or in trash"})
		return
	}

	item, err := h.loadNote(ctx, noteID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load note"})
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *Handler) archiveNote(c *gin.Context) {
	h.transitionNoteStatus(c, "archive")
}

func (h *Handler) unarchiveNote(c *gin.Context) {
	h.transitionNoteStatus(c, "unarchive")
}

func (h *Handler) trashNote(c *gin.Context) {
	h.transitionNoteStatus(c, "trash")
}

func (h *Handler) restoreNote(c *gin.Context) {
	h.transitionNoteStatus(c, "restore")
}

func (h *Handler) deleteNote(c *gin.Context) {
	noteID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid note id"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	tag, err := h.db.Pool.Exec(ctx, `
		DELETE FROM investigation_notes
		WHERE id = $1 AND status = 'trashed'
	`, noteID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete note"})
		return
	}
	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "note must be in trash before permanent delete"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

func (h *Handler) transitionNoteStatus(c *gin.Context, action string) {
	noteID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid note id"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var query string
	switch action {
	case "archive":
		query = `
			UPDATE investigation_notes
			SET status = 'archived', archived_at = NOW(), trashed_at = NULL, updated_at = NOW()
			WHERE id = $1 AND status = 'active'
		`
	case "unarchive":
		query = `
			UPDATE investigation_notes
			SET status = 'active', archived_at = NULL, updated_at = NOW()
			WHERE id = $1 AND status = 'archived'
		`
	case "trash":
		query = `
			UPDATE investigation_notes
			SET status = 'trashed', trashed_at = NOW(), archived_at = NULL, updated_at = NOW()
			WHERE id = $1 AND status IN ('active', 'archived')
		`
	case "restore":
		query = `
			UPDATE investigation_notes
			SET status = 'active', trashed_at = NULL, archived_at = NULL, updated_at = NOW()
			WHERE id = $1 AND status = 'trashed'
		`
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid action"})
		return
	}

	tag, err := h.db.Pool.Exec(ctx, query, noteID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update note status"})
		return
	}
	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status transition"})
		return
	}

	item, err := h.loadNote(ctx, noteID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load note"})
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *Handler) loadNote(ctx context.Context, noteID uuid.UUID) (*models.InvestigationNote, error) {
	row := h.db.Pool.QueryRow(ctx, `
		SELECT n.id, n.title, n.body, n.status, n.target_id, n.snapshot_id, n.collection_id,
			n.graph_node_id, n.graph_node_label, n.graph_node_type, n.refs,
			n.trashed_at, n.archived_at, n.created_at, n.updated_at,
			t.host, c.name
		FROM investigation_notes n
		LEFT JOIN targets t ON t.id = n.target_id
		LEFT JOIN target_collections c ON c.id = n.collection_id
		WHERE n.id = $1
	`, noteID)

	var item models.InvestigationNote
	var refsJSON []byte
	var targetHost *string
	var collectionName *string
	err := row.Scan(
		&item.ID,
		&item.Title,
		&item.Body,
		&item.Status,
		&item.TargetID,
		&item.SnapshotID,
		&item.CollectionID,
		&item.GraphNodeID,
		&item.GraphNodeLabel,
		&item.GraphNodeType,
		&refsJSON,
		&item.TrashedAt,
		&item.ArchivedAt,
		&item.CreatedAt,
		&item.UpdatedAt,
		&targetHost,
		&collectionName,
	)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(refsJSON, &item.References); err != nil {
		return nil, err
	}
	if item.References == nil {
		item.References = []models.NoteReference{}
	}
	item.TargetHost = targetHost
	item.CollectionName = collectionName
	return &item, nil
}

func scanNoteRows(rows pgx.Rows) ([]models.InvestigationNote, error) {
	var items []models.InvestigationNote
	for rows.Next() {
		var item models.InvestigationNote
		var refsJSON []byte
		var targetHost *string
		var collectionName *string
		if err := rows.Scan(
			&item.ID,
			&item.Title,
			&item.Body,
			&item.Status,
			&item.TargetID,
			&item.SnapshotID,
			&item.CollectionID,
			&item.GraphNodeID,
			&item.GraphNodeLabel,
			&item.GraphNodeType,
			&refsJSON,
			&item.TrashedAt,
			&item.ArchivedAt,
			&item.CreatedAt,
			&item.UpdatedAt,
			&targetHost,
			&collectionName,
		); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(refsJSON, &item.References); err != nil {
			return nil, err
		}
		if item.References == nil {
			item.References = []models.NoteReference{}
		}
		item.TargetHost = targetHost
		item.CollectionName = collectionName
		items = append(items, item)
	}
	return items, rows.Err()
}

func hasNoteAnchor(targetID, snapshotID, collectionID *uuid.UUID, graphNodeID *string) bool {
	if targetID != nil || snapshotID != nil || collectionID != nil {
		return true
	}
	return trimOptionalString(graphNodeID) != nil
}

func trimOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func deriveNoteTitle(body string) string {
	plain := strings.Join(strings.Fields(notes.PlainText(body)), " ")
	if len(plain) > 80 {
		return plain[:80] + "…"
	}
	return plain
}

func normalizeNoteReferences(refs []models.NoteReference) ([]models.NoteReference, error) {
	if refs == nil {
		return []models.NoteReference{}, nil
	}
	out := make([]models.NoteReference, 0, len(refs))
	for _, ref := range refs {
		refType := strings.TrimSpace(strings.ToLower(ref.Type))
		ref.Label = strings.TrimSpace(ref.Label)
		ref.URL = strings.TrimSpace(ref.URL)
		ref.ID = strings.TrimSpace(ref.ID)
		switch refType {
		case "url":
			if ref.URL == "" || !strings.HasPrefix(ref.URL, "http://") && !strings.HasPrefix(ref.URL, "https://") {
				return nil, errInvalidReference
			}
		case "target", "snapshot", "collection", "graph_node":
			if ref.ID == "" {
				return nil, errInvalidReference
			}
			if _, err := uuid.Parse(ref.ID); err != nil && refType != "graph_node" {
				return nil, errInvalidReference
			}
		default:
			return nil, errInvalidReference
		}
		out = append(out, models.NoteReference{
			Type:  refType,
			Label: ref.Label,
			URL:   ref.URL,
			ID:    ref.ID,
		})
	}
	return out, nil
}

var errInvalidReference = &referenceError{message: "invalid reference entry"}

type referenceError struct {
	message string
}

func (e *referenceError) Error() string {
	return e.message
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	digits := []byte{}
	for value > 0 {
		digits = append([]byte{byte('0' + value%10)}, digits...)
		value /= 10
	}
	return string(digits)
}