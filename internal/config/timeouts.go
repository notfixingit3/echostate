package config

import "time"

// GathererTimeout resolves per-gatherer scan timeouts from live settings.
func GathererTimeout(profile string) time.Duration {
	settings := GetSettings()
	switch profile {
	case "ct":
		sec := settings.CTHTTPTimeoutSec + 20
		if sec <= 20 {
			sec = defaultSettings.CTHTTPTimeoutSec + 20
		}
		return time.Duration(sec) * time.Second
	case "traceroute":
		sec := settings.TracerouteTimeoutSec
		if sec <= 0 {
			sec = defaultSettings.TracerouteTimeoutSec
		}
		return time.Duration(sec) * time.Second
	case "screenshot":
		sec := settings.ScreenshotTimeoutSec
		if sec <= 0 {
			sec = defaultSettings.ScreenshotTimeoutSec
		}
		return time.Duration(sec) * time.Second
	default:
		sec := settings.DefaultGathererTimeoutSec
		if sec <= 0 {
			sec = defaultSettings.DefaultGathererTimeoutSec
		}
		return time.Duration(sec) * time.Second
	}
}

// ScannerHTTPTimeout returns the HTTP client timeout for in-scan fetches.
func ScannerHTTPTimeout() time.Duration {
	sec := GetSettings().ScannerHTTPTimeoutSec
	if sec <= 0 {
		sec = defaultSettings.ScannerHTTPTimeoutSec
	}
	return time.Duration(sec) * time.Second
}

// EnrichmentHTTPTimeout returns the default HTTP timeout for enrichment providers.
func EnrichmentHTTPTimeout() time.Duration {
	sec := GetSettings().EnrichmentHTTPTimeoutSec
	if sec <= 0 {
		sec = defaultSettings.EnrichmentHTTPTimeoutSec
	}
	return time.Duration(sec) * time.Second
}

// WaybackHTTPTimeout returns the HTTP timeout for Internet Archive CDX queries.
func WaybackHTTPTimeout() time.Duration {
	sec := GetSettings().WaybackHTTPTimeoutSec
	if sec <= 0 {
		sec = defaultSettings.WaybackHTTPTimeoutSec
	}
	return time.Duration(sec) * time.Second
}