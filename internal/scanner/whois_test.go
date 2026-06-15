package scanner

import (
	"context"
	"reflect"
	"strings"
	"testing"
)

func TestWHOIS(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live network test in short mode")
	}

	_, value, err := gatherWHOIS(context.Background(), "example.com")
	if err != nil {
		t.Fatalf("gatherWHOIS returned error: %v", err)
	}

	if value["domain"] != "example.com" {
		t.Errorf("domain = %v, want example.com", value["domain"])
	}

	raw, ok := value["raw"].(string)
	if !ok || raw == "" {
		t.Error("expected non-empty raw whois text")
	}

	if len(raw) > maxWHOISRaw {
		t.Errorf("raw length %d exceeds max %d", len(raw), maxWHOISRaw)
	}
}

func TestWHOISError(t *testing.T) {
	orig := whoisLookup
	whoisLookup = func(domain string, servers ...string) (string, error) {
		return "", context.DeadlineExceeded
	}
	defer func() { whoisLookup = orig }()

	_, _, err := gatherWHOIS(context.Background(), "this-is-not-a-real-domain-12345.invalidtld")
	if err == nil {
		t.Fatal("expected error for invalid/unresolvable domain")
	}
}

func TestWHOISParse(t *testing.T) {
	validRaw := `Domain Name: example.com
Registrar: Example Registrar Inc.
Domain Status: clientTransferProhibited https://icann.org/epp#clientTransferProhibited
Name Server: ns1.example.com
Name Server: ns2.example.com
Creation Date: 2001-01-01T00:00:00Z
Expiration Date: 2025-12-31T00:00:00Z
`

	padding := strings.Repeat("% padding line\n", 900)
	largeRaw := padding + validRaw

	tests := []struct {
		name    string
		raw     string
		wantErr bool
		check   func(t *testing.T, value map[string]any)
	}{
		{
			name:    "valid parse",
			raw:     validRaw,
			wantErr: false,
			check: func(t *testing.T, value map[string]any) {
				if got, want := value["domain"], "example.com"; got != want {
					t.Errorf("domain = %v, want %v", got, want)
				}
				if got, want := value["registrar"], "Example Registrar Inc."; got != want {
					t.Errorf("registrar = %v, want %v", got, want)
				}
				if got, want := value["expiration_date"], "2025-12-31T00:00:00Z"; got != want {
					t.Errorf("expiration_date = %v, want %v", got, want)
				}
				ns, _ := value["name_servers"].([]string)
				if want := []string{"ns1.example.com", "ns2.example.com"}; !reflect.DeepEqual(ns, want) {
					t.Errorf("name_servers = %v, want %v", ns, want)
				}
				status, _ := value["status"].([]string)
				if len(status) != 1 || status[0] != "clientTransferProhibited" {
					t.Errorf("status = %v, want [clientTransferProhibited]", status)
				}
			},
		},
		{
			name:    "truncation",
			raw:     largeRaw,
			wantErr: false,
			check: func(t *testing.T, value map[string]any) {
				raw, _ := value["raw"].(string)
				if len(raw) != maxWHOISRaw {
					t.Errorf("raw length = %d, want %d", len(raw), maxWHOISRaw)
				}
				if value["domain"] != "example.com" {
					t.Errorf("domain = %v, want example.com", value["domain"])
				}
			},
		},
		{
			name:    "parse failure",
			raw:     "This is not whois data.",
			wantErr: true,
			check: func(t *testing.T, value map[string]any) {
				raw, _ := value["raw"].(string)
				if raw != "This is not whois data." {
					t.Errorf("raw = %q, want original raw text on parse failure", raw)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orig := whoisLookup
			whoisLookup = func(domain string, servers ...string) (string, error) {
				return tt.raw, nil
			}
			defer func() { whoisLookup = orig }()

			_, value, err := gatherWHOIS(context.Background(), "example.com")
			if (err != nil) != tt.wantErr {
				t.Fatalf("gatherWHOIS error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.check != nil {
				tt.check(t, value)
			}
		})
	}
}
