package db

import "testing"

func TestParseMXHost(t *testing.T) {
	got := parseMXHost("10 mail.example.com.")
	if got != "mail.example.com" {
		t.Fatalf("parseMXHost() = %q", got)
	}
}

func TestNormalizeDNSHost(t *testing.T) {
	got := normalizeDNSHost("ns1.example.com.")
	if got != "ns1.example.com" {
		t.Fatalf("normalizeDNSHost() = %q", got)
	}
}