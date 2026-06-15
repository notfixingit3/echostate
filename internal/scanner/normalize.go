package scanner

import (
	"net/url"
	"strings"
)

// NormalizeHost strips scheme, www. prefix, path, query, fragment, port,
// username/password, and lowercases the remaining host or IP.
func NormalizeHost(host string) string {
	if host = strings.TrimSpace(host); host == "" {
		return ""
	}

	// Prepend a scheme if missing so url.Parse treats it as a host.
	input := host
	if !strings.Contains(input, "://") {
		input = "https://" + input
	}

	u, err := url.Parse(input)
	if err != nil {
		// Fallback: strip the most common prefix manually.
		host = strings.TrimPrefix(host, "http://")
		host = strings.TrimPrefix(host, "https://")
		host = strings.TrimPrefix(host, "www.")
		return strings.ToLower(host)
	}

	host = u.Hostname()
	if host == "" {
		host = u.Path
	}

	host = strings.TrimPrefix(host, "www.")
	host = strings.TrimPrefix(host, "WWW.")
	return strings.ToLower(host)
}
