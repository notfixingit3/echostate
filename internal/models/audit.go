package models

import (
	"time"

	"github.com/google/uuid"
)

// AuditEvent is a persisted admin/security audit log entry.
type AuditEvent struct {
	ID           uuid.UUID      `json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	UserID       *uuid.UUID     `json:"user_id,omitempty"`
	ActorName    string         `json:"actor_name"`
	ActorRole    string         `json:"actor_role"`
	Action       string         `json:"action"`
	ResourceType string         `json:"resource_type"`
	ResourceID   string         `json:"resource_id"`
	Detail       map[string]any `json:"detail"`
	IP           string         `json:"ip"`
	UserAgent    string         `json:"user_agent"`
}
