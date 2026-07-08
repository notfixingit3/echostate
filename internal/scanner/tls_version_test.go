package scanner

import (
	"crypto/tls"
	"testing"
)

func TestTLSVersionName(t *testing.T) {
	t.Parallel()

	if got := tlsVersionName(tls.VersionTLS13); got != "TLS 1.3" {
		t.Fatalf("got %q", got)
	}
}
