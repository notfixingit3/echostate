package scanner

import "github.com/notfixingit3/echostate/internal/models"

// finalizeScanResult merges cross-gatherer intel that depends on rendered page content.
func finalizeScanResult(result *models.ScanResult) {
	if result == nil {
		return
	}
	mergeWebStorageIntel(result)
}

func mergeWebStorageIntel(result *models.ScanResult) {
	web := result.Web
	if len(web) == 0 {
		return
	}

	storage := result.Storage
	if storage == nil {
		storage = map[string]any{}
	}

	existingBuckets := toMapSlice(storage["buckets"])
	webBuckets := toMapSlice(web["detected_buckets"])
	if merged := mergeBucketLists(existingBuckets, webBuckets); len(merged) > 0 {
		storage["buckets"] = merged
	} else if _, ok := storage["buckets"]; !ok {
		storage["buckets"] = []any{}
	}

	existingHints := toMapSlice(storage["provider_hints"])
	webHints := toMapSlice(web["provider_hints"])
	if merged := mergeProviderHints(existingHints, webHints); len(merged) > 0 {
		storage["provider_hints"] = merged
	}

	result.Storage = storage
}

func toMapSlice(raw any) []map[string]any {
	switch items := raw.(type) {
	case []map[string]any:
		return items
	case []any:
		var out []map[string]any
		for _, item := range items {
			if m, ok := item.(map[string]any); ok {
				out = append(out, m)
			}
		}
		return out
	default:
		return nil
	}
}
