package scanner

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"golang.org/x/net/dns/dnsmessage"
)

const (
	dnsTypeDS     dnsmessage.Type = 43
	dnsTypeDNSKEY dnsmessage.Type = 48
)

var lookupDNSSECFunc = lookupDNSSEC

func lookupDNSSEC(ctx context.Context, host string) (map[string]any, error) {
	host = NormalizeHost(host)
	if host == "" || net.ParseIP(host) != nil {
		return nil, fmt.Errorf("dnssec applies to domain names")
	}

	server := dnsServerAddr()
	zone := host
	var dnskeyCount int
	var signedZone string
	var authenticData bool

	for attempt := 0; attempt < 8 && zone != ""; attempt++ {
		count, ad, err := queryDNSKEY(ctx, server, zone)
		if err == nil && count > 0 {
			dnskeyCount = count
			signedZone = zone
			authenticData = ad
			break
		}
		dot := strings.Index(zone, ".")
		if dot < 0 {
			break
		}
		zone = zone[dot+1:]
	}

	dsCount := 0
	if signedZone != "" {
		if count, _, err := queryDS(ctx, server, signedZone); err == nil {
			dsCount = count
		}
	}

	status := "unsigned"
	if dnskeyCount > 0 {
		status = "signed"
	}

	return map[string]any{
		"status":           status,
		"signed":           dnskeyCount > 0,
		"zone":             signedZone,
		"dnskey_count":     dnskeyCount,
		"ds_count":         dsCount,
		"authentic_data":   authenticData,
		"do_bit_requested": true,
	}, nil
}

func queryDNSKEY(ctx context.Context, server, zone string) (int, bool, error) {
	response, err := queryDNSWithDO(ctx, server, zone, dnsTypeDNSKEY)
	if err != nil {
		return 0, false, err
	}
	count := countAnswers(response, dnsTypeDNSKEY)
	return count, response.Header.AuthenticData, nil
}

func queryDS(ctx context.Context, server, zone string) (int, bool, error) {
	response, err := queryDNSWithDO(ctx, server, zone, dnsTypeDS)
	if err != nil {
		return 0, false, err
	}
	count := countAnswers(response, dnsTypeDS)
	return count, response.Header.AuthenticData, nil
}

func queryDNSWithDO(ctx context.Context, server, zone string, qtype dnsmessage.Type) (dnsmessage.Message, error) {
	dialer := net.Dialer{Timeout: 5 * time.Second}
	conn, err := dialer.DialContext(ctx, "udp", server)
	if err != nil {
		return dnsmessage.Message{}, err
	}
	defer conn.Close()

	msg := dnsmessage.Message{
		Header: dnsmessage.Header{
			Response:         false,
			RecursionDesired: true,
		},
		Questions: []dnsmessage.Question{{
			Name:  dnsmessage.MustNewName(zone + "."),
			Type:  qtype,
			Class: dnsmessage.ClassINET,
		}},
		Additionals: []dnsmessage.Resource{{
			Header: dnsmessage.ResourceHeader{
				Name:  dnsmessage.MustNewName("."),
				Type:  dnsmessage.TypeOPT,
				Class: 1232,
				TTL:   1 << 15, // DO bit
			},
			Body: &dnsmessage.OPTResource{},
		}},
	}

	payload, err := msg.Pack()
	if err != nil {
		return dnsmessage.Message{}, err
	}

	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(5 * time.Second)
	}
	if err := conn.SetDeadline(deadline); err != nil {
		return dnsmessage.Message{}, err
	}
	if _, err := conn.Write(payload); err != nil {
		return dnsmessage.Message{}, err
	}

	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		return dnsmessage.Message{}, err
	}

	var response dnsmessage.Message
	if err := response.Unpack(buf[:n]); err != nil {
		return dnsmessage.Message{}, err
	}
	return response, nil
}

func countAnswers(response dnsmessage.Message, qtype dnsmessage.Type) int {
	count := 0
	for _, resource := range response.Answers {
		if resource.Header.Type == qtype {
			count++
		}
	}
	for _, resource := range response.Authorities {
		if resource.Header.Type == qtype {
			count++
		}
	}
	return count
}