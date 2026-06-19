package diff

import (
	"testing"

	"github.com/notfixingit3/echostate/internal/models"
)

func TestCompute_CAAAdded(t *testing.T) {
	previous := map[string]any{"dns": map[string]any{}}
	current := &models.ScanResult{
		DNS: map[string]any{
			"CAA": []any{
				map[string]any{"record": `0 issue "ca.example.com"`},
			},
		},
	}
	entries := Compute(previous, current)
	if len(entries) != 1 || entries[0].Type != "caa_added" {
		t.Fatalf("Compute() = %#v", entries)
	}
}

func TestCompute_MTASTSMode(t *testing.T) {
	previous := map[string]any{
		"dns": map[string]any{
			"MTA_STS": map[string]any{"mode": "testing"},
		},
	}
	current := &models.ScanResult{
		DNS: map[string]any{
			"MTA_STS": map[string]any{"mode": "enforce"},
		},
	}
	entries := Compute(previous, current)
	if len(entries) != 1 || entries[0].Type != "mta_sts_mode" {
		t.Fatalf("Compute() = %#v", entries)
	}
}

func TestCompute_SecurityContactAdded(t *testing.T) {
	previous := map[string]any{
		"crawl": map[string]any{
			"security_txt": map[string]any{
				"contacts": []any{"security@example.com"},
			},
		},
	}
	current := &models.ScanResult{
		Crawl: map[string]any{
			"security_txt": map[string]any{
				"contacts": []any{"security@example.com", "abuse@example.com"},
			},
		},
	}
	entries := Compute(previous, current)
	if len(entries) != 1 || entries[0].Type != "security_contact_added" {
		t.Fatalf("Compute() = %#v", entries)
	}
}

func TestCompute_RedirectChainChanged(t *testing.T) {
	previous := map[string]any{
		"web": map[string]any{
			"redirect_chain": []any{
				map[string]any{"url": "https://example.com", "status": 301},
				map[string]any{"url": "https://www.example.com", "status": 200},
			},
		},
	}
	current := &models.ScanResult{
		Web: map[string]any{
			"redirect_chain": []any{
				map[string]any{"url": "https://example.com", "status": 301},
				map[string]any{"url": "https://cdn.example.com", "status": 301},
				map[string]any{"url": "https://www.example.com", "status": 200},
			},
		},
	}
	entries := Compute(previous, current)
	if len(entries) != 1 || entries[0].Type != "redirect_chain_changed" {
		t.Fatalf("Compute() = %#v", entries)
	}
}

func TestCompute_DNSSECSignedToUnsigned(t *testing.T) {
	previous := map[string]any{
		"dns": map[string]any{
			"DNSSEC": map[string]any{"status": "signed"},
		},
	}
	current := &models.ScanResult{
		DNS: map[string]any{
			"DNSSEC": map[string]any{"status": "unsigned"},
		},
	}
	entries := Compute(previous, current)
	if len(entries) != 1 || entries[0].Type != "dnssec_status" || entries[0].Severity != SeverityWarning {
		t.Fatalf("Compute() = %#v", entries)
	}
}

func TestCompute_NewCTCertificate(t *testing.T) {
	previous := map[string]any{
		"ct": map[string]any{
			"certificates": []any{
				map[string]any{"serial": "111", "issuer": "Old CA"},
			},
		},
	}
	current := &models.ScanResult{
		CT: map[string]any{
			"certificates": []any{
				map[string]any{"serial": "111", "issuer": "Old CA"},
				map[string]any{"serial": "222", "issuer": "New CA"},
			},
		},
	}
	entries := Compute(previous, current)
	if len(entries) != 1 || entries[0].Type != "new_ct_certificate" {
		t.Fatalf("Compute() = %#v", entries)
	}
}

func TestCompute_TLSOCSPStaplingDisabled(t *testing.T) {
	previous := map[string]any{
		"tls": map[string]any{"ocsp_stapled": true},
	}
	current := &models.ScanResult{
		TLS: map[string]any{"ocsp_stapled": false},
	}
	entries := Compute(previous, current)
	if len(entries) != 1 || entries[0].Type != "tls_ocsp_stapling" || entries[0].Severity != SeverityWarning {
		t.Fatalf("Compute() = %#v", entries)
	}
}