package scanner

import (
	"fmt"
	"strings"
)

func enrichInfraLabels(data map[string]any) {
	if len(data) == 0 {
		return
	}

	labels := map[string]any{}

	if cname, ok := data["CNAME"].(string); ok && strings.TrimSpace(cname) != "" {
		if provider := detectCDNFromHostname(cname); provider != "" {
			labels["cdn_provider"] = provider
			labels["cname_target"] = strings.TrimSuffix(strings.TrimSpace(cname), ".")
		}
	}

	switch mxRecords := data["MX"].(type) {
	case []string:
		if provider := detectMailProvider(mxRecords); provider != "" {
			labels["mail_provider"] = provider
		}
	case []any:
		var hosts []string
		for _, item := range mxRecords {
			hosts = append(hosts, fmt.Sprint(item))
		}
		if provider := detectMailProvider(hosts); provider != "" {
			labels["mail_provider"] = provider
		}
	}

	if len(labels) == 0 {
		return
	}
	data["INFRA_LABELS"] = labels
}

func detectCDNFromHostname(host string) string {
	host = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(host), "."))
	switch {
	case strings.Contains(host, "cloudflare.net"), strings.Contains(host, "cloudflare-dns.com"):
		return "Cloudflare"
	case strings.Contains(host, "fastly.net"), strings.Contains(host, "fastlylb.net"):
		return "Fastly"
	case strings.Contains(host, "akamaiedge.net"), strings.Contains(host, "akamai.net"), strings.Contains(host, "edgekey.net"):
		return "Akamai"
	case strings.Contains(host, "cloudfront.net"):
		return "Amazon CloudFront"
	case strings.Contains(host, "azureedge.net"), strings.Contains(host, "azurefd.net"):
		return "Azure CDN"
	case strings.Contains(host, "googledomains.com"), strings.Contains(host, "googleusercontent.com"):
		return "Google"
	case strings.Contains(host, "incapdns.net"):
		return "Imperva"
	default:
		return ""
	}
}

func detectMailProvider(mxRecords []string) string {
	if len(mxRecords) == 0 {
		return ""
	}

	joined := strings.ToLower(strings.Join(mxRecords, " "))
	switch {
	case strings.Contains(joined, "google.com"), strings.Contains(joined, "googlemail.com"), strings.Contains(joined, "smtp.google.com"):
		return "Google Workspace"
	case strings.Contains(joined, "outlook.com"), strings.Contains(joined, "protection.outlook.com"), strings.Contains(joined, "mail.protection.outlook.com"):
		return "Microsoft 365"
	case strings.Contains(joined, "pphosted.com"), strings.Contains(joined, "mimecast.com"):
		return "Mimecast"
	case strings.Contains(joined, "protonmail.ch"), strings.Contains(joined, "proton.me"):
		return "Proton Mail"
	case strings.Contains(joined, "zoho.com"):
		return "Zoho Mail"
	case strings.Contains(joined, "messagelabs.com"):
		return "Symantec Email Security"
	default:
		return ""
	}
}