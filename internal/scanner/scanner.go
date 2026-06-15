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
	gatherers []Gatherer
}

// NewScanner builds a scanner with the default set of gatherers.
func NewScanner() *Scanner {
	return &Scanner{
		gatherers: []Gatherer{
			gatherWHOIS,
			gatherASN,
			gatherWeb,
		},
	}
}

// Run executes all gatherers concurrently and returns a unified result.
func (s *Scanner) Run(ctx context.Context, host string) (*models.ScanResult, error) {
	result := &models.ScanResult{
		Host:      host,
		ScannedAt: time.Now().UTC(),
		WHOIS:     make(map[string]any),
		ASN:       make(map[string]any),
		Web:       make(map[string]any),
		Errors:    []string{},
	}

	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, g := range s.gatherers {
		wg.Add(1)
		go func(gatherer Gatherer) {
			defer wg.Done()

			key, value, err := gatherer(ctx, host)
			if err != nil {
				mu.Lock()
				result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", key, err))
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
			}
			mu.Unlock()
		}(g)
	}

	wg.Wait()
	return result, nil
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

func gatherWHOIS(ctx context.Context, host string) (string, map[string]any, error) {
	// TODO: implement WHOIS lookup with github.com/likexian/whois
	return "whois", map[string]any{"status": "pending"}, nil
}

func gatherASN(ctx context.Context, host string) (string, map[string]any, error) {
	// TODO: implement passive BGP/ASN lookup
	return "asn", map[string]any{"status": "pending"}, nil
}

func gatherWeb(ctx context.Context, host string) (string, map[string]any, error) {
	// TODO: implement chromedp DOM scraping for copyright data
	return "web", map[string]any{"status": "pending"}, nil
}
