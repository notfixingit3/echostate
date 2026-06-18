package enrichment

import "testing"

func TestExtractEmails_DedupesAndLimits(t *testing.T) {
	payload := map[string]any{
		"whois": map[string]any{
			"registrant_email": "admin@example.org",
			"raw":              "Contact: ops@example.org\nAlso: admin@example.org",
		},
	}
	emails := extractEmails(payload)
	if len(emails) != 2 {
		t.Fatalf("extractEmails() = %#v", emails)
	}
}