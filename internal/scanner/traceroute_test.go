package scanner

import (
	"context"
	"strings"
	"testing"
)

func TestParseTracerouteOutput(t *testing.T) {
	output := strings.TrimSpace(`
traceroute to example.com (93.184.216.34), 20 hops max, 46 byte packets
 1  192.168.1.1  1.234 ms
 2  10.0.0.1  2.345 ms
 3  * * *
 4  93.184.216.34  18.765 ms
`)

	hops := parseTracerouteOutput(output)
	if len(hops) != 4 {
		t.Fatalf("parseTracerouteOutput() len = %d, want 4", len(hops))
	}

	if hops[0]["hop"] != 1 || hops[0]["ip"] != "192.168.1.1" {
		t.Fatalf("unexpected hop 1: %#v", hops[0])
	}
	if hops[2]["timeout"] != true {
		t.Fatalf("expected hop 3 timeout, got %#v", hops[2])
	}
	if hops[3]["ip"] != "93.184.216.34" {
		t.Fatalf("unexpected final hop: %#v", hops[3])
	}
}

func TestParseTracerouteLine_HostnameFormat(t *testing.T) {
	hop, ok := parseTracerouteLine(" 5  core-router (203.0.113.10)  12.345 ms")
	if !ok {
		t.Fatal("expected parse ok")
	}
	if hop.Hop != 5 || hop.IP != "203.0.113.10" || hop.RTTMs != 12.345 {
		t.Fatalf("unexpected hop: %#v", hop)
	}
}

func TestGatherTraceroute_SkipsIP(t *testing.T) {
	_, data, err := gatherTraceroute(context.Background(), "93.184.216.34")
	if err != nil {
		t.Fatalf("gatherTraceroute() error = %v", err)
	}
	if data["skipped"] == nil {
		t.Fatalf("expected skipped message, got %#v", data)
	}
}

func TestGatherTraceroute_FromMockOutput(t *testing.T) {
	orig := runTracerouteFunc
	defer func() { runTracerouteFunc = orig }()

	runTracerouteFunc = func(ctx context.Context, host string) (string, error) {
		return `traceroute to example.com (93.184.216.34), 20 hops max
 1  192.168.1.1  1.000 ms
 2  93.184.216.34  10.000 ms`, nil
	}

	_, data, err := gatherTraceroute(context.Background(), "example.com")
	if err != nil {
		t.Fatalf("gatherTraceroute() error = %v", err)
	}
	hops, ok := data["hops"].([]map[string]any)
	if !ok {
		t.Fatalf("expected hops slice, got %#v", data["hops"])
	}
	if len(hops) != 2 {
		t.Fatalf("len(hops) = %d, want 2", len(hops))
	}
}

func TestGatherTraceroute_MultiVantage(t *testing.T) {
	origLocal := runTracerouteFunc
	origExternal := fetchExternalTraceFunc
	defer func() {
		runTracerouteFunc = origLocal
		fetchExternalTraceFunc = origExternal
	}()

	runTracerouteFunc = func(ctx context.Context, host string) (string, error) {
		return `traceroute to example.com
 1  192.168.1.1  1.000 ms
 2  93.184.216.34  10.000 ms`, nil
	}
	fetchExternalTraceFunc = func(ctx context.Context, host string) (string, error) {
		return `traceroute to example.com
 1  198.51.100.1  2.000 ms
 2  93.184.216.34  12.000 ms`, nil
	}

	_, data, err := gatherTraceroute(context.Background(), "example.com")
	if err != nil {
		t.Fatalf("gatherTraceroute() error = %v", err)
	}

	vantages, ok := data["vantages"].([]map[string]any)
	if !ok {
		t.Fatalf("expected vantages slice, got %#v", data["vantages"])
	}
	if len(vantages) != 2 {
		t.Fatalf("len(vantages) = %d, want 2", len(vantages))
	}
}