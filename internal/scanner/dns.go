package scanner

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
)

func gatherDNS(ctx context.Context, host string) (string, map[string]any, error) {
	host = NormalizeHost(host)
	if host == "" {
		return "dns", nil, fmt.Errorf("empty host")
	}

	resolver := getResolver()

	var wg sync.WaitGroup
	var mu sync.Mutex
	data := make(map[string]any)

	// A & AAAA
	wg.Add(1)
	go func() {
		defer wg.Done()
		if addrs, err := resolver.LookupIPAddr(ctx, host); err == nil {
			var v4, v6 []string
			for _, a := range addrs {
				if ip4 := a.IP.To4(); ip4 != nil {
					v4 = append(v4, ip4.String())
					continue
				}
				if a.IP.To16() != nil {
					v6 = append(v6, a.IP.String())
				}
			}
			mu.Lock()
			if len(v4) > 0 {
				data["A"] = v4
			}
			if len(v6) > 0 {
				data["AAAA"] = v6
			}
			mu.Unlock()
		}
	}()

	if parsed := net.ParseIP(host); parsed != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			names, err := resolver.LookupAddr(ctx, host)
			if err != nil || len(names) == 0 {
				return
			}
			ptr := make([]string, 0, len(names))
			for _, name := range names {
				name = strings.TrimSuffix(strings.TrimSpace(name), ".")
				if name != "" {
					ptr = append(ptr, name)
				}
			}
			if len(ptr) > 0 {
				mu.Lock()
				data["PTR"] = ptr
				mu.Unlock()
			}
		}()
	}

	// MX
	wg.Add(1)
	go func() {
		defer wg.Done()
		if mxs, err := resolver.LookupMX(ctx, host); err == nil {
			var mxStrs []string
			for _, mx := range mxs {
				mxStrs = append(mxStrs, fmt.Sprintf("%d %s", mx.Pref, mx.Host))
			}
			mu.Lock()
			data["MX"] = mxStrs
			mu.Unlock()
		}
	}()

	// NS
	wg.Add(1)
	go func() {
		defer wg.Done()
		if nss, err := resolver.LookupNS(ctx, host); err == nil {
			var nsStrs []string
			for _, ns := range nss {
				nsStrs = append(nsStrs, ns.Host)
			}
			mu.Lock()
			data["NS"] = nsStrs
			mu.Unlock()
		}
	}()

	// TXT
	wg.Add(1)
	go func() {
		defer wg.Done()
		if txts, err := resolver.LookupTXT(ctx, host); err == nil {
			mu.Lock()
			data["TXT"] = txts
			mu.Unlock()
		}
	}()

	// CNAME
	wg.Add(1)
	go func() {
		defer wg.Done()
		if cname, err := resolver.LookupCNAME(ctx, host); err == nil {
			mu.Lock()
			data["CNAME"] = cname
			mu.Unlock()
		}
	}()

	// DMARC
	wg.Add(1)
	go func() {
		defer wg.Done()
		if dmarc, err := resolver.LookupTXT(ctx, "_dmarc."+host); err == nil {
			mu.Lock()
			data["DMARC"] = dmarc
			mu.Unlock()
		}
	}()

	// SOA (zone authority; walks up labels when queried on a subdomain)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if soa, err := lookupSOAFunc(ctx, host); err == nil && soa != nil {
			mu.Lock()
			data["SOA"] = soa
			mu.Unlock()
		}
	}()

	// CAA (certificate authority authorization)
	wg.Add(1)
	go func() {
		defer wg.Done()
		if records, err := lookupCAAFunc(ctx, host); err == nil && len(records) > 0 {
			mu.Lock()
			data["CAA"] = records
			mu.Unlock()
		}
	}()

	wg.Wait()

	enrichMailSecurity(ctx, resolver, host, data)
	enrichMailTransport(ctx, host, data)
	enrichInfraLabels(data)

	if net.ParseIP(host) == nil {
		if dnssec, err := lookupDNSSECFunc(ctx, host); err == nil && len(dnssec) > 0 {
			data["DNSSEC"] = dnssec
		}
	}

	if len(data) == 0 {
		return "dns", nil, fmt.Errorf("no dns records found for %s", host)
	}

	return "dns", data, nil
}
