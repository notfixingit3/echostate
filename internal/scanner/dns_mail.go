package scanner

import (
	"context"
	"net"
	"strings"
)

var dkimSelectors = []string{"default", "google", "selector1", "k1", "s1", "dkim"}

func enrichMailSecurity(ctx context.Context, resolver *net.Resolver, host string, data map[string]any) {
	if txts, ok := data["TXT"].([]string); ok {
		if spf := parseSPFRecord(txts); spf != nil {
			data["SPF"] = spf
		}
	}

	if dmarcTXT, ok := data["DMARC"].([]string); ok {
		if dmarc := parseDMARCRecord(dmarcTXT); dmarc != nil {
			data["DMARC_PARSED"] = dmarc
		}
	}

	if dkim := lookupDKIMRecords(ctx, resolver, host); len(dkim) > 0 {
		data["DKIM"] = dkim
	}
}

func parseSPFRecord(txts []string) map[string]any {
	for _, txt := range txts {
		record := strings.TrimSpace(txt)
		if !strings.HasPrefix(strings.ToLower(record), "v=spf1") {
			continue
		}
		parts := strings.Fields(record)
		result := map[string]any{
			"record": record,
		}
		for _, part := range parts[1:] {
			lower := strings.ToLower(part)
			if strings.HasSuffix(lower, "all") {
				result["policy"] = part
			}
		}
		return result
	}
	return nil
}

func parseDMARCRecord(txts []string) map[string]any {
	for _, txt := range txts {
		record := strings.TrimSpace(txt)
		if !strings.HasPrefix(strings.ToLower(record), "v=dmarc1") {
			continue
		}
		result := map[string]any{
			"record": record,
		}
		for _, part := range strings.Split(record, ";") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			key, value, ok := strings.Cut(part, "=")
			if !ok {
				continue
			}
			switch strings.ToLower(strings.TrimSpace(key)) {
			case "p":
				result["policy"] = strings.TrimSpace(value)
			case "sp":
				result["subdomain_policy"] = strings.TrimSpace(value)
			case "pct":
				result["percentage"] = strings.TrimSpace(value)
			case "rua":
				result["aggregate_report_uri"] = strings.TrimSpace(value)
			}
		}
		return result
	}
	return nil
}

func lookupDKIMRecords(ctx context.Context, resolver *net.Resolver, host string) []map[string]any {
	var records []map[string]any
	seen := make(map[string]struct{})

	for _, selector := range dkimSelectors {
		name := selector + "._domainkey." + host
		txts, err := resolver.LookupTXT(ctx, name)
		if err != nil {
			continue
		}
		for _, txt := range txts {
			record := strings.TrimSpace(txt)
			if !strings.HasPrefix(strings.ToLower(record), "v=dkim1") {
				continue
			}
			if _, ok := seen[record]; ok {
				continue
			}
			seen[record] = struct{}{}
			records = append(records, map[string]any{
				"selector": selector,
				"name":     name,
				"record":   record,
			})
		}
	}

	return records
}