package enrichment

import (
	"testing"

	"github.com/notfixingit3/echostate/internal/config"
)

func TestDiffEnrichmentAdded(t *testing.T) {
	t.Parallel()

	entries := DiffEnrichment(nil, map[string]any{
		"status": "completed",
		"shodan": map[string]any{
			"host": map[string]any{"ip": "1.2.3.4", "org": "Example Org"},
		},
	})
	if len(entries) != 1 {
		t.Fatalf("entries = %#v", entries)
	}
	if entries[0].Type != "enrichment_shodan_added" {
		t.Fatalf("type = %s", entries[0].Type)
	}
}

func TestDiffEnrichmentHIBPBreach(t *testing.T) {
	t.Parallel()

	entries := DiffEnrichment(
		map[string]any{
			"hibp": map[string]any{
				"results": []any{
					map[string]any{"email": "clean@example.com", "pwned": false},
				},
			},
		},
		map[string]any{
			"hibp": map[string]any{
				"results": []any{
					map[string]any{"email": "clean@example.com", "pwned": false},
					map[string]any{"email": "pwned@example.com", "pwned": true},
				},
			},
		},
	)

	var found bool
	for _, entry := range entries {
		if entry.Type == "enrichment_hibp_breach" {
			found = true
		}
	}
	if !found {
		t.Fatalf("entries = %#v", entries)
	}
}

func TestShouldEnqueueEnrichment(t *testing.T) {
	t.Parallel()

	if !ShouldEnqueueEnrichment(config.SystemSettings{}, "example.com") {
		t.Fatal("expected wayback-capable domain to enqueue")
	}
	if ShouldEnqueueEnrichment(config.SystemSettings{}, "1.2.3.4") {
		t.Fatal("did not expect IP-only target without API keys")
	}
}
