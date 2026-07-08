package scanner

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/notfixingit3/echostate/internal/config"
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

func TestRedactScannerTraceroutePrefix_PrivateAndISP(t *testing.T) {
	enabled := true
	config.UpdateSettings(config.SystemSettings{TracerouteRedactScannerPrefix: &enabled})

	hops := []map[string]any{
		{"hop": 1, "ip": "192.168.1.1", "rtt_ms": 1.0},
		{"hop": 2, "ip": "203.0.113.10", "rtt_ms": 5.0},
		{"hop": 3, "ip": "93.184.216.34", "rtt_ms": 10.0},
	}

	trimmed, redacted := redactScannerTraceroutePrefix(hops)
	if redacted != 2 {
		t.Fatalf("redacted = %d, want 2", redacted)
	}
	if len(trimmed) != 1 {
		t.Fatalf("len(trimmed) = %d, want 1", len(trimmed))
	}
	if trimmed[0]["hop"] != 1 || trimmed[0]["ip"] != "93.184.216.34" {
		t.Fatalf("unexpected trimmed hop: %#v", trimmed[0])
	}
}

func TestRedactScannerTraceroutePrefix_Disabled(t *testing.T) {
	disabled := false
	config.UpdateSettings(config.SystemSettings{TracerouteRedactScannerPrefix: &disabled})

	hops := []map[string]any{
		{"hop": 1, "ip": "192.168.1.1"},
		{"hop": 2, "ip": "93.184.216.34"},
	}

	trimmed, redacted := redactScannerTraceroutePrefix(hops)
	if redacted != 0 || len(trimmed) != 2 {
		t.Fatalf("expected unchanged hops, got redacted=%d len=%d", redacted, len(trimmed))
	}
}

func TestRedactScannerTraceroutePrefix_ShortPathKeepsDestination(t *testing.T) {
	enabled := true
	config.UpdateSettings(config.SystemSettings{TracerouteRedactScannerPrefix: &enabled})

	hops := []map[string]any{
		{"hop": 1, "ip": "192.168.1.1"},
		{"hop": 2, "ip": "93.184.216.34"},
	}

	trimmed, redacted := redactScannerTraceroutePrefix(hops)
	if redacted != 1 {
		t.Fatalf("redacted = %d, want 1", redacted)
	}
	if len(trimmed) != 1 || trimmed[0]["ip"] != "93.184.216.34" {
		t.Fatalf("unexpected trimmed hops: %#v", trimmed)
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
	if len(hops) != 1 {
		t.Fatalf("len(hops) = %d, want 1", len(hops))
	}
	if hops[0]["ip"] != "93.184.216.34" {
		t.Fatalf("unexpected destination hop: %#v", hops[0])
	}
	if data["local_prefix_redacted"] != 1 {
		t.Fatalf("local_prefix_redacted = %#v, want 1", data["local_prefix_redacted"])
	}
}

func TestGatherTraceroute_ExternalUnavailableSkipped(t *testing.T) {
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
		return "", errExternalTracerouteUnavailable
	}

	_, data, err := gatherTraceroute(context.Background(), "example.com")
	if err != nil {
		t.Fatalf("gatherTraceroute() error = %v", err)
	}

	vantages, ok := data["vantages"].([]map[string]any)
	if !ok || len(vantages) != 2 {
		t.Fatalf("expected 2 vantages, got %#v", data["vantages"])
	}
	if vantages[1]["skipped"] == nil {
		t.Fatalf("expected external vantage skipped, got %#v", vantages[1])
	}
	if warning := data["warning"]; warning != nil {
		t.Fatalf("expected no traceroute warning when external API is unavailable, got %#v", warning)
	}
}

func TestFormatTracerouteVantageWarning(t *testing.T) {
	got := formatTracerouteVantageWarning("external", fmt.Errorf("external traceroute: HTTP 404"))
	if got != "external: HTTP 404" {
		t.Fatalf("formatTracerouteVantageWarning() = %q", got)
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
