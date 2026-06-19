package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/notfixingit3/echostate/internal/db"
	"github.com/notfixingit3/echostate/internal/models"
)

// Entry is a single audit log row.
type Entry struct {
	UserID       *uuid.UUID
	ActorName    string
	ActorRole    string
	Action       string
	ResourceType string
	ResourceID   string
	Detail       map[string]any
	IP           string
	UserAgent    string
}

// Service persists and queries audit events.
type Service struct {
	db *db.DB
}

func New(database *db.DB) *Service {
	return &Service{db: database}
}

// Record inserts an audit event. Failures are returned to the caller.
func (s *Service) Record(ctx context.Context, entry Entry) error {
	if strings.TrimSpace(entry.Action) == "" {
		return fmt.Errorf("audit action is required")
	}

	detail := entry.Detail
	if detail == nil {
		detail = map[string]any{}
	}
	detailJSON, err := json.Marshal(detail)
	if err != nil {
		return fmt.Errorf("marshal audit detail: %w", err)
	}

	_, err = s.db.Pool.Exec(ctx, `
		INSERT INTO audit_events (
			user_id, actor_name, actor_role, action,
			resource_type, resource_id, detail, ip, user_agent
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`,
		entry.UserID,
		strings.TrimSpace(entry.ActorName),
		strings.TrimSpace(entry.ActorRole),
		strings.TrimSpace(entry.Action),
		strings.TrimSpace(entry.ResourceType),
		strings.TrimSpace(entry.ResourceID),
		detailJSON,
		strings.TrimSpace(entry.IP),
		strings.TrimSpace(entry.UserAgent),
	)
	if err != nil {
		return fmt.Errorf("insert audit event: %w", err)
	}
	return nil
}

// ListOptions filters paginated audit queries.
type ListOptions struct {
	Page   int
	Limit  int
	Action string
	Query  string
}

// List returns audit events newest first.
func (s *Service) List(ctx context.Context, opts ListOptions) (models.PaginatedResponse[models.AuditEvent], error) {
	page := opts.Page
	if page < 1 {
		page = 1
	}
	limit := opts.Limit
	if limit < 1 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	offset := (page - 1) * limit

	action := strings.TrimSpace(opts.Action)
	query := strings.TrimSpace(opts.Query)

	where := make([]string, 0, 2)
	args := make([]any, 0, 4)
	argPos := 1

	if action != "" {
		where = append(where, fmt.Sprintf("action = $%d", argPos))
		args = append(args, action)
		argPos++
	}
	if query != "" {
		pattern := "%" + query + "%"
		where = append(where, fmt.Sprintf(`(
			actor_name ILIKE $%d OR
			resource_id ILIKE $%d OR
			detail::text ILIKE $%d
		)`, argPos, argPos, argPos))
		args = append(args, pattern)
		argPos++
	}

	whereSQL := ""
	if len(where) > 0 {
		whereSQL = "WHERE " + strings.Join(where, " AND ")
	}

	countQuery := "SELECT COUNT(*) FROM audit_events " + whereSQL
	var total int
	if err := s.db.Pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return models.PaginatedResponse[models.AuditEvent]{}, fmt.Errorf("count audit events: %w", err)
	}

	listQuery := fmt.Sprintf(`
		SELECT id, created_at, user_id, actor_name, actor_role, action,
			resource_type, resource_id, detail, ip, user_agent
		FROM audit_events
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, argPos, argPos+1)
	listArgs := append(append([]any{}, args...), limit, offset)

	rows, err := s.db.Pool.Query(ctx, listQuery, listArgs...)
	if err != nil {
		return models.PaginatedResponse[models.AuditEvent]{}, fmt.Errorf("list audit events: %w", err)
	}
	defer rows.Close()

	events := make([]models.AuditEvent, 0, limit)
	for rows.Next() {
		var event models.AuditEvent
		var userID *uuid.UUID
		var detailJSON []byte
		if err := rows.Scan(
			&event.ID,
			&event.CreatedAt,
			&userID,
			&event.ActorName,
			&event.ActorRole,
			&event.Action,
			&event.ResourceType,
			&event.ResourceID,
			&detailJSON,
			&event.IP,
			&event.UserAgent,
		); err != nil {
			return models.PaginatedResponse[models.AuditEvent]{}, fmt.Errorf("scan audit event: %w", err)
		}
		event.UserID = userID
		if len(detailJSON) > 0 {
			_ = json.Unmarshal(detailJSON, &event.Detail)
		}
		if event.Detail == nil {
			event.Detail = map[string]any{}
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return models.PaginatedResponse[models.AuditEvent]{}, fmt.Errorf("iterate audit events: %w", err)
	}

	return models.PaginatedResponse[models.AuditEvent]{
		Data:  events,
		Page:  page,
		Limit: limit,
		Total: total,
	}, nil
}

