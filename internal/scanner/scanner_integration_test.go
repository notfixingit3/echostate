//go:build integration

package scanner

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestScannerIntegration performs a live, end-to-end run of the scanner against
// example.com. It requires outbound network access for WHOIS and ASN lookups,
// and an optional browserless/chrome endpoint for web data. If the browser
// endpoint is unavailable, the web gatherer is expected to report an error
// gracefully rather than fail the scan.
func TestScannerIntegration(t *testing.T) {
	if os.Getenv("ECHOSTATE_INTEGRATION") == "skip" {
		t.Skip("ECHOSTATE_INTEGRATION=skip set")
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
	hasFatal := false
	for _, e := range result.Errors {
		if e == "" {
			continue
		}
		// Treat runtime panics / unexpected fatals as failures.
		if containsAny(e, []string{"panic", "runtime error", "index out of range"}) {
			hasFatal = true
		}
	}
	if hasFatal {
		t.Errorf("unexpected fatal errors in result.Errors: %v", result.Errors)
	}

	t.Logf("WHOIS: %v", result.WHOIS)
	t.Logf("ASN: %v", result.ASN)
	t.Logf("Web: %v", result.Web)
	t.Logf("Errors: %v", result.Errors)
}

func containsAny(s string, subs []string) bool {
	for _, sub := range subs {
		if contains(s, sub) {
			return true
		}
	}
	return false
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) > 0 && containsImpl(s, sub))
}

func containsImpl(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
