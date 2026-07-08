package scanner

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"
)

const certSpotterAPIBase = "https://api.certspotter.com/v1/issuances"

type certSpotterIssuance struct {
	DNSNames []string `json:"dns_names"`
}

func fetchCertSpotter(ctx context.Context, domain string) ([]string, error) {
	queryURL := fmt.Sprintf(
		"%s?domain=%s&include_subdomains=true&expand=dns_names",
		certSpotterAPIBase,
		url.QueryEscape(domain),
	)

	body, _, err := httpGetTimeout(ctx, queryURL, ctResponseMaxBytes, 20*time.Second)
	if err != nil {
		return nil, err
	}

	var issuances []certSpotterIssuance
	if err := json.Unmarshal(body, &issuances); err != nil {
		return nil, fmt.Errorf("parse certspotter JSON: %w", err)
	}

	seen := make(map[string]struct{})
	var subdomains []string
	domain = strings.ToLower(domain)

	for _, issuance := range issuances {
		for _, rawName := range issuance.DNSNames {
			name := normalizeCTName(rawName)
			if name == "" || !isCTSubdomain(name, domain) {
				continue
			}
			if _, ok := seen[name]; ok {
				continue
			}
			seen[name] = struct{}{}
			subdomains = append(subdomains, name)
			if len(subdomains) >= ctMaxSubdomains {
				sort.Strings(subdomains)
				return subdomains, nil
			}
		}
	}

	sort.Strings(subdomains)
	return subdomains, nil
}
