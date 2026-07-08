package diff

import (
	"encoding/json"
	"testing"

	"github.com/notfixingit3/echostate/internal/models"
)

func TestDiffGraph_BGPDrift(t *testing.T) {
	previous := map[string]any{
		"asn": map[string]any{
			"routing": map[string]any{
				"visible_origins": []any{"15169"},
				"as_paths":        []any{"15169 174"},
				"rpki":            map[string]any{"overall": "valid"},
			},
		},
	}
	current := &models.ScanResult{
		ASN: map[string]any{
			"routing": map[string]any{
				"visible_origins": []any{"15169", "64512"},
				"as_paths":        []any{"64512 174"},
				"rpki":            map[string]any{"overall": "invalid"},
			},
		},
	}

	entries := diffGraph(previous, mustMap(current))
	types := entryTypes(entries)
	for _, want := range []string{"graph_bgp_origin_added", "graph_as_path_added", "graph_as_path_removed", "graph_rpki_changed"} {
		if !contains(types, want) {
			t.Fatalf("types %v missing %q", types, want)
		}
	}
}

func TestDiffGraph_TracerouteDrift(t *testing.T) {
	previous := map[string]any{
		"traceroute": map[string]any{
			"hops": []any{
				map[string]any{"hop": 1, "ip": "10.0.0.1"},
				map[string]any{"hop": 2, "ip": "10.0.0.2"},
			},
		},
	}
	current := &models.ScanResult{
		Traceroute: map[string]any{
			"hops": []any{
				map[string]any{"hop": 1, "ip": "10.0.0.1"},
				map[string]any{"hop": 2, "ip": "10.0.0.9"},
			},
		},
	}

	entries := diffGraph(previous, mustMap(current))
	types := entryTypes(entries)
	if !contains(types, "graph_traceroute_hop_added") || !contains(types, "graph_traceroute_hop_removed") {
		t.Fatalf("diffGraph() = %#v", entries)
	}
}

func mustMap(result *models.ScanResult) map[string]any {
	data, err := json.Marshal(result)
	if err != nil {
		panic(err)
	}
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		panic(err)
	}
	return out
}

func entryTypes(entries []Entry) []string {
	var out []string
	for _, e := range entries {
		out = append(out, e.Type)
	}
	return out
}

func contains(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}
