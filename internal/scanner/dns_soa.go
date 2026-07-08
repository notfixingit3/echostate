package scanner

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/notfixingit3/echostate/internal/config"
	"golang.org/x/net/dns/dnsmessage"
)

var (
	lookupSOAFunc = lookupSOA
	querySOAFunc  = querySOA
)

func lookupSOA(ctx context.Context, host string) (map[string]any, error) {
	host = NormalizeHost(host)
	if host == "" {
		return nil, fmt.Errorf("empty host")
	}

	server := dnsServerAddr()
	zone := host
	for attempt := 0; attempt < 8 && zone != ""; attempt++ {
		soa, err := querySOAFunc(ctx, server, zone)
		if err == nil && soa != nil {
			soa["zone"] = zone
			return soa, nil
		}
		dot := strings.Index(zone, ".")
		if dot < 0 {
			break
		}
		zone = zone[dot+1:]
	}
	return nil, fmt.Errorf("no SOA record for %s", host)
}

func dnsServerAddr() string {
	settings := config.GetSettings()
	if settings.DNSServers != "" {
		for _, part := range strings.Split(settings.DNSServers, ",") {
			if server := strings.TrimSpace(part); server != "" {
				return net.JoinHostPort(server, "53")
			}
		}
	}
	return "8.8.8.8:53"
}

func querySOA(ctx context.Context, server, zone string) (map[string]any, error) {
	dialer := net.Dialer{Timeout: 5 * time.Second}
	conn, err := dialer.DialContext(ctx, "udp", server)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	msg := dnsmessage.Message{
		Header: dnsmessage.Header{
			Response:         false,
			OpCode:           0,
			RecursionDesired: true,
		},
		Questions: []dnsmessage.Question{{
			Name:  dnsmessage.MustNewName(zone + "."),
			Type:  dnsmessage.TypeSOA,
			Class: dnsmessage.ClassINET,
		}},
	}

	payload, err := msg.Pack()
	if err != nil {
		return nil, err
	}

	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(5 * time.Second)
	}
	if err := conn.SetDeadline(deadline); err != nil {
		return nil, err
	}
	if _, err := conn.Write(payload); err != nil {
		return nil, err
	}

	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}

	var response dnsmessage.Message
	if err := response.Unpack(buf[:n]); err != nil {
		return nil, err
	}

	if soa := extractSOAFromResources(response.Answers); soa != nil {
		return soa, nil
	}
	if soa := extractSOAFromResources(response.Authorities); soa != nil {
		return soa, nil
	}
	return nil, fmt.Errorf("no SOA in response for %s", zone)
}

func extractSOAFromResources(resources []dnsmessage.Resource) map[string]any {
	for _, resource := range resources {
		if resource.Header.Type != dnsmessage.TypeSOA {
			continue
		}
		soa, ok := resource.Body.(*dnsmessage.SOAResource)
		if !ok {
			continue
		}
		return soaResourceToMap(soa)
	}
	return nil
}

func soaResourceToMap(soa *dnsmessage.SOAResource) map[string]any {
	mname := strings.TrimSuffix(soa.NS.String(), ".")
	rname := strings.TrimSuffix(soa.MBox.String(), ".")
	return map[string]any{
		"record":      formatSOARecord(soa),
		"mname":       mname,
		"rname":       rname,
		"serial":      soa.Serial,
		"refresh":     soa.Refresh,
		"retry":       soa.Retry,
		"expire":      soa.Expire,
		"minimum_ttl": soa.MinTTL,
	}
}

func formatSOARecord(soa *dnsmessage.SOAResource) string {
	return fmt.Sprintf(
		"%s %s %d %d %d %d %d",
		strings.TrimSuffix(soa.NS.String(), "."),
		strings.TrimSuffix(soa.MBox.String(), "."),
		soa.Serial,
		soa.Refresh,
		soa.Retry,
		soa.Expire,
		soa.MinTTL,
	)
}
