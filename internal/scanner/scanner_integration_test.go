//go:build integration

package scanner

import (
	"context"
	"net"
	"os"
	"strings"
	"testing"
	"time"
)

// TestScannerIntegration performs a live, end-to-end run of the scanner against
// example.com. It requires outbound network access for WHOIS and ASN lookups,
// and an optional browserless/chrome endpoint for web data. If the network is
// unavailable, the test skips. If the browser endpoint is unavailable, the web
// gatherer is expected to report an error gracefully rather than fail the scan.
func TestScannerIntegration(t *testing.T) {
	if os.Getenv("ECHOSTATE_INTEGRATION") == "skip" {
		t.Skip("ECHOSTATE_INTEGRATION=skip set")
	}

	if !networkAvailable() {
		t.Skip("network unavailable; skipping integration test")
	}

	browserWSURL := os.Getenv("BROWSER_WS_URL")
	if browserWSURL == "" {
		browserWSURL = "ws://localhost:3000/"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	s := NewScanner(browserWSURL)
	result, err := s.Run(ctx, "example.com")
	if err != nil {
		t.Fatalf("scanner.Run failed: %v", err)
	}

	if result.Host != "example.com" {
		t.Errorf("expected host example.com, got %q", result.Host)
	}

	// At least one of WHOIS or ASN should succeed when network is available.
	if len(result.WHOIS) == 0 && len(result.ASN) == 0 {
		t.Error("expected WHOIS or ASN data, both are empty")
	}

	// Web may fail if no browser is available; that's acceptable, but we should
	// not see a panic or fatal error in result.Errors.
	for _, e := range result.Errors {
		if strings.Contains(e, "panic") || strings.Contains(e, "runtime error") {
			t.Errorf("unexpected fatal error: %s", e)
		}
	}

	t.Logf("WHOIS: %v", result.WHOIS)
	t.Logf("ASN: %v", result.ASN)
	t.Logf("Web: %v", result.Web)
	t.Logf("Errors: %v", result.Errors)
}

// networkAvailable returns true if we can resolve a well-known public DNS name.
func networkAvailable() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resolver := &net.Resolver{}
	_, err := resolver.LookupHost(ctx, "example.com")
	return err == nil
}
