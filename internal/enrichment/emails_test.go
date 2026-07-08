package enrichment

import "testing"

func TestExtractEmails_FromSecurityTxtAndWeb(t *testing.T) {
	payload := map[string]any{
		"crawl": map[string]any{
			"security_txt": map[string]any{
				"contacts": []any{"security@example.org"},
			},
		},
		"web": map[string]any{
			"contact_emails": []any{"ops@example.org"},
		},
	}
	emails := extractEmails(payload)
	if len(emails) != 2 {
		t.Fatalf("extractEmails() = %#v", emails)
	}
}

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
