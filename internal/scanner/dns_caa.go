package scanner

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"golang.org/x/net/dns/dnsmessage"
)

const dnsTypeCAA dnsmessage.Type = 257

var (
	lookupCAAFunc = lookupCAA
	queryCAAFunc  = queryCAA
)

func lookupCAA(ctx context.Context, host string) ([]map[string]any, error) {
	host = NormalizeHost(host)
	if host == "" {
		return nil, fmt.Errorf("empty host")
	}

	server := dnsServerAddr()
	zone := host
	for attempt := 0; attempt < 8 && zone != ""; attempt++ {
		records, err := queryCAAFunc(ctx, server, zone)
		if err == nil && len(records) > 0 {
			return records, nil
		}
		dot := strings.Index(zone, ".")
		if dot < 0 {
			break
		}
		zone = zone[dot+1:]
	}
	return nil, fmt.Errorf("no CAA records for %s", host)
}

func queryCAA(ctx context.Context, server, zone string) ([]map[string]any, error) {
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
			Type:  dnsTypeCAA,
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

	records := extractCAAFromResources(response.Answers)
	if len(records) > 0 {
		return records, nil
	}
	records = extractCAAFromResources(response.Authorities)
	if len(records) > 0 {
		return records, nil
	}
	return nil, fmt.Errorf("no CAA in response for %s", zone)
}

func extractCAAFromResources(resources []dnsmessage.Resource) []map[string]any {
	var records []map[string]any
	for _, resource := range resources {
		if resource.Header.Type != dnsTypeCAA {
			continue
		}
		unknown, ok := resource.Body.(*dnsmessage.UnknownResource)
		if !ok {
			continue
		}
		if parsed, ok := parseCAARData(unknown.Data); ok {
			records = append(records, parsed)
		}
	}
	return records
}

func parseCAARData(data []byte) (map[string]any, bool) {
	if len(data) < 2 {
		return nil, false
	}
	flags := int(data[0])
	tagLen := int(data[1])
	if tagLen <= 0 || len(data) < 2+tagLen {
		return nil, false
	}
	tag := string(data[2 : 2+tagLen])
	value := strings.TrimSpace(string(data[2+tagLen:]))
	if tag == "" || value == "" {
		return nil, false
	}
	return map[string]any{
		"flags":  flags,
		"tag":    tag,
		"value":  value,
		"record": fmt.Sprintf("%d %s %q", flags, tag, value),
	}, true
}
