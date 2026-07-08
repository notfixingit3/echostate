package scanner

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/notfixingit3/echostate/internal/config"
)

func newHTTPClient(timeout time.Duration) *http.Client {
	if timeout <= 0 {
		timeout = config.ScannerHTTPTimeout()
	}
	return &http.Client{
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}
}

func httpGet(ctx context.Context, rawURL string, maxBytes int64) ([]byte, string, error) {
	return httpGetTimeout(ctx, rawURL, maxBytes, config.ScannerHTTPTimeout())
}

func httpGetTimeout(ctx context.Context, rawURL string, maxBytes int64, timeout time.Duration) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", "EchoState/1.0 (+https://github.com/notfixingit3/echostate)")

	resp, err := newHTTPClient(timeout).Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	reader := io.LimitReader(resp.Body, maxBytes)
	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, "", err
	}

	finalURL := resp.Request.URL.String()
	if resp.Request.URL.Host != "" {
		finalURL = resp.Request.URL.String()
	}

	return body, finalURL, nil
}

func fetchHostResource(ctx context.Context, host, path string, maxBytes int64) ([]byte, error) {
	host = NormalizeHost(host)
	if host == "" {
		return nil, fmt.Errorf("empty host")
	}

	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		body, _, err := httpGet(ctx, path, maxBytes)
		return body, err
	}

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	body, _, err := httpGet(ctx, "https://"+host+path, maxBytes)
	if err == nil {
		return body, nil
	}

	body, _, err = httpGet(ctx, "http://"+host+path, maxBytes)
	return body, err
}
