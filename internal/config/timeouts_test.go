package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGathererTimeout_UsesLiveSettings(t *testing.T) {
	UpdateSettings(SystemSettings{
		DefaultGathererTimeoutSec: 33,
		TracerouteTimeoutSec:      55,
	})
	t.Cleanup(func() { UpdateSettings(defaultSettings) })

	assert.Equal(t, 33*time.Second, GathererTimeout("default"))
	assert.Equal(t, 55*time.Second, GathererTimeout("traceroute"))
}

func TestNormalizeSettings_HTTPTimeouts(t *testing.T) {
	normalized := normalizeSettings(SystemSettings{
		ScannerHTTPTimeoutSec:    0,
		EnrichmentHTTPTimeoutSec: 0,
		WaybackHTTPTimeoutSec:    0,
	})
	assert.Equal(t, 8, normalized.ScannerHTTPTimeoutSec)
	assert.Equal(t, 12, normalized.EnrichmentHTTPTimeoutSec)
	assert.Equal(t, 30, normalized.WaybackHTTPTimeoutSec)
}

func TestNormalizeSettings_LogLevel(t *testing.T) {
	normalized := normalizeSettings(SystemSettings{LogLevel: ""})
	assert.Equal(t, "info", normalized.LogLevel)

	normalized = normalizeSettings(SystemSettings{LogLevel: "debug"})
	assert.Equal(t, "debug", normalized.LogLevel)
}
