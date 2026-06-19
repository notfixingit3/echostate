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
		if rdap, ok := whois["rdap"].(map[string]any); ok {
			for _, key := range []string{"registrant_email", "abuse_email", "registrar_email"} {
				if value := strings.TrimSpace(fmt.Sprint(rdap[key])); value != "" && value != "<nil>" {
					add(value)
				}
			}
		}
		if raw, ok := whois["raw"].(string); ok {
			add(raw)
		}
	}

	if web, ok := payload["web"].(map[string]any); ok {
		if contactEmails, ok := web["contact_emails"].([]any); ok {
			for _, item := range contactEmails {
				add(fmt.Sprint(item))
			}
		}
		if text, ok := web["copyrights"].([]any); ok {
			for _, item := range text {
				add(fmt.Sprint(item))
			}
		}
	}

	if crawl, ok := payload["crawl"].(map[string]any); ok {
		if securityTxt, ok := crawl["security_txt"].(map[string]any); ok {
			if contacts, ok := securityTxt["contacts"].([]any); ok {
				for _, item := range contacts {
					add(fmt.Sprint(item))
				}
			}
			for _, key := range []string{"contacts", "hiring"} {
				if values, ok := securityTxt[key].([]string); ok {
					for _, item := range values {
						add(item)
					}
				}
			}
		}
	}

	sort.Strings(emails)
	if len(emails) > maxEmailsPerSnapshot {
		emails = emails[:maxEmailsPerSnapshot]
	}
	return emails
}