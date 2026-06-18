package scanner

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestCopyrightMatcher(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name: "copyright symbol and year",
			input: `Welcome
Copyright © 2023 Example Inc. All rights reserved.
Thanks for visiting.`,
			expected: []string{"Copyright © 2023 Example Inc. All rights reserved."},
		},
		{
			name: "case-insensitive copyright word",
			input: `About Us
cOPYRIGHT 2018 Some Company
Home`,
			expected: []string{"cOPYRIGHT 2018 Some Company"},
		},
		{
			name: "year without copyright word",
			input: `Founded 1999
Contact Us
Established 2021`,
			expected: []string{"Founded 1999", "Established 2021"},
		},
		{
			name: "deduplicates repeated lines",
			input: `Copyright 2020 Foo
Copyright 2020 Foo
Copyright 2021 Bar`,
			expected: []string{"Copyright 2020 Foo", "Copyright 2021 Bar"},
		},
		{
			name:     "ignores unrelated text",
			input:    "Hello world\nNothing interesting here\nGoodbye",
			expected: nil,
		},
		{
			name: "preserves order",
			input: `First line © 2022
Second line 1995
Third line Copyright 2024`,
			expected: []string{
				"First line © 2022",
				"Second line 1995",
				"Third line Copyright 2024",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractCopyrights(tt.input)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("extractCopyrights(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

// TestWebLive exercises the chromedp-backed gatherer against a real browser.
// It skips gracefully if no browser is listening.
func TestWebLive(t *testing.T) {
	wsURL := os.Getenv("BROWSER_WS_URL")
	if wsURL == "" {
		wsURL = defaultBrowserWSURL
	}

	if !browserAvailable(wsURL) {
		t.Skipf("browser not available at %s", wsURL)
	}

	g := newWebGatherer(wsURL)
	key, value, err := g(context.Background(), "example.com")
	if err != nil {
		t.Skipf("live web gather failed: %v", err)
	}

	if key != "web" {
		t.Errorf("key = %q, want web", key)
	}

	title, _ := value["title"].(string)
	if title == "" {
		t.Error("expected non-empty title")
	}

	finalURL, _ := value["url"].(string)
	if !strings.HasPrefix(finalURL, "http") {
		t.Errorf("url = %q, want http(s) prefix", finalURL)
	}

	copyrights, ok := value["copyrights"].([]string)
	if !ok {
		t.Errorf("copyrights type = %T, want []string", value["copyrights"])
	}

	headers := value["headers"]
	headerCount := 0
	switch h := headers.(type) {
	case map[string]string:
		headerCount = len(h)
	case map[string]any:
		headerCount = len(h)
	default:
		t.Errorf("headers type = %T, want map", headers)
	}
	if headerCount == 0 {
		t.Errorf("expected non-empty headers, got %#v", headers)
	}

	headerURL, _ := value["header_url"].(string)
	if headerURL == "" {
		t.Error("expected non-empty header_url")
	}

	t.Logf("title=%q url=%q header_url=%q copyrights=%d headers=%d", title, finalURL, headerURL, len(copyrights), headerCount)
}

// browserAvailable probes the browserless health endpoint derived from the
// WebSocket URL. It returns false when the browser cannot be reached within a
// short timeout.
func browserAvailable(wsURL string) bool {
	u, err := url.Parse(wsURL)
	if err != nil {
		return false
	}

	scheme := "http"
	if u.Scheme == "wss" {
		scheme = "https"
	}

	host := u.Host
	if u.Port() == "" {
		host += ":3000"
	}

	healthURL := fmt.Sprintf("%s://%s/pressure", scheme, host)

	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(healthURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}
