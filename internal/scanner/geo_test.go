package scanner

import (
	"context"
	"testing"
)

func TestEnrichHopGeo_UsesCache(t *testing.T) {
	orig := lookupGeoFunc
	defer func() { lookupGeoFunc = orig }()

	calls := 0
	lookupGeoFunc = func(ctx context.Context, ip string) (geoResult, error) {
		calls++
		return geoResult{
			Country:   "US",
			Latitude:  37.7,
			Longitude: -97.8,
		}, nil
	}

	hops := []map[string]any{
		{"hop": 1, "ip": "8.8.8.8"},
		{"hop": 2, "ip": "8.8.8.8"},
	}
	enrichHopGeo(context.Background(), hops)

	if calls != 1 {
		t.Fatalf("lookup calls = %d, want 1", calls)
	}
	if hops[1]["country"] != "US" {
		t.Fatalf("expected cached geo on hop 2, got %#v", hops[1])
	}
}

func TestLookupGeo_SkipsPrivateIP(t *testing.T) {
	_, err := lookupGeo(context.Background(), "192.168.1.1")
	if err == nil {
		t.Fatal("expected error for private ip")
	}
}