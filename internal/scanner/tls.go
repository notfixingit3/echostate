package scanner

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"
)

func gatherTLS(ctx context.Context, host string) (string, map[string]any, error) {
	host = NormalizeHost(host)
	if host == "" {
		return "tls", nil, fmt.Errorf("empty host")
	}

	dialCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	dialer := &net.Dialer{}
	conn, err := tls.DialWithDialer(dialer, "tcp", host+":443", &tls.Config{
		InsecureSkipVerify: true, // We want to see the cert even if it's invalid
	})
	if err != nil {
		return "tls", nil, fmt.Errorf("tls dial: %w", err)
	}
	defer conn.Close()

	// Use context to ensure we don't block forever
	go func() {
		<-dialCtx.Done()
		conn.Close()
	}()

	state := conn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		return "tls", nil, fmt.Errorf("no peer certificates found")
	}

	cert := state.PeerCertificates[0]

	var ipSans []string
	for _, ip := range cert.IPAddresses {
		ipSans = append(ipSans, ip.String())
	}

	data := map[string]any{
		"issuer":              cert.Issuer.String(),
		"subject":             cert.Subject.String(),
		"not_before":          cert.NotBefore,
		"not_after":           cert.NotAfter,
		"dns_names":           cert.DNSNames,
		"ip_sans":             ipSans,
		"version":             cert.Version,
		"signature_algorithm": cert.SignatureAlgorithm.String(),
		"cipher_suite":        tls.CipherSuiteName(state.CipherSuite),
	}

	if jarmHash, err := fingerprintJARM(ctx, host, 443); err == nil {
		data["jarm"] = jarmHash
	}

	return "tls", data, nil
}
