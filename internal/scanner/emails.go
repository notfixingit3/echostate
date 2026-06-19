package scanner

import (
	"regexp"
	"sort"
	"strings"
)

var emailPattern = regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)

const maxEmailsExtracted = 5

func extractEmailsFromText(raw string) []string {
	seen := make(map[string]struct{})
	var emails []string

	for _, match := range emailPattern.FindAllString(raw, -1) {
		email := strings.ToLower(strings.TrimSpace(match))
		if email == "" || strings.HasSuffix(email, "@example.com") {
			continue
		}
		if _, ok := seen[email]; ok {
			continue
		}
		seen[email] = struct{}{}
		emails = append(emails, email)
	}

	sort.Strings(emails)
	if len(emails) > maxEmailsExtracted {
		emails = emails[:maxEmailsExtracted]
	}
	return emails
}