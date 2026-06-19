package scanner

import (
	"testing"

	"github.com/notfixingit3/echostate/internal/models"
)

func TestFinalizeScanResultMergesWebStorage(t *testing.T) {
	t.Parallel()

	result := &models.ScanResult{
		Web: map[string]any{
			"detected_buckets": []any{
				map[string]any{"provider": "s3", "name": "spa-bucket", "url": "https://spa-bucket.s3.amazonaws.com"},
			},
			"provider_hints": []any{
				map[string]any{"provider": "firebase", "pattern": "firebaseapp.com"},
			},
		},
		Storage: map[string]any{
			"buckets": []any{
				map[string]any{"provider": "s3", "name": "static-bucket", "url": "https://static-bucket.s3.amazonaws.com"},
			},
		},
	}

	finalizeScanResult(result)

	buckets := toMapSlice(result.Storage["buckets"])
	if len(buckets) != 2 {
		t.Fatalf("buckets = %d, want 2", len(buckets))
	}

	hints := toMapSlice(result.Storage["provider_hints"])
	if len(hints) != 1 {
		t.Fatalf("provider_hints = %#v", result.Storage["provider_hints"])
	}
}