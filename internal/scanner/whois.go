package scanner

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/likexian/whois"
	whoisparser "github.com/likexian/whois-parser"
)

const maxWHOISRaw = 8192

// whoisLookup is a package-level shim for whois.Whois so tests can inject
// raw responses without hitting the network.
var whoisLookup = whois.Whois

// gatherWHOIS normalizes the host, performs a WHOIS lookup with a 10-second
// timeout, parses the response, and returns selected fields plus a truncated
// copy of the raw text. RDAP is used as a fallback when classic WHOIS is thin.
func gatherWHOIS(ctx context.Context, host string) (string, map[string]any, error) {
	host = NormalizeHost(host)
	if host == "" {
		return "whois", nil, fmt.Errorf("empty host")
	}
	if net.ParseIP(host) != nil {
		return "whois", nil, fmt.Errorf("whois lookup not applicable to IP addresses")
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	type result struct {
		raw string
		err error
	}

	ch := make(chan result, 1)
	go func() {
		raw, err := whoisLookup(host)
		ch <- result{raw: raw, err: err}
	}()

	var (
		value map[string]any
		err   error
	)

	select {
	case <-ctx.Done():
		return "whois", nil, fmt.Errorf("whois lookup timed out: %w", ctx.Err())
	case res := <-ch:
		if res.err != nil {
			value = map[string]any{}
			err = res.err
		} else {
			_, value, err = parseWHOISResponse(res.raw)
		}
	}

	if whoisNeedsRdapFallback(value) {
		if rdap, rdapErr := rdapFetchFunc(ctx, host); rdapErr == nil && len(rdap) > 0 {
			if value == nil {
				value = map[string]any{}
			}
			mergeWhoisRdap(value, rdap)
			if err != nil {
				err = nil
			}
		}
	}

	if len(value) == 0 {
		if err != nil {
			return "whois", nil, err
		}
		return "whois", nil, fmt.Errorf("no whois data for %s", host)
	}

	return "whois", value, err
}

func parseWHOISResponse(raw string) (string, map[string]any, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "whois", nil, fmt.Errorf("empty whois response")
	}

	info, err := whoisparser.Parse(raw)

	value := map[string]any{
		"raw": truncateString(raw, maxWHOISRaw),
	}

	if info.Domain != nil {
		value["domain"] = info.Domain.Domain
		value["expiration_date"] = info.Domain.ExpirationDate
		value["name_servers"] = info.Domain.NameServers
		value["status"] = info.Domain.Status
	}

	if info.Registrar != nil {
		value["registrar"] = info.Registrar.Name
	}

	if err != nil {
		return "whois", value, fmt.Errorf("parse whois: %w", err)
	}

	return "whois", value, nil
}

func truncateString(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}