package enrichment

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var emailPattern = regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)

const maxEmailsPerSnapshot = 3

func extractEmails(payload map[string]any) []string {
	seen := map[string]struct{}{}
	var emails []string

	add := func(raw string) {
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
	}

	if whois, ok := payload["whois"].(map[string]any); ok {
		for _, key := range []string{"registrant_email", "admin_email", "tech_email", "abuse_email", "email"} {
			if value := strings.TrimSpace(fmt.Sprint(whois[key])); value != "" && value != "<nil>" {
				add(value)
			}
		}
		if raw, ok := whois["raw"].(string); ok {
			add(raw)
		}
	}
	if crawl, ok := payload["crawl"].(map[string]any); ok {
		if robots, ok := crawl["robots_txt"].(string); ok {
			add(robots)
		}
	}

	sort.Strings(emails)
	if len(emails) > maxEmailsPerSnapshot {
		emails = emails[:maxEmailsPerSnapshot]
	}
	return emails
}