package scanner

import (
	"encoding/json"
	"testing"
)

func TestParseCTCertificates(t *testing.T) {
	t.Parallel()

	payload := []ctCertRecord{
		{
			NameValue:    "example.com\nwww.example.com",
			IssuerName:   "C=US, O=Let's Encrypt, CN=R3",
			NotBefore:    "2024-01-01T00:00:00",
			NotAfter:     "2024-04-01T00:00:00",
			SerialNumber: "01AB",
			CommonName:   "example.com",
			ID:           100,
		},
		{
			NameValue:    "example.com",
			IssuerName:   "C=US, O=Let's Encrypt, CN=R3",
			NotBefore:    "2023-01-01T00:00:00",
			NotAfter:     "2023-04-01T00:00:00",
			SerialNumber: "01AB",
			ID:           99,
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}

	certs, err := parseCTCertificates(body)
	if err != nil {
		t.Fatal(err)
	}
	if len(certs) != 1 {
		t.Fatalf("certs = %d, want 1 deduped", len(certs))
	}
	if certs[0]["issuer"] == "" || certs[0]["serial"] != "01AB" {
		t.Fatalf("cert = %#v", certs[0])
	}
}