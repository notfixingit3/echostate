package pdf

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/notfixingit3/echostate/internal/models"
)

func TestRenderReport(t *testing.T) {
	result := &models.ScanResult{
		Host:      "example.com",
		ScannedAt: time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC),
		WHOIS: map[string]any{
			"domain":          "example.com",
			"registrar":       "Example Registrar, Inc.",
			"expiration_date": "2030-01-01T00:00:00Z",
			"name_servers":    []string{"a.iana-servers.net", "b.iana-servers.net"},
			"status":          []string{"clientDeleteProhibited", "serverDeleteProhibited"},
			"raw":             strings.Repeat("Registrar: EXAMPLE\n", 80),
		},
		ASN: map[string]any{
			"asn":       "15169",
			"prefix":    "142.250.0.0/15",
			"country":   "US",
			"registry":  "arin",
			"allocated": "2000-03-30",
			"as_name":   "GOOGLE, US",
			"ip":        "142.250.80.46",
			"routing": map[string]any{
				"hijack_risk":     "low",
				"visible_origins": []string{"AS15169"},
				"path_profile": map[string]any{
					"stability": "stable",
				},
				"notes": []string{"RPKI validation passed"},
			},
		},
		DNS: map[string]any{
			"A":    []string{"93.184.216.34"},
			"MX":   []string{"10 mail.example.com"},
			"SPF":  map[string]any{"policy": "-all", "record": "v=spf1 -all"},
			"MAIL_POSTURE": map[string]any{
				"grade":    "A",
				"score":    95,
				"findings": []string{"SPF policy is strict (-all)"},
			},
		},
		TLS: map[string]any{
			"issuer":         "Example CA",
			"not_after":      "2030-01-01T00:00:00Z",
			"days_remaining": 1200,
			"jarm":           "abc123",
			"ja3s":           "def456",
		},
		Web: map[string]any{
			"title":            "Example Domain",
			"url":              "https://example.com/",
			"copyrights":       []string{"© 2026 Example Inc.", "All rights reserved."},
			"tech_stack":       []string{"framework:react", "cdn:cloudflare"},
			"security_headers": map[string]any{"strict-transport-security": "max-age=31536000"},
		},
		Favicon: map[string]any{
			"mmh3": "123456",
			"url":  "https://example.com/favicon.ico",
		},
		Crawl: map[string]any{
			"sitemap_urls":      []string{"https://example.com/sitemap.xml"},
			"sitemap_url_count": 1,
		},
		Storage: map[string]any{
			"buckets": []map[string]any{
				{"provider": "s3", "name": "example-bucket"},
			},
		},
		CT: map[string]any{
			"domain":     "example.com",
			"subdomains": []string{"www.example.com", "api.example.com"},
			"count":      2,
		},
		Traceroute: map[string]any{
			"destination": "93.184.216.34",
			"hop_count":   2,
			"hops": []map[string]any{
				{"hop": 1, "ip": "10.0.0.1"},
				{"hop": 2, "ip": "93.184.216.34"},
			},
		},
		Screenshot: map[string]any{
			"url":    "https://example.com/",
			"width":  "1280",
			"height": "720",
		},
		Enrichment: map[string]any{
			"shodan": map[string]any{
				"query": "123456",
				"total": 3,
			},
		},
		Errors: []string{"web gatherer timed out"},
	}

	pdfBytes, err := RenderReport(ReportData{
		Result: result,
		Changes: []string{
			"Added tls",
			"Changed dns.A",
		},
		ChangeDetails: []models.ChangeDetail{
			{
				Type:     "tls_expiry",
				Severity: "high",
				Summary:  "Certificate expires in 14 days",
				Field:    "tls.not_after",
			},
		},
		ScreenshotJPEG: minimalJPEG(t),
		ClientIP:       "203.0.113.10",
		PWhois: &PWhoisInfo{
			OriginAS:    "AS15169",
			OrgName:     "Google LLC",
			CountryCode: "US",
			Prefix:      "142.250.0.0/15",
		},
		PWhoisData: map[string]any{
			"asn": "15169",
			"org": "Google LLC",
		},
	})
	require.NoError(t, err)
	require.NotEmpty(t, pdfBytes)
	assert.True(t, bytes.HasPrefix(pdfBytes, []byte("%PDF")), "output should start with PDF magic bytes")
}

func TestRenderReport_EmptyData(t *testing.T) {
	result := &models.ScanResult{
		Host:      "empty.test",
		ScannedAt: time.Time{},
	}

	pdfBytes, err := RenderReport(ReportData{Result: result})
	require.NoError(t, err)
	require.NotEmpty(t, pdfBytes)
	assert.True(t, bytes.HasPrefix(pdfBytes, []byte("%PDF")))
}

func TestRenderReport_NilResult(t *testing.T) {
	_, err := RenderReport(ReportData{})
	require.Error(t, err)
}

func TestRenderReport_EnrichmentPending(t *testing.T) {
	result := &models.ScanResult{
		Host:      "pending.example",
		ScannedAt: time.Date(2026, 6, 18, 12, 0, 0, 0, time.UTC),
		Enrichment: map[string]any{
			"status": "pending",
		},
	}

	pdfBytes, err := RenderReport(ReportData{Result: result})
	require.NoError(t, err)
	require.NotEmpty(t, pdfBytes)
	assert.True(t, bytes.HasPrefix(pdfBytes, []byte("%PDF")))
}

func TestRenderReport_EnrichmentShodanFormatting(t *testing.T) {
	result := &models.ScanResult{
		Host:      "shodan.example",
		ScannedAt: time.Date(2026, 6, 18, 12, 0, 0, 0, time.UTC),
		Enrichment: map[string]any{
			"status":       "completed",
			"completed_at": "2026-06-18T12:00:05Z",
			"shodan": map[string]any{
				"host": map[string]any{
					"ip":         "203.0.113.1",
					"org":        "Example Org",
					"ports":      []any{80, 443},
					"vuln_count": 2,
					"vulns":      []any{"CVE-2024-0001", "CVE-2024-0002"},
				},
				"favicon_search": map[string]any{
					"query": "12345",
					"total": 3,
					"matches": []any{
						map[string]any{"ip_str": "198.51.100.1", "org": "Match Org", "ports": []any{443}},
					},
				},
			},
		},
	}

	pdfBytes, err := RenderReport(ReportData{Result: result})
	require.NoError(t, err)
	require.NotEmpty(t, pdfBytes)
	assert.True(t, bytes.HasPrefix(pdfBytes, []byte("%PDF")))
}

func minimalJPEG(t *testing.T) []byte {
	t.Helper()

	const tinyJPEG = "/9j/4AAQSkZJRgABAQAAAQABAAD/2wBDAAgGBgcGBQgHBwcJCQgKDBQNDAsLDBkSEw8UHRofHh0aHBwgJC4nICIsIxwcKDcpLDAxNDQ0Hyc5PTgyPC4zNDL/2wBDAQkJCQwLDBgNDRgyIRwhMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjL/wAARCAABAAEDASIAAhEBAxEB/8QAFQABAQAAAAAAAAAAAAAAAAAAAAb/xAAUEAEAAAAAAAAAAAAAAAAAAAAA/8QAFQEBAQAAAAAAAAAAAAAAAAAAAAX/xAAUEQEAAAAAAAAAAAAAAAAAAAAA/9oADAMBAAIRAxEAPwCdABmX/9k="
	data, err := base64.StdEncoding.DecodeString(tinyJPEG)
	require.NoError(t, err)
	return data
}