package scanner

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/notfixingit3/echostate/internal/config"
)

const (
	ctMaxSubdomains    = 250
	ctResponseMaxBytes = 4 * 1024 * 1024
	ctGatherTimeout    = 75 * time.Second
	ctHTTPTimeout      = 60 * time.Second
	ctMaxFastRetries   = 3
)

var ctAPIBaseURL = "https://crt.sh/"

// gatherCT queries crt.sh certificate transparency logs for subdomains of the
// target's registrable domain.
func gatherCT(ctx context.Context, host string) (string, map[string]any, error) {
	host = NormalizeHost(host)
	if host == "" {
		return "ct", nil, fmt.Errorf("empty host")
	}

	if net.ParseIP(host) != nil {
		return "ct", map[string]any{
			"skipped": "certificate transparency applies to domain names, not IPs",
		}, nil
	}

	domain := ctSearchDomain(host)
	if domain == "" {
		return "ct", nil, fmt.Errorf("unable to derive search domain from %q", host)
	}

	gatherCtx, cancel := context.WithTimeout(ctx, ctGatherTimeoutFromSettings())
	defer cancel()

	subdomains, certs, source, err := fetchCTIntel(gatherCtx, domain)
	if err != nil {
		return "ct", nil, err
	}

	result := map[string]any{
		"domain":     domain,
		"source":     source,
		"subdomains": subdomains,
		"count":      len(subdomains),
	}
	if len(certs) > 0 {
		result["certificates"] = certs
		result["certificate_count"] = len(certs)
	}
	return "ct", result, nil
}

func fetchCTIntel(ctx context.Context, domain string) ([]string, []map[string]any, string, error) {
	queryURL := fmt.Sprintf("%s?q=%s&output=json", ctAPIBaseURL, url.QueryEscape("%."+domain))
	body, err := fetchCRTSh(ctx, queryURL)
	if err == nil {
		subdomains, parseErr := parseCTResponse(body, domain)
		certs, certErr := parseCTCertificates(body)
		if parseErr == nil {
			if certErr != nil {
				certs = nil
			}
			return subdomains, certs, "crt.sh", nil
		}
		err = parseErr
	}

	subdomains, fallbackErr := fetchCertSpotter(ctx, domain)
	if fallbackErr == nil && len(subdomains) > 0 {
		return subdomains, nil, "certspotter", nil
	}
	if err != nil {
		return nil, nil, "", fmt.Errorf("crt.sh request: %w", err)
	}
	if fallbackErr != nil {
		return nil, nil, "", fmt.Errorf("certspotter request: %w", fallbackErr)
	}
	return subdomains, nil, "certspotter", nil
}

func fetchCRTSh(ctx context.Context, queryURL string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt < ctMaxFastRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(attempt) * 2 * time.Second
			timer := time.NewTimer(backoff)
			select {
			case <-ctx.Done():
				timer.Stop()
				return nil, ctx.Err()
			case <-timer.C:
			}
		}

		body, _, err := httpGetTimeout(ctx, queryURL, ctResponseMaxBytes, ctHTTPTimeoutFromSettings())
		if err == nil {
			return body, nil
		}
		lastErr = err
		if !isFastRetryableCTError(err) || attempt == ctMaxFastRetries-1 {
			break
		}
	}
	return nil, lastErr
}

// isFastRetryableCTError reports whether crt.sh failed quickly enough to retry
// within the gather budget. Slow client timeouts get a single long attempt only.
func isFastRetryableCTError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "client.timeout exceeded") ||
		strings.Contains(msg, "timeout exceeded while awaiting headers") ||
		strings.Contains(msg, "context deadline exceeded") {
		return false
	}
	return strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "eof") ||
		strings.Contains(msg, "http 502") ||
		strings.Contains(msg, "http 503") ||
		strings.Contains(msg, "http 504") ||
		strings.Contains(msg, "http 429")
}

func ctSearchDomain(host string) string {
	host = strings.TrimSuffix(host, ".")
	parts := strings.Split(host, ".")
	if len(parts) < 2 {
		return ""
	}
	if len(parts) == 2 {
		return strings.Join(parts, ".")
	}
	return parts[len(parts)-2] + "." + parts[len(parts)-1]
}

func parseCTResponse(body []byte, domain string) ([]string, error) {
	var records []ctCertRecord
	if err := json.Unmarshal(body, &records); err != nil {
		return nil, fmt.Errorf("parse crt.sh JSON: %w", err)
	}

	seen := make(map[string]struct{})
	var subdomains []string
	domain = strings.ToLower(domain)

	for _, record := range records {
		for _, rawName := range strings.FieldsFunc(record.NameValue, func(r rune) bool {
			return r == '\n' || r == '\r'
		}) {
			name := normalizeCTName(rawName)
			if name == "" || !isCTSubdomain(name, domain) {
				continue
			}
			if _, ok := seen[name]; ok {
				continue
			}
			seen[name] = struct{}{}
			subdomains = append(subdomains, name)
			if len(subdomains) >= ctMaxSubdomains {
				sort.Strings(subdomains)
				return subdomains, nil
			}
		}
	}

	sort.Strings(subdomains)
	return subdomains, nil
}

func normalizeCTName(name string) string {
	name = strings.TrimSpace(strings.ToLower(name))
	name = strings.TrimPrefix(name, "*.")
	return strings.TrimSuffix(name, ".")
}

func isCTSubdomain(name, domain string) bool {
	return name == domain || strings.HasSuffix(name, "."+domain)
}

func ctHTTPTimeoutFromSettings() time.Duration {
	sec := config.GetSettings().CTHTTPTimeoutSec
	if sec > 0 {
		return time.Duration(sec) * time.Second
	}
	return ctHTTPTimeout
}

func ctGatherTimeoutFromSettings() time.Duration {
	sec := config.GetSettings().CTHTTPTimeoutSec + 15
	if sec > 0 {
		return time.Duration(sec) * time.Second
	}
	return ctGatherTimeout
}
