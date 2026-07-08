package scanner

import "testing"

func TestExtractEmailsFromText(t *testing.T) {
	emails := extractEmailsFromText("Reach us at ops@example.org or sales@example.org")
	if len(emails) != 2 {
		t.Fatalf("emails = %#v", emails)
	}
}
