package scanner

import (
	"context"
	"net"
	"strings"

	"github.com/notfixingit3/echostate/internal/config"
)

var defaultDKIMSelectors = []string{
	"default", "google", "selector1", "selector2", "k1", "s1", "s2", "dkim",
	"mail", "mandrill", "sendgrid", "smtp", "mx", "dkim1", "dkim2",
}

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

	if bimi := lookupBIMIRecord(ctx, resolver, host); len(bimi) > 0 {
		data["BIMI"] = bimi
	}
}

func dkimSelectorList() []string {
	seen := make(map[string]struct{})
	var selectors []string
	add := func(name string) {
		name = strings.TrimSpace(strings.ToLower(name))
		if name == "" {
			return
		}
		if _, ok := seen[name]; ok {
			return
		}
		seen[name] = struct{}{}
		selectors = append(selectors, name)
	}

	for _, selector := range defaultDKIMSelectors {
		add(selector)
	}
	settings := config.GetSettings()
	for _, part := range strings.Split(settings.DKIMSelectors, ",") {
		add(part)
	}
	return selectors
}

func lookupBIMIRecord(ctx context.Context, resolver *net.Resolver, host string) map[string]any {
	txts, err := resolver.LookupTXT(ctx, "default._bimi."+host)
	if err != nil || len(txts) == 0 {
		return nil
	}
	for _, txt := range txts {
		record := strings.TrimSpace(txt)
		if !strings.HasPrefix(strings.ToLower(record), "v=bimi1") {
			continue
		}
		result := map[string]any{"record": record}
		for _, part := range strings.Split(record, ";") {
			part = strings.TrimSpace(part)
			key, value, ok := strings.Cut(part, "=")
			if !ok {
				continue
			}
			switch strings.ToLower(strings.TrimSpace(key)) {
			case "l":
				result["logo_url"] = strings.TrimSpace(value)
			case "a":
				result["authority"] = strings.TrimSpace(value)
			}
		}
		return result
	}
	return nil
}

func enrichMailTransport(ctx context.Context, host string, data map[string]any) {
	if mtaSts, err := fetchMTASTS(ctx, host); err == nil && len(mtaSts) > 0 {
		data["MTA_STS"] = mtaSts
	}
	if tlsRpt := lookupTLSRPT(ctx, host); len(tlsRpt) > 0 {
		data["TLS_RPT"] = tlsRpt
	}
	if posture := computeMailPosture(data); posture != nil {
		data["MAIL_POSTURE"] = posture
	}
}

func fetchMTASTS(ctx context.Context, host string) (map[string]any, error) {
	body, err := fetchHostResource(ctx, host, "/.well-known/mta-sts.txt", 32*1024)
	if err != nil || len(body) == 0 {
		return nil, err
	}
	parsed := parseMTASTS(string(body))
	if len(parsed) == 0 {
		return nil, nil
	}
	parsed["source"] = "/.well-known/mta-sts.txt"
	return parsed, nil
}

func parseMTASTS(content string) map[string]any {
	result := map[string]any{}
	var mxHosts []string

	for _, line := range strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key = strings.ToLower(strings.TrimSpace(key))
		value = strings.TrimSpace(value)
		switch key {
		case "version":
			result["version"] = value
		case "mode":
			result["mode"] = strings.ToLower(value)
		case "max_age":
			result["max_age"] = value
		case "mx":
			if value != "" {
				mxHosts = append(mxHosts, value)
			}
		}
	}

	if len(mxHosts) > 0 {
		result["mx"] = mxHosts
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func lookupTLSRPT(ctx context.Context, host string) map[string]any {
	resolver := getResolver()
	txts, err := resolver.LookupTXT(ctx, "_smtp._tls."+host)
	if err != nil || len(txts) == 0 {
		return nil
	}
	return parseTLSRPT(txts)
}

func parseTLSRPT(txts []string) map[string]any {
	for _, txt := range txts {
		record := strings.TrimSpace(txt)
		if !strings.HasPrefix(strings.ToLower(record), "v=tlsrptv1") {
			continue
		}
		result := map[string]any{"record": record}
		for _, part := range strings.Split(record, ";") {
			part = strings.TrimSpace(part)
			key, value, ok := strings.Cut(part, "=")
			if !ok {
				continue
			}
			switch strings.ToLower(strings.TrimSpace(key)) {
			case "rua":
				result["rua"] = strings.TrimSpace(value)
			}
		}
		return result
	}
	return nil
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

	for _, selector := range dkimSelectorList() {
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