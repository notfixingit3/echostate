package scanner

import (
	"reflect"
	"testing"

	"github.com/chromedp/cdproto/network"
)

func TestHeadersFromNetwork(t *testing.T) {
	got := headersFromNetwork(network.Headers{
		"Server":       "nginx/1.24.0",
		"Set-Cookie":   []any{"a=1", "b=2"},
		"X-Test-Count": 42,
	})

	want := map[string]string{
		"Server":       "nginx/1.24.0",
		"Set-Cookie":   "a=1, b=2",
		"X-Test-Count": "42",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("headersFromNetwork() = %#v, want %#v", got, want)
	}
}

func TestExtractSecurityHeaders(t *testing.T) {
	got := extractSecurityHeaders(map[string]string{
		"Server":                    "nginx",
		"Strict-Transport-Security": "max-age=31536000",
		"Content-Security-Policy":   "default-src 'self'",
		"X-Frame-Options":           "SAMEORIGIN",
	})

	want := map[string]string{
		"strict-transport-security": "max-age=31536000",
		"content-security-policy":   "default-src 'self'",
		"x-frame-options":           "SAMEORIGIN",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("extractSecurityHeaders() = %#v, want %#v", got, want)
	}
}

func TestExtractTechStackFromHeaders(t *testing.T) {
	got := extractTechStackFromHeaders(map[string]string{
		"Server":           "nginx/1.24.0",
		"X-Powered-By":     "PHP/8.2.0",
		"X-AspNet-Version": "4.0.30319",
		"CF-RAY":           "abc123-SJC",
		"Via":              "1.1 varnish",
		"Content-Type":     "text/html",
	})

	want := map[string]struct{}{
		"Server: nginx/1.24.0":       {},
		"X-Powered-By: PHP/8.2.0":    {},
		"X-AspNet-Version: 4.0.30319": {},
		"CDN: Cloudflare":            {},
		"Via: 1.1 varnish":           {},
	}

	if len(got) != len(want) {
		t.Fatalf("extractTechStackFromHeaders() = %#v, want %d entries", got, len(want))
	}
	for _, item := range got {
		if _, ok := want[item]; !ok {
			t.Fatalf("extractTechStackFromHeaders() unexpected entry %q in %#v", item, got)
		}
	}
}

func TestMergeTechStack(t *testing.T) {
	got := mergeTechStack(
		[]string{"Server: nginx", "X-Powered-By: PHP/8.2"},
		[]string{"generator: WordPress 6.4", "Server: nginx"},
		[]string{"cms: WordPress"},
	)

	want := []string{
		"Server: nginx",
		"X-Powered-By: PHP/8.2",
		"generator: WordPress 6.4",
		"cms: WordPress",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("mergeTechStack() = %#v, want %#v", got, want)
	}
}