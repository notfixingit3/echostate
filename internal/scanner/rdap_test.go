package scanner

import "testing"

func TestParseRDAPDomain(t *testing.T) {
	t.Parallel()

	payload := map[string]any{
		"ldhName": "EXAMPLE.COM",
		"status":  []any{"client transfer prohibited"},
		"events": []any{
			map[string]any{"eventAction": "registration", "eventDate": "2001-01-01T00:00:00Z"},
			map[string]any{"eventAction": "expiration", "eventDate": "2030-12-31T00:00:00Z"},
		},
		"nameservers": []any{
			map[string]any{"ldhName": "ns1.example.com."},
			map[string]any{"ldhName": "ns2.example.com."},
		},
		"entities": []any{
			map[string]any{
				"roles": []any{"registrar"},
				"vcardArray": []any{
					"vcard",
					[]any{
						[]any{"fn", map[string]any{}, "text", "Example Registrar Inc."},
						[]any{"email", map[string]any{}, "text", "mailto:registrar@example.com"},
					},
				},
			},
			map[string]any{
				"roles": []any{"registrant"},
				"vcardArray": []any{
					"vcard",
					[]any{
						[]any{"org", map[string]any{}, "text", "Example Org"},
						[]any{"email", map[string]any{}, "text", "mailto:owner@example.com"},
					},
				},
			},
			map[string]any{
				"roles": []any{"abuse"},
				"vcardArray": []any{
					"vcard",
					[]any{
						[]any{"email", map[string]any{}, "text", "mailto:abuse@example.com"},
					},
				},
			},
		},
	}

	result := parseRDAPDomain(payload)
	if result == nil {
		t.Fatal("parseRDAPDomain returned nil")
	}
	if result["domain"] != "example.com" {
		t.Fatalf("domain = %v", result["domain"])
	}
	if result["registrar"] != "Example Registrar Inc." {
		t.Fatalf("registrar = %v", result["registrar"])
	}
	if result["registrant_email"] != "owner@example.com" {
		t.Fatalf("registrant_email = %v", result["registrant_email"])
	}
	if result["abuse_email"] != "abuse@example.com" {
		t.Fatalf("abuse_email = %v", result["abuse_email"])
	}
	if result["expiration_date"] != "2030-12-31T00:00:00Z" {
		t.Fatalf("expiration_date = %v", result["expiration_date"])
	}
	ns, ok := result["name_servers"].([]string)
	if !ok || len(ns) != 2 {
		t.Fatalf("name_servers = %#v", result["name_servers"])
	}
}

func TestWhoisNeedsRdapFallback(t *testing.T) {
	t.Parallel()

	if !whoisNeedsRdapFallback(nil) {
		t.Fatal("expected fallback for nil data")
	}
	if !whoisNeedsRdapFallback(map[string]any{"raw": "thin"}) {
		t.Fatal("expected fallback for thin whois")
	}
	if whoisNeedsRdapFallback(map[string]any{
		"registrar":       "Example Registrar",
		"domain":          "example.com",
		"expiration_date": "2030-01-01",
	}) {
		t.Fatal("did not expect fallback for complete whois")
	}
}

func TestMergeWhoisRdap(t *testing.T) {
	t.Parallel()

	dst := map[string]any{
		"domain":    "example.com",
		"registrar": "Classic Registrar",
	}
	mergeWhoisRdap(dst, map[string]any{
		"source":           "rdap",
		"registrar":        "RDAP Registrar",
		"registrant_email": "owner@example.com",
		"abuse_email":      "abuse@example.com",
		"expiration_date":  "2030-12-31",
	})

	if dst["registrar"] != "Classic Registrar" {
		t.Fatalf("registrar overwritten = %v", dst["registrar"])
	}
	if dst["registrant_email"] != "owner@example.com" {
		t.Fatalf("registrant_email = %v", dst["registrant_email"])
	}
	if dst["rdap_source"] != "rdap" {
		t.Fatalf("rdap_source = %v", dst["rdap_source"])
	}
	if dst["rdap"] == nil {
		t.Fatal("expected nested rdap object")
	}
}
