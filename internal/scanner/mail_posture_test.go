package scanner

import "testing"

func TestComputeMailPosture(t *testing.T) {
	posture := computeMailPosture(map[string]any{
		"SPF": map[string]any{"policy": "-all"},
		"DMARC_PARSED": map[string]any{
			"policy": "reject",
		},
		"DKIM": []any{
			map[string]any{"selector": "google"},
		},
		"MTA_STS": map[string]any{"mode": "enforce"},
		"TLS_RPT": map[string]any{"rua": "mailto:reports@example.com"},
	})

	if posture["grade"] != "A" {
		t.Fatalf("grade = %v, want A", posture["grade"])
	}
	if intValMap(posture, "score") < 90 {
		t.Fatalf("score = %v, want >= 90", posture["score"])
	}
}

func TestComputeMailPostureBIMI(t *testing.T) {
	withBIMI := computeMailPosture(map[string]any{
		"SPF": map[string]any{"policy": "-all"},
		"DMARC_PARSED": map[string]any{
			"policy": "reject",
		},
		"DKIM": []any{
			map[string]any{"selector": "google"},
		},
		"BIMI": map[string]any{"record": "v=BIMI1; l=https://example.com/logo.svg"},
	})
	withoutBIMI := computeMailPosture(map[string]any{
		"SPF": map[string]any{"policy": "-all"},
		"DMARC_PARSED": map[string]any{
			"policy": "reject",
		},
		"DKIM": []any{
			map[string]any{"selector": "google"},
		},
	})

	if intValMap(withBIMI, "score") <= intValMap(withoutBIMI, "score") {
		t.Fatalf("BIMI should increase score: with=%v without=%v", withBIMI["score"], withoutBIMI["score"])
	}
}

func TestComputeMailPostureWeak(t *testing.T) {
	posture := computeMailPosture(map[string]any{
		"SPF": map[string]any{"policy": "+all"},
		"DMARC_PARSED": map[string]any{
			"policy": "none",
		},
	})

	if posture["grade"] != "F" {
		t.Fatalf("grade = %v, want F", posture["grade"])
	}
}

func intValMap(m map[string]any, key string) int {
	switch v := m[key].(type) {
	case int:
		return v
	case float64:
		return int(v)
	default:
		return 0
	}
}
