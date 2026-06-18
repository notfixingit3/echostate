package db

import (
	"testing"

	"github.com/notfixingit3/echostate/internal/models"
)

func TestBuildVantageDivergence(t *testing.T) {
	paths := []models.GraphPath{
		{
			TargetID:     "t1",
			TargetLabel:  "example.com",
			Vantage:      "local",
			VantageLabel: "Local scanner",
			Hops: []models.GraphPathHop{
				{Hop: 1, IP: "10.0.0.1"},
				{Hop: 2, IP: "93.184.216.34"},
			},
		},
		{
			TargetID:     "t1",
			TargetLabel:  "example.com",
			Vantage:      "external",
			VantageLabel: "External vantage",
			Hops: []models.GraphPathHop{
				{Hop: 1, IP: "198.51.100.1"},
				{Hop: 2, IP: "93.184.216.34"},
			},
		},
	}

	divergence := buildVantageDivergence(paths)
	if len(divergence) != 1 {
		t.Fatalf("len(divergence) = %d, want 1", len(divergence))
	}
	if divergence[0].DivergesAt != 1 {
		t.Fatalf("DivergesAt = %d, want 1", divergence[0].DivergesAt)
	}
	if len(divergence[0].OnlyInA) == 0 || len(divergence[0].OnlyInB) == 0 {
		t.Fatalf("expected unique hop IPs, got %#v", divergence[0])
	}
}