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

	leafSerial := state.PeerCertificates[0].SerialNumber.String()

	cert := state.PeerCertificates[0]
	now := time.Now().UTC()

	var ipSans []string
	for _, ip := range cert.IPAddresses {
		ipSans = append(ipSans, ip.String())
	}

	chain := make([]map[string]any, 0, len(state.PeerCertificates))
	for index, chainCert := range state.PeerCertificates {
		chain = append(chain, map[string]any{
			"index":      index,
			"issuer":     chainCert.Issuer.String(),
			"subject":    chainCert.Subject.String(),
			"not_before": chainCert.NotBefore,
			"not_after":  chainCert.NotAfter,
		})
	}

	daysRemaining := int(cert.NotAfter.Sub(now).Hours() / 24)

	data := map[string]any{
		"issuer":              cert.Issuer.String(),
		"subject":             cert.Subject.String(),
		"not_before":          cert.NotBefore,
		"not_after":           cert.NotAfter,
		"days_remaining":      daysRemaining,
		"expired":             now.After(cert.NotAfter),
		"dns_names":           cert.DNSNames,
		"ip_sans":             ipSans,
		"version":             cert.Version,
		"serial":              leafSerial,
		"tls_version":         tlsVersionName(state.Version),
		"negotiated_cipher":   tls.CipherSuiteName(state.CipherSuite),
		"signature_algorithm": cert.SignatureAlgorithm.String(),
		"cipher_suite":        tls.CipherSuiteName(state.CipherSuite),
		"ocsp_stapled":        len(state.OCSPResponse) > 0,
		"ocsp_response_bytes": len(state.OCSPResponse),
		"chain":               chain,
		"chain_length":        len(chain),
	}

	if jarmHash, err := fingerprintJARM(ctx, host, 443); err == nil {
		data["jarm"] = jarmHash
	}
	if ja3sHash, err := fingerprintJA3S(ctx, host, 443); err == nil {
		data["ja3s"] = ja3sHash
	}

	return "tls", data, nil
}

func tlsVersionName(version uint16) string {
	switch version {
	case tls.VersionTLS13:
		return "TLS 1.3"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS10:
		return "TLS 1.0"
	default:
		return fmt.Sprintf("0x%04x", version)
	}
}
