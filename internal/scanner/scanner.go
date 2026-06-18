package scanner

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/notfixingit3/echostate/internal/models"
)

// Gatherer is a function that collects a portion of reconnaissance data.
type Gatherer func(ctx context.Context, host string) (key string, value map[string]any, err error)

// Scanner orchestrates passive reconnaissance tasks.
type Scanner struct {
	browserWSURL string
	gatherers    []Gatherer
}

// NewScanner builds a scanner with the default set of gatherers.
func NewScanner(browserWSURL string) *Scanner {
	if browserWSURL == "" {
		browserWSURL = "ws://localhost:3000/"
	}
	return &Scanner{
		browserWSURL: browserWSURL,
		gatherers: []Gatherer{
			gatherWHOIS,
			gatherASN,
			gatherTLS,
			gatherDNS,
			newWebGatherer(browserWSURL),
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
		Errors:    []string{},
	}

	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, g := range s.gatherers {
		wg.Add(1)
		go func(gatherer Gatherer) {
			defer wg.Done()

			gatherCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			key, value, err := gatherer(gatherCtx, normalized)
			if err != nil {
				mu.Lock()
				result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", key, err))
				if len(value) > 0 {
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
					}
				}
				mu.Unlock()
				return
			}

			mu.Lock()
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
			}
			mu.Unlock()
		}(g)
	}

	wg.Wait()
	return result, nil
}

func mergeMap(dst, src map[string]any) {
	for k, v := range src {
		dst[k] = v
	}
}

// Hash computes a SHA256 hash of the scan result JSON.
func Hash(result *models.ScanResult) (string, error) {
	data, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("marshal result: %w", err)
	}

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}


