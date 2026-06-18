package pdf

import (
	"bytes"
	"encoding/base64"
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
		},
		ASN: map[string]any{
			"asn":       "15169",
			"prefix":    "142.250.0.0/15",
			"country":   "US",
			"registry":  "arin",
			"allocated": "2000-03-30",
			"as_name":   "GOOGLE, US",
			"ip":        "142.250.80.46",
		},
		DNS: map[string]any{
			"a": []string{"93.184.216.34"},
		},
		TLS: map[string]any{
			"issuer":    "Example CA",
			"not_after": "2030-01-01T00:00:00Z",
			"jarm":      "abc123",
		},
		Web: map[string]any{
			"title":      "Example Domain",
			"url":        "https://example.com/",
			"copyrights": []string{"© 2026 Example Inc.", "All rights reserved."},
		},
		Screenshot: map[string]any{
			"url":    "https://example.com/",
			"width":  "1280",
			"height": "720",
		},
		Errors: []string{"web gatherer timed out"},
	}

	pdfBytes, err := RenderReport(ReportData{
		Result: result,
		Changes: []string{
			"Added tls",
			"Changed dns.a",
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

func minimalJPEG(t *testing.T) []byte {
	t.Helper()

	const tinyJPEG = "/9j/4AAQSkZJRgABAQAAAQABAAD/2wBDAAgGBgcGBQgHBwcJCQgKDBQNDAsLDBkSEw8UHRofHh0aHBwgJC4nICIsIxwcKDcpLDAxNDQ0Hyc5PTgyPC4zNDL/2wBDAQkJCQwLDBgNDRgyIRwhMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjL/wAARCAABAAEDASIAAhEBAxEB/8QAFQABAQAAAAAAAAAAAAAAAAAAAAb/xAAUEAEAAAAAAAAAAAAAAAAAAAAA/8QAFQEBAQAAAAAAAAAAAAAAAAAAAAX/xAAUEQEAAAAAAAAAAAAAAAAAAAAA/9oADAMBAAIRAxEAPwCdABmX/9k="
	data, err := base64.StdEncoding.DecodeString(tinyJPEG)
	require.NoError(t, err)
	return data
}