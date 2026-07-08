package db

import "testing"

func TestCompareBGPRoutes(t *testing.T) {
	previous := map[string]any{
		"asn": map[string]any{
			"routing": map[string]any{
				"hijack_risk":     "low",
				"visible_origins": []any{"15169"},
				"as_paths":        []any{"3257 15169"},
			},
		},
	}
	current := map[string]any{
		"asn": map[string]any{
			"routing": map[string]any{
				"hijack_risk":     "high",
				"visible_origins": []any{"15169", "174"},
				"as_paths":        []any{"174 15169"},
			},
		},
	}

	diff := compareBGPRoutes(previous, current)
	if !diff.Changed {
		t.Fatal("expected changed diff")
	}
	if diff.HijackRiskFrom != "low" || diff.HijackRiskTo != "high" {
		t.Fatalf("unexpected hijack risk diff: %#v", diff)
	}
	if len(diff.VisibleOriginsAdded) != 1 || diff.VisibleOriginsAdded[0] != "174" {
		t.Fatalf("unexpected visible origins added: %#v", diff.VisibleOriginsAdded)
	}
}

func TestCompareTracerouteRoutes(t *testing.T) {
	previous := map[string]any{
		"traceroute": map[string]any{
			"hops": []any{
				map[string]any{"hop": 1, "ip": "10.0.0.1"},
				map[string]any{"hop": 2, "ip": "93.184.216.34"},
			},
		},
	}
	current := map[string]any{
		"traceroute": map[string]any{
			"hops": []any{
				map[string]any{"hop": 1, "ip": "10.0.0.1"},
				map[string]any{"hop": 2, "ip": "203.0.113.10"},
				map[string]any{"hop": 3, "ip": "93.184.216.34"},
			},
		},
	}

	diff := compareTracerouteRoutes(previous, current)
	if !diff.Changed {
		t.Fatal("expected changed traceroute diff")
	}
	if len(diff.HopsAdded) != 1 || diff.HopsAdded[0] != "203.0.113.10" {
		t.Fatalf("unexpected hops added: %#v", diff.HopsAdded)
	}
}

func TestVisibleOriginRisk(t *testing.T) {
	if visibleOriginRisk("low", false, "valid") != "high" {
		t.Fatal("expected high risk for non-matching origin")
	}
	if visibleOriginRisk("low", true, "valid") != "low" {
		t.Fatal("expected low risk for matching valid origin")
	}
}

func TestGraphEdgeColor(t *testing.T) {
	color := graphEdgeColor("VISIBLE_ORIGIN", map[string]any{"risk": "high"})
	if color != "#ef4444" {
		t.Fatalf("graphEdgeColor() = %q", color)
	}
}
