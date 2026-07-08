package scanner

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchPeeringDB(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/net":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{
					{
						"id":   42,
						"name": "Example Networks",
						"ix": []map[string]any{
							{"id": 1, "name": "AMS-IX", "country": "NL", "city": "Amsterdam"},
						},
						"netixlan": []map[string]any{
							{
								"ix_id":   1,
								"speed":   10000,
								"ipaddr4": "203.0.113.10",
							},
						},
					},
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	origBase := peeringDBBaseURLVar
	peeringDBBaseURLVar = server.URL + "/api"
	defer func() { peeringDBBaseURLVar = origBase }()

	result, err := fetchPeeringDB(context.Background(), "64496")
	if err != nil {
		t.Fatalf("fetchPeeringDB() error = %v", err)
	}
	if result["ix_count"] != 1 {
		t.Fatalf("ix_count = %#v, want 1", result["ix_count"])
	}
	ixlan, ok := result["ixlan"].([]map[string]any)
	if !ok || len(ixlan) != 1 {
		t.Fatalf("unexpected ixlan: %#v", result["ixlan"])
	}
	if ixlan[0]["ix_name"] != "AMS-IX" {
		t.Fatalf("ix_name = %#v", ixlan[0]["ix_name"])
	}
}
