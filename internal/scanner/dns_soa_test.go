package scanner

import (
	"context"
	"fmt"
	"testing"

	"golang.org/x/net/dns/dnsmessage"
)

func TestSOAResourceToMap(t *testing.T) {
	soa := &dnsmessage.SOAResource{
		NS:      dnsmessage.MustNewName("ns1.example.com."),
		MBox:    dnsmessage.MustNewName("hostmaster.example.com."),
		Serial:  2024010101,
		Refresh: 3600,
		Retry:   600,
		Expire:  86400,
		MinTTL:  300,
	}

	got := soaResourceToMap(soa)
	if got["mname"] != "ns1.example.com" {
		t.Fatalf("mname = %#v", got["mname"])
	}
	if got["rname"] != "hostmaster.example.com" {
		t.Fatalf("rname = %#v", got["rname"])
	}
	if got["serial"] != uint32(2024010101) {
		t.Fatalf("serial = %#v", got["serial"])
	}
	if got["minimum_ttl"] != uint32(300) {
		t.Fatalf("minimum_ttl = %#v", got["minimum_ttl"])
	}
}

func TestLookupSOA_WalksZoneParents(t *testing.T) {
	origQuery := querySOAFunc
	t.Cleanup(func() { querySOAFunc = origQuery })

	calls := make([]string, 0, 2)
	querySOAFunc = func(ctx context.Context, server, zone string) (map[string]any, error) {
		calls = append(calls, zone)
		if zone == "example.com" {
			return map[string]any{
				"mname":  "ns1.example.com",
				"serial": uint32(1),
			}, nil
		}
		return nil, fmt.Errorf("no SOA")
	}

	got, err := lookupSOA(context.Background(), "api.example.com")
	if err != nil {
		t.Fatalf("lookupSOA() error: %v", err)
	}
	if got["zone"] != "example.com" {
		t.Fatalf("zone = %#v, want example.com", got["zone"])
	}
	if len(calls) < 2 || calls[0] != "api.example.com" || calls[1] != "example.com" {
		t.Fatalf("query calls = %#v", calls)
	}
}