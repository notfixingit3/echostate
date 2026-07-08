package scanner

import (
	"context"
	"net"
	"net/url"
	"strings"
	"time"
)

// resolveBrowserWSURL pre-resolves Docker service hostnames to an IP so chromedp
// does not rely on its own DNS lookup (which can time out in container networks).
func resolveBrowserWSURL(ctx context.Context, wsURL string) string {
	u, err := url.Parse(wsURL)
	if err != nil {
		return wsURL
	}

	host := u.Hostname()
	if host == "" || host == "localhost" || host == "127.0.0.1" || net.ParseIP(host) != nil {
		return wsURL
	}

	port := u.Port()
	if port == "" {
		port = "3000"
	}

	var ips []string
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		lookupCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		ips, lastErr = net.DefaultResolver.LookupHost(lookupCtx, host)
		cancel()
		if lastErr == nil && len(ips) > 0 {
			u.Host = net.JoinHostPort(ips[0], port)
			return u.String()
		}
		if attempt < 2 {
			time.Sleep(time.Duration(attempt+1) * 250 * time.Millisecond)
		}
	}

	_ = lastErr
	return wsURL
}

// browserHTTPBaseURL returns the browserless HTTP health endpoint for wsURL.
func browserHTTPBaseURL(wsURL string) string {
	u, err := url.Parse(wsURL)
	if err != nil {
		return ""
	}

	scheme := "http"
	if strings.EqualFold(u.Scheme, "wss") {
		scheme = "https"
	}

	host := u.Host
	if u.Port() == "" {
		host = net.JoinHostPort(u.Hostname(), "3000")
	}

	return scheme + "://" + host
}
