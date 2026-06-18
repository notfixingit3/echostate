package scanner

import "testing"

func TestParseSPFRecord(t *testing.T) {
	got := parseSPFRecord([]string{"v=spf1 include:_spf.google.com ~all"})
	if got == nil {
		t.Fatal("expected SPF parse result")
	}
	if got["policy"] != "~all" {
		t.Fatalf("policy = %#v, want ~all", got["policy"])
	}
}

func TestParseDMARCRecord(t *testing.T) {
	got := parseDMARCRecord([]string{"v=DMARC1; p=reject; sp=quarantine; pct=100; rua=mailto:dmarc@example.com"})
	if got == nil {
		t.Fatal("expected DMARC parse result")
	}
	if got["policy"] != "reject" {
		t.Fatalf("policy = %#v, want reject", got["policy"])
	}
	if got["subdomain_policy"] != "quarantine" {
		t.Fatalf("subdomain_policy = %#v, want quarantine", got["subdomain_policy"])
	}
}