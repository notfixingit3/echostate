package diff

import (
	"testing"

	"github.com/notfixingit3/echostate/internal/models"
)

func TestCompute_CTSubdomain(t *testing.T) {
	previous := map[string]any{
		"ct": map[string]any{"subdomains": []any{"www.example.com"}},
	}
	current := &models.ScanResult{
		CT: map[string]any{"subdomains": []any{"www.example.com", "api.example.com"}},
	}
	entries := Compute(previous, current)
	if len(entries) != 1 || entries[0].Type != "new_ct_subdomain" {
		t.Fatalf("Compute() = %#v", entries)
	}
}

func TestCompute_WordPressPlugins(t *testing.T) {
	previous := map[string]any{
		"web": map[string]any{
			"wordpress_plugins": []any{
				map[string]any{"slug": "akismet", "version": "5.3"},
				map[string]any{"slug": "woocommerce", "version": "8.8.0"},
			},
		},
	}
	current := &models.ScanResult{
		Web: map[string]any{
			"wordpress_plugins": []any{
				map[string]any{"slug": "woocommerce", "version": "8.9.1"},
				map[string]any{"slug": "jetpack"},
			},
		},
	}

	entries := Compute(previous, current)
	if len(entries) != 3 {
		t.Fatalf("Compute() = %#v", entries)
	}
}

func TestCompute_SOAMnameChange(t *testing.T) {
	previous := map[string]any{
		"dns": map[string]any{
			"SOA": map[string]any{
				"zone":   "example.com",
				"mname":  "ns1.example.com",
				"serial": 100,
			},
		},
	}
	current := &models.ScanResult{
		DNS: map[string]any{
			"SOA": map[string]any{
				"zone":   "example.com",
				"mname":  "ns2.example.com",
				"serial": 101,
			},
		},
	}

	entries := Compute(previous, current)
	if len(entries) != 2 {
		t.Fatalf("Compute() = %#v", entries)
	}
	if entries[0].Type != "soa_mname" || entries[0].Severity != SeverityWarning {
		t.Fatalf("mname entry = %#v", entries[0])
	}
	if entries[1].Type != "soa_serial" {
		t.Fatalf("serial entry = %#v", entries[1])
	}
}

func TestCompute_CertExpiryWarning(t *testing.T) {
	previous := map[string]any{
		"tls": map[string]any{"days_remaining": 45},
	}
	current := &models.ScanResult{
		TLS: map[string]any{"days_remaining": 20},
	}
	entries := Compute(previous, current)
	if len(entries) != 1 || entries[0].Severity != SeverityWarning {
		t.Fatalf("Compute() = %#v", entries)
	}
}
