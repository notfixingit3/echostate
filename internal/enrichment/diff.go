package enrichment

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/notfixingit3/echostate/internal/diff"
)

// DiffEnrichment compares previous and current enrichment payloads.
func DiffEnrichment(previous, current map[string]any) []diff.Entry {
	if len(current) == 0 {
		return nil
	}

	var entries []diff.Entry
	providers := []struct {
		key  string
		typ  string
		name string
	}{
		{"shodan", "enrichment_shodan", "Shodan"},
		{"censys", "enrichment_censys", "Censys"},
		{"hibp", "enrichment_hibp", "HIBP"},
		{"riskiq", "enrichment_riskiq", "RiskIQ"},
		{"wayback", "enrichment_wayback", "Wayback"},
		{"virustotal", "enrichment_virustotal", "VirusTotal"},
	}

	for _, provider := range providers {
		curBlock, curOK := current[provider.key].(map[string]any)
		if !curOK || len(curBlock) == 0 {
			continue
		}
		prevBlock, _ := previous[provider.key].(map[string]any)
		if len(previous) == 0 || len(prevBlock) == 0 {
			entries = append(entries, diff.Entry{
				Type:     provider.typ + "_added",
				Severity: diff.SeverityInfo,
				Field:    "enrichment." + provider.key,
				Summary:  fmt.Sprintf("%s enrichment available", provider.name),
			})
			continue
		}
		if !jsonEqual(prevBlock, curBlock) {
			entries = append(entries, diff.Entry{
				Type:     provider.typ + "_changed",
				Severity: diff.SeverityInfo,
				Field:    "enrichment." + provider.key,
				Summary:  fmt.Sprintf("%s enrichment changed", provider.name),
			})
		}
	}

	entries = append(entries, diffHIBPBreaches(previous, current)...)
	return entries
}

func diffHIBPBreaches(previous, current map[string]any) []diff.Entry {
	curHIBP, ok := current["hibp"].(map[string]any)
	if !ok {
		return nil
	}
	prevPwned := hibpPwnedEmails(previous)
	curResults, _ := curHIBP["results"].([]any)

	var entries []diff.Entry
	for _, item := range curResults {
		row, ok := item.(map[string]any)
		if !ok {
			continue
		}
		email := strings.TrimSpace(fmt.Sprint(row["email"]))
		pwned := row["pwned"] == true
		if !pwned || email == "" {
			continue
		}
		if _, seen := prevPwned[email]; seen {
			continue
		}
		entries = append(entries, diff.Entry{
			Type:     "enrichment_hibp_breach",
			Severity: diff.SeverityWarning,
			Field:    "enrichment.hibp",
			Summary:  fmt.Sprintf("HIBP breach hit for %s", email),
			Detail:   email,
		})
	}
	return entries
}

func hibpPwnedEmails(enrichment map[string]any) map[string]struct{} {
	out := make(map[string]struct{})
	hibp, ok := enrichment["hibp"].(map[string]any)
	if !ok {
		return out
	}
	results, ok := hibp["results"].([]any)
	if !ok {
		return out
	}
	for _, item := range results {
		row, ok := item.(map[string]any)
		if !ok || row["pwned"] != true {
			continue
		}
		email := strings.TrimSpace(fmt.Sprint(row["email"]))
		if email != "" {
			out[email] = struct{}{}
		}
	}
	return out
}

func jsonEqual(a, b any) bool {
	aj, _ := json.Marshal(a)
	bj, _ := json.Marshal(b)
	return string(aj) == string(bj)
}
