package scanner

import (
	"fmt"
	"strings"
)

const cookieIntelJS = `(() => {
	const names = new Set();
	document.cookie.split(';').forEach((part) => {
		const name = part.split('=')[0]?.trim();
		if (name) names.add(name);
	});
	return Array.from(names).sort();
})()`

func parseSetCookieNames(raw any) []string {
	switch values := raw.(type) {
	case string:
		return []string{extractCookieName(values)}
	case []any:
		var names []string
		for _, item := range values {
			if name := extractCookieName(fmt.Sprint(item)); name != "" {
				names = append(names, name)
			}
		}
		return names
	default:
		text := strings.TrimSpace(fmt.Sprint(raw))
		if text == "" {
			return nil
		}
		return []string{extractCookieName(text)}
	}
}

func extractCookieName(header string) string {
	header = strings.TrimSpace(header)
	if header == "" {
		return ""
	}
	name, _, _ := strings.Cut(header, ";")
	name, _, _ = strings.Cut(strings.TrimSpace(name), "=")
	return strings.TrimSpace(name)
}

func mergeCookieNames(parts ...[]string) []string {
	seen := make(map[string]struct{})
	var out []string
	for _, list := range parts {
		for _, name := range list {
			name = strings.TrimSpace(name)
			if name == "" {
				continue
			}
			if _, ok := seen[name]; ok {
				continue
			}
			seen[name] = struct{}{}
			out = append(out, name)
		}
	}
	return out
}