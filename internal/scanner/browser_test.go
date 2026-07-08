package scanner

import (
	"context"
	"net/url"
	"testing"
)

func TestResolveBrowserWSURL_LocalhostUnchanged(t *testing.T) {
	t.Parallel()

	got := resolveBrowserWSURL(context.Background(), "ws://localhost:3000/")
	if got != "ws://localhost:3000/" {
		t.Fatalf("got %q, want ws://localhost:3000/", got)
	}
}

func TestResolveBrowserWSURL_IPUnchanged(t *testing.T) {
	t.Parallel()

	got := resolveBrowserWSURL(context.Background(), "ws://192.168.1.10:3000/")
	if got != "ws://192.168.1.10:3000/" {
		t.Fatalf("got %q, want ws://192.168.1.10:3000/", got)
	}
}

func TestBrowserHTTPBaseURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   string
		want string
	}{
		{in: "ws://browser:3000/", want: "http://browser:3000"},
		{in: "wss://browserless.example.com/", want: "https://browserless.example.com:3000"},
		{in: "ws://localhost:9222/devtools", want: "http://localhost:9222"},
	}

	for _, tt := range tests {
		if got := browserHTTPBaseURL(tt.in); got != tt.want {
			t.Errorf("browserHTTPBaseURL(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestResolveBrowserWSURL_InvalidURL(t *testing.T) {
	t.Parallel()

	got := resolveBrowserWSURL(context.Background(), "not-a-url")
	if got != "not-a-url" {
		t.Fatalf("got %q, want not-a-url", got)
	}
}

func TestResolveBrowserWSURL_ParsesHost(t *testing.T) {
	t.Parallel()

	raw := "ws://browser:3000/"
	got := resolveBrowserWSURL(context.Background(), raw)

	parsed, err := url.Parse(got)
	if err != nil {
		t.Fatalf("parse resolved url: %v", err)
	}

	if parsed.Port() != "3000" {
		t.Fatalf("port = %q, want 3000", parsed.Port())
	}
}
