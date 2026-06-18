package scanner

import (
	"context"
	"fmt"
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
			var ips []string
			for _, a := range addrs {
				ips = append(ips, a.String())
			}
			mu.Lock()
			data["A"] = ips
			mu.Unlock()
		}
	}()

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

	wg.Wait()

	enrichMailSecurity(ctx, resolver, host, data)

	if len(data) == 0 {
		return "dns", nil, fmt.Errorf("no dns records found for %s", host)
	}

	return "dns", data, nil
}
