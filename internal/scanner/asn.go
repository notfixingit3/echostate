package scanner

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/notfixingit3/echostate/internal/config"
)

const (
	cymruOriginDomain = "origin.asn.cymru.com"
	cymruASNDomain    = "asn.cymru.com"
	asnGatherTimeout  = 10 * time.Second
	dnsLookupTimeout  = 5 * time.Second
)

// asnLookupFuncs are swappable so tests can mock DNS responses without
// standing up a real resolver.
var (
	lookupIPAddrFunc = func(ctx context.Context, resolver *net.Resolver, host string) ([]net.IPAddr, error) {
		return resolver.LookupIPAddr(ctx, host)
	}
	lookupTXTFunc = func(ctx context.Context, resolver *net.Resolver, name string) ([]string, error) {
		return resolver.LookupTXT(ctx, name)
	}
)

// gatherASN performs a passive ASN/BGP lookup for host using Team Cymru's
// DNS services. It normalizes the host, resolves it to an IPv4 address when
// necessary, and queries origin.asn.cymru.com and asn.cymru.com for origin
// ASN and AS description data.
func gatherASN(ctx context.Context, host string) (string, map[string]any, error) {
	host = NormalizeHost(host)
	if host == "" {
		return "asn", nil, fmt.Errorf("empty host")
	}

	ctx, cancel := context.WithTimeout(ctx, asnGatherTimeout)
	defer cancel()

	var resolver *net.Resolver
	settings := config.GetSettings()
	if settings.DNSServers != "" {
		servers := strings.Split(settings.DNSServers, ",")
		resolver = &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{Timeout: dnsLookupTimeout}
				return d.DialContext(ctx, "udp", strings.TrimSpace(servers[0])+":53")
			},
		}
	} else {
		resolver = &net.Resolver{}
	}

	ip := host
	if parsed := net.ParseIP(host); parsed == nil || parsed.To4() == nil {
		resolveCtx, resolveCancel := context.WithTimeout(ctx, dnsLookupTimeout)
		defer resolveCancel()

		addrs, err := lookupIPAddrFunc(resolveCtx, resolver, host)
		if err != nil {
			return "asn", nil, fmt.Errorf("resolve host: %w", err)
		}

		var v4 net.IP
		for _, addr := range addrs {
			if v4 = addr.IP.To4(); v4 != nil {
				break
			}
		}
		if v4 == nil {
			return "asn", nil, fmt.Errorf("no IPv4 address found for %s", host)
		}
		ip = v4.String()
	}

	origin, err := queryCymruOrigin(ctx, resolver, ip)
	if err != nil {
		return "asn", nil, err
	}

	if asName, ok := origin["as_name"].(string); !ok || strings.TrimSpace(asName) == "" {
		name, err := queryCymruASNName(ctx, resolver, origin["asn"])
		if err == nil && name != "" {
			origin["as_name"] = name
		}
	}

	origin["ip"] = ip
	return "asn", origin, nil
}

// queryCymruOrigin looks up the reversed IPv4 address under
// origin.asn.cymru.com and parses the TXT record.
func queryCymruOrigin(ctx context.Context, resolver *net.Resolver, ip string) (map[string]any, error) {
	rev := reverseIPForCymru(ip)
	qname := rev + "." + cymruOriginDomain

	txtCtx, txtCancel := context.WithTimeout(ctx, dnsLookupTimeout)
	defer txtCancel()

	txts, err := lookupTXTFunc(txtCtx, resolver, qname)
	if err != nil {
		return nil, fmt.Errorf("cymru origin lookup: %w", err)
	}
	if len(txts) == 0 {
		return nil, fmt.Errorf("no cymru origin TXT record for %s", ip)
	}

	parsed, err := parseCymruTXT(txts[0])
	if err != nil {
		return nil, fmt.Errorf("parse cymru origin TXT: %w", err)
	}
	return parsed, nil
}

// queryCymruASNName looks up AS<asn>.asn.cymru.com to retrieve the AS name.
func queryCymruASNName(ctx context.Context, resolver *net.Resolver, asn any) (string, error) {
	asnStr, ok := asn.(string)
	if !ok || asnStr == "" {
		return "", fmt.Errorf("missing asn")
	}

	qname := "AS" + asnStr + "." + cymruASNDomain

	txtCtx, txtCancel := context.WithTimeout(ctx, dnsLookupTimeout)
	defer txtCancel()

	txts, err := lookupTXTFunc(txtCtx, resolver, qname)
	if err != nil {
		return "", err
	}
	if len(txts) == 0 {
		return "", fmt.Errorf("no cymru asn TXT record for AS%s", asnStr)
	}

	return parseCymruASNNameTXT(txts[0])
}

// reverseIPForCymru returns the reversed-octet IPv4 form used by Team Cymru
// DNS zones. For example, "1.2.3.4" becomes "4.3.2.1".
func reverseIPForCymru(ip string) string {
	parts := strings.Split(ip, ".")
	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}
	return strings.Join(parts, ".")
}

// parseCymruTXT parses a Team Cymru origin ASN TXT record. The canonical DNS
// format is "ASN | BGP Prefix | CC | Registry | Allocated"; some records may
// also include an AS Name as a sixth field.
func parseCymruTXT(txt string) (map[string]any, error) {
	txt = strings.TrimSpace(txt)
	parts := strings.Split(txt, " | ")
	if len(parts) < 5 {
		return nil, fmt.Errorf("unexpected cymru TXT format: %q", txt)
	}

	result := map[string]any{
		"asn":       strings.TrimSpace(parts[0]),
		"prefix":    strings.TrimSpace(parts[1]),
		"country":   strings.TrimSpace(parts[2]),
		"registry":  strings.TrimSpace(parts[3]),
		"allocated": strings.TrimSpace(parts[4]),
		"as_name":   "",
	}
	if len(parts) >= 6 {
		result["as_name"] = strings.TrimSpace(parts[5])
	}
	return result, nil
}

// parseCymruASNNameTXT parses a Team Cymru asn.cymru.com TXT record, which
// has the format "ASN | CC | Registry | Allocated | AS Name".
func parseCymruASNNameTXT(txt string) (string, error) {
	txt = strings.TrimSpace(txt)
	parts := strings.Split(txt, " | ")
	if len(parts) < 5 {
		return "", fmt.Errorf("unexpected cymru ASN name TXT format: %q", txt)
	}
	return strings.TrimSpace(parts[4]), nil
}
