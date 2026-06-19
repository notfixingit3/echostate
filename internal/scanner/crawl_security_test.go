package scanner

import "testing"

func TestParseSecurityTxt(t *testing.T) {
	t.Parallel()

	parsed := parseSecurityTxt(`Contact: mailto:security@example.com
Contact: https://example.com/security
Expires: 2030-12-31T00:00:00Z
Canonical: https://example.com/.well-known/security.txt
Policy: https://example.com/security-policy`)

	contacts, ok := parsed["contacts"].([]string)
	if !ok || len(contacts) != 2 {
		t.Fatalf("contacts = %#v", parsed["contacts"])
	}
	if contacts[0] != "security@example.com" {
		t.Fatalf("contact = %q", contacts[0])
	}
	if parsed["canonical"] != "https://example.com/.well-known/security.txt" {
		t.Fatalf("canonical = %#v", parsed["canonical"])
	}
}

func TestParseMTASTS(t *testing.T) {
	t.Parallel()

	parsed := parseMTASTS(`version: STSv1
mode: enforce
max_age: 86400
mx: mail.example.com
mx: *.example.com`)

	if parsed["mode"] != "enforce" {
		t.Fatalf("mode = %#v", parsed["mode"])
	}
	mx, ok := parsed["mx"].([]string)
	if !ok || len(mx) != 2 {
		t.Fatalf("mx = %#v", parsed["mx"])
	}
}

func TestParseTLSRPT(t *testing.T) {
	t.Parallel()

	parsed := parseTLSRPT([]string{`v=TLSRPTv1; rua=mailto:reports@example.com`})
	if parsed["rua"] != "mailto:reports@example.com" {
		t.Fatalf("rua = %#v", parsed["rua"])
	}
}