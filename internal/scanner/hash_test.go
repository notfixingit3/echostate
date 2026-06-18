package scanner

import (
	"testing"
	"time"

	"github.com/notfixingit3/echostate/internal/models"
)

func TestHashForDedup_IgnoresScreenshotThumbnail(t *testing.T) {
	base := &models.ScanResult{
		Host:      "example.com",
		ScannedAt: time.Now().UTC(),
		WHOIS:     map[string]any{"domain": "example.com"},
		Screenshot: map[string]any{
			"url":         "https://example.com",
			"thumbnail":   "abc",
			"captured_at": time.Now().UTC().Format(time.RFC3339),
		},
	}

	other := &models.ScanResult{
		Host:      "example.com",
		ScannedAt: base.ScannedAt.Add(2 * time.Hour),
		WHOIS:     map[string]any{"domain": "example.com"},
		Screenshot: map[string]any{
			"url":         "https://example.com",
			"thumbnail":   "def",
			"captured_at": time.Now().UTC().Add(3 * time.Hour).Format(time.RFC3339),
		},
	}

	hash1, err := HashForDedup(base)
	if err != nil {
		t.Fatalf("HashForDedup() error = %v", err)
	}
	hash2, err := HashForDedup(other)
	if err != nil {
		t.Fatalf("HashForDedup() error = %v", err)
	}
	if hash1 != hash2 {
		t.Fatalf("hashes differ: %s vs %s", hash1, hash2)
	}
}