package handlers

import (
	"context"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/notfixingit3/echostate/internal/audit"
	"github.com/notfixingit3/echostate/internal/middleware"
	"github.com/notfixingit3/echostate/internal/observability/log"
)

func (h *Handler) recordAudit(c *gin.Context, action, resourceType, resourceID string, detail map[string]any) {
	if h.audit == nil {
		return
	}

	ip := resolveClientIP(c)
	entry := audit.Entry{
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Detail:       detail,
		IP:           ip,
		UserAgent:    c.Request.UserAgent(),
	}
	if user := middleware.GetUser(c); user != nil {
		entry.UserID = &user.ID
		entry.ActorName = user.DisplayName
		entry.ActorRole = user.Role
	}

	go func(entry audit.Entry) {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		entry.Detail = mergeIPInfo(entry.Detail, lookupIPInfo(ctx, entry.IP))
		if err := h.audit.Record(ctx, entry); err != nil {
			log.FromContext(ctx).Error("audit record failed",
				slog.String("action", entry.Action),
				slog.String("error", err.Error()),
			)
		}
	}(entry)
}

func (h *Handler) recordAuditActor(
	action, resourceType, resourceID string,
	userID uuid.UUID,
	actorName, actorRole, ip, userAgent string,
	detail map[string]any,
) {
	if h.audit == nil {
		return
	}

	entry := audit.Entry{
		UserID:       &userID,
		ActorName:    actorName,
		ActorRole:    actorRole,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Detail:       detail,
		IP:           ip,
		UserAgent:    userAgent,
	}

	go func(entry audit.Entry) {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		entry.Detail = mergeIPInfo(entry.Detail, lookupIPInfo(ctx, entry.IP))
		if err := h.audit.Record(ctx, entry); err != nil {
			log.FromContext(ctx).Error("audit record failed",
				slog.String("action", entry.Action),
				slog.String("error", err.Error()),
			)
		}
	}(entry)
}
