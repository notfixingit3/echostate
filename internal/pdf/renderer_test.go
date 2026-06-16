package pdf

import (
	"bytes"
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
		Web: map[string]any{
			"title":      "Example Domain",
			"url":        "https://example.com/",
			"copyrights": []string{"© 2026 Example Inc.", "All rights reserved."},
		},
		Errors: []string{"web gatherer timed out"},
	}

	pdfBytes, err := RenderReport(result)
	require.NoError(t, err)
	require.NotEmpty(t, pdfBytes)
	assert.True(t, bytes.HasPrefix(pdfBytes, []byte("%PDF")), "output should start with PDF magic bytes")
}

func TestRenderReport_EmptyData(t *testing.T) {
	result := &models.ScanResult{
		Host:      "empty.test",
		ScannedAt: time.Time{},
	}

	pdfBytes, err := RenderReport(result)
	require.NoError(t, err)
	require.NotEmpty(t, pdfBytes)
	assert.True(t, bytes.HasPrefix(pdfBytes, []byte("%PDF")))
}

func TestRenderReport_NilResult(t *testing.T) {
	_, err := RenderReport(nil)
	require.Error(t, err)
}
