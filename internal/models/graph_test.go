package models

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestGraphResponseEmptyEdgesEncodesAsArray(t *testing.T) {
	resp := GraphResponse{
		View:  "ct",
		Nodes: []GraphNode{},
		Edges: make([]GraphEdge, 0),
		Stats: map[string]int{},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if strings.Contains(string(data), `"edges":null`) {
		t.Fatalf("expected empty edges array, got %s", data)
	}
}
