package scanner

import (
	"fmt"
	"net"
	"regexp"
	"strings"
)

var hostnameLabel = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$`)

// ValidateScanTarget normalizes and validates a scan target hostname or IP.
// Hostnames must be fully qualified (contain a dot), e.g. example.com.
func ValidateScanTarget(host string) (string, error) {
	normalized := NormalizeHost(host)
	if normalized == "" {
		return "", fmt.Errorf("host is required")
	}

	if ip := net.ParseIP(normalized); ip != nil {
		return normalized, nil
	}

	if !strings.Contains(normalized, ".") {
		return "", fmt.Errorf(
			"invalid host %q: hostnames must include a domain suffix (e.g. example.com)",
			normalized,
		)
	}

	if len(normalized) > 253 {
		return "", fmt.Errorf("host is too long")
	}

	labels := strings.Split(normalized, ".")
	if len(labels) < 2 {
		return "", fmt.Errorf("invalid host %q", normalized)
	}

	for _, label := range labels {
		if label == "" || len(label) > 63 || !hostnameLabel.MatchString(label) {
			return "", fmt.Errorf("invalid host %q", normalized)
		}
	}

	tld := labels[len(labels)-1]
	if len(tld) < 2 {
		return "", fmt.Errorf("invalid host %q: domain suffix is too short", normalized)
	}

	return normalized, nil
}