package scanner

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	jarm "github.com/hdm/jarm-go"
)

func fingerprintJARM(ctx context.Context, host string, port int) (string, error) {
	host = NormalizeHost(host)
	if host == "" {
		return "", fmt.Errorf("empty host")
	}

	probes := jarm.GetProbes(host, port)
	results := make([]string, 0, len(probes))

	for _, probe := range probes {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		dialer := net.Dialer{Timeout: 3 * time.Second}
		addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
		conn, err := dialer.DialContext(ctx, "tcp", addr)
		if err != nil {
			results = append(results, "")
			continue
		}

		data := jarm.BuildProbe(probe)
		_ = conn.SetWriteDeadline(time.Now().Add(3 * time.Second))
		if _, err := conn.Write(data); err != nil {
			results = append(results, "")
			_ = conn.Close()
			continue
		}

		buff := make([]byte, 1484)
		_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
		_, _ = conn.Read(buff)
		_ = conn.Close()

		ans, err := jarm.ParseServerHello(buff, probe)
		if err != nil {
			results = append(results, "")
			continue
		}
		results = append(results, ans)
	}

	hash := jarm.RawHashToFuzzyHash(strings.Join(results, ","))
	if hash == jarm.ZeroHash {
		return "", fmt.Errorf("jarm fingerprint unavailable")
	}

	return hash, nil
}
