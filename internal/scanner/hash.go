package scanner

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/notfixingit3/echostate/internal/models"
)

// HashForDedup computes a stable SHA256 hash for snapshot deduplication.
// Volatile gatherer fields (screenshot bytes, capture timestamps) are excluded.
func HashForDedup(result *models.ScanResult) (string, error) {
	if result == nil {
		return "", fmt.Errorf("nil scan result")
	}

	copy := *result
	copy.ScannedAt = time.Time{}
	copy.Screenshot = sanitizeScreenshotForHash(copy.Screenshot)
	copy.Traceroute = sanitizeTracerouteForHash(copy.Traceroute)

	data, err := json.Marshal(&copy)
	if err != nil {
		return "", fmt.Errorf("marshal result: %w", err)
	}

	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

func sanitizeScreenshotForHash(screenshot map[string]any) map[string]any {
	if len(screenshot) == 0 {
		return screenshot
	}
	out := make(map[string]any, len(screenshot))
	for key, value := range screenshot {
		switch key {
		case "thumbnail", "captured_at", "blob_id":
			continue
		default:
			out[key] = value
		}
	}
	return out
}

func sanitizeTracerouteForHash(traceroute map[string]any) map[string]any {
	if len(traceroute) == 0 {
		return traceroute
	}
	out := make(map[string]any, len(traceroute))
	for key, value := range traceroute {
		if key == "probed_at" {
			continue
		}
		out[key] = value
	}
	return out
}