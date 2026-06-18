package scanner

import "testing"

func TestBuildJA3SClientHello(t *testing.T) {
	payload := buildJA3SClientHello("example.com")
	if len(payload) < 100 {
		t.Fatalf("payload too short: %d bytes", len(payload))
	}
	if payload[0] != 0x16 {
		t.Fatalf("record type = %#x, want 0x16", payload[0])
	}
}