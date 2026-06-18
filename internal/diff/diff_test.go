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