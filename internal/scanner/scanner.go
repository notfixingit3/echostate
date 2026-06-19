package scanner

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/notfixingit3/echostate/internal/config"
	"github.com/notfixingit3/echostate/internal/models"
)

// Gatherer is a function that collects a portion of reconnaissance data.
type Gatherer func(ctx context.Context, host string) (key string, value map[string]any, err error)

type timedGatherer struct {
	profile string
	fn      Gatherer
}

// Scanner orchestrates passive reconnaissance tasks.
type Scanner struct {
	browserWSURL string
	gatherers    []timedGatherer
}

// NewScanner builds a scanner with the default set of gatherers.
func NewScanner(browserWSURL string) *Scanner {
	if browserWSURL == "" {
		browserWSURL = "ws://localhost:3000/"
	}
	return &Scanner{
		browserWSURL: browserWSURL,
		gatherers: []timedGatherer{
			{profile: "default", fn: gatherWHOIS},
			{profile: "default", fn: gatherASN},
			{profile: "default", fn: gatherTLS},
			{profile: "default", fn: gatherDNS},
			{profile: "default", fn: gatherFavicon},
			{profile: "default", fn: gatherCrawl},
			{profile: "default", fn: gatherStorage},
			{profile: "ct", fn: gatherCT},
			{profile: "traceroute", fn: gatherTraceroute},
			{profile: "default", fn: newWebGatherer(browserWSURL)},
			{profile: "screenshot", fn: newScreenshotGatherer(browserWSURL)},
		},
	}
}

// Run executes all gatherers concurrently and returns a unified result.
func (s *Scanner) Run(ctx context.Context, host string) (*models.ScanResult, error) {
	normalized := NormalizeHost(host)

	result := &models.ScanResult{
		Host:      normalized,
		ScannedAt: time.Now().UTC(),
		WHOIS:     make(map[string]any),
		ASN:       make(map[string]any),
		Web:       make(map[string]any),
		TLS:       make(map[string]any),
		DNS:       make(map[string]any),
		Favicon:   make(map[string]any),
		Crawl:     make(map[string]any),
		Storage:   make(map[string]any),
		CT:         make(map[string]any),
		Traceroute:  make(map[string]any),
		Screenshot:  make(map[string]any),
		Errors:      []string{},
	}

	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, g := range s.gatherers {
		wg.Add(1)
		go func(g timedGatherer) {
			defer wg.Done()

			gatherCtx, cancel := context.WithTimeout(ctx, config.GathererTimeout(g.profile))
			defer cancel()

			key, value, err := g.fn(gatherCtx, normalized)
			if err != nil {
				mu.Lock()
				result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", key, err))
				if len(value) > 0 {
					mergeGathererResult(result, key, value)
				}
				mu.Unlock()
				return
			}

			mu.Lock()
			assignGathererResult(result, key, value)
			mu.Unlock()
		}(g)
	}

	wg.Wait()
	finalizeScanResult(result)
	return result, nil
}

func assignGathererResult(result *models.ScanResult, key string, value map[string]any) {
	switch key {
	case "whois":
		result.WHOIS = value
	case "asn":
		result.ASN = value
	case "web":
		result.Web = value
	case "tls":
		result.TLS = value
	case "dns":
		result.DNS = value
	case "favicon":
		result.Favicon = value
	case "crawl":
		result.Crawl = value
	case "storage":
		result.Storage = value
	case "ct":
		result.CT = value
	case "traceroute":
		result.Traceroute = value
	case "screenshot":
		result.Screenshot = value
	}
}

func mergeGathererResult(result *models.ScanResult, key string, value map[string]any) {
	switch key {
	case "whois":
		mergeMap(result.WHOIS, value)
	case "asn":
		mergeMap(result.ASN, value)
	case "web":
		mergeMap(result.Web, value)
	case "tls":
		mergeMap(result.TLS, value)
	case "dns":
		mergeMap(result.DNS, value)
	case "favicon":
		mergeMap(result.Favicon, value)
	case "crawl":
		mergeMap(result.Crawl, value)
	case "storage":
		mergeMap(result.Storage, value)
	case "ct":
		mergeMap(result.CT, value)
	case "traceroute":
		mergeMap(result.Traceroute, value)
	case "screenshot":
		mergeMap(result.Screenshot, value)
	}
}

func mergeMap(dst, src map[string]any) {
	for k, v := range src {
		dst[k] = v
	}
}

// Hash computes a deduplication hash of the scan result JSON.
func Hash(result *models.ScanResult) (string, error) {
	return HashForDedup(result)
}