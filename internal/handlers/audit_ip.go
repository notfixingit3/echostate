package handlers

import (
	"context"
	"net"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/notfixingit3/echostate/internal/pwhois"
)

const auditIPLookupTimeout = 5 * time.Second

func resolveClientIP(c *gin.Context) string {
	ip := strings.TrimSpace(c.ClientIP())
	if ip != "" && ip != "::1" {
		return ip
	}
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if candidate := strings.TrimSpace(parts[0]); candidate != "" {
			return candidate
		}
	}
	if xri := strings.TrimSpace(c.GetHeader("X-Real-IP")); xri != "" {
		return xri
	}
	return ip
}

func shouldEnrichIP(ip string) bool {
	parsed := net.ParseIP(strings.TrimSpace(ip))
	if parsed == nil {
		return false
	}
	return !parsed.IsLoopback() && !parsed.IsPrivate() && !parsed.IsUnspecified()
}

func ipInfoFromRecord(rec pwhois.PWHOISRecord) map[string]any {
	info := make(map[string]any, 6)
	if v := strings.TrimSpace(rec.OriginAS); v != "" {
		info["origin_as"] = v
	}
	if v := strings.TrimSpace(rec.OrgName); v != "" {
		info["org_name"] = v
	}
	if v := strings.TrimSpace(rec.AsnOrgName); v != "" && info["org_name"] == nil {
		info["org_name"] = v
	}
	if v := strings.TrimSpace(rec.CountryCode); v != "" {
		info["country_code"] = v
	}
	if v := strings.TrimSpace(rec.City); v != "" {
		info["city"] = v
	}
	if v := strings.TrimSpace(rec.Prefix); v != "" {
		info["prefix"] = v
	}
	if len(info) == 0 {
		return nil
	}
	return info
}

func lookupIPInfo(ctx context.Context, ip string) map[string]any {
	if !shouldEnrichIP(ip) {
		return nil
	}
	lookupCtx, cancel := context.WithTimeout(ctx, auditIPLookupTimeout)
	defer cancel()

	records, err := pwhois.Lookup(lookupCtx, []string{ip})
	if err != nil || len(records) == 0 {
		return nil
	}
	return ipInfoFromRecord(records[0])
}

func mergeIPInfo(detail map[string]any, info map[string]any) map[string]any {
	if len(info) == 0 {
		return detail
	}
	if detail == nil {
		detail = make(map[string]any)
	}
	detail["ip_info"] = info
	return detail
}
