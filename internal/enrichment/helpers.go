package enrichment

import (
	"fmt"
	"net"
	"strings"
	"time"
)

func resolveIP(payload map[string]any) string {
	asn, _ := payload["asn"].(map[string]any)
	return strings.TrimSpace(fmt.Sprint(asn["ip"]))
}

func resolveHost(payload map[string]any) string {
	host := strings.TrimSpace(fmt.Sprint(payload["host"]))
	if host == "" || host == "<nil>" {
		return ""
	}
	if net.ParseIP(host) != nil {
		return ""
	}
	return strings.ToLower(host)
}

func truncateMatches(raw any) any {
	items, ok := raw.([]any)
	if !ok || len(items) <= 5 {
		return raw
	}
	return items[:5]
}

func truncateAny(items []any, limit int) []any {
	if len(items) <= limit {
		return items
	}
	return items[:limit]
}

func enrichmentMeta(status string) map[string]any {
	return map[string]any{
		"status":       status,
		"completed_at": time.Now().UTC().Format(time.RFC3339),
	}
}