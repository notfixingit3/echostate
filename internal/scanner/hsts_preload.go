package scanner

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const hstsPreloadTimeout = 8 * time.Second

var lookupHSTSPreloadFunc = lookupHSTSPreload

func lookupHSTSPreload(ctx context.Context, domain string) (map[string]any, error) {
	domain = NormalizeHost(domain)
	if domain == "" || net.ParseIP(domain) != nil {
		return nil, fmt.Errorf("invalid domain")
	}

	endpoint := fmt.Sprintf(
		"https://hstspreload.org/api/v2/preloadable-status?domain=%s",
		url.QueryEscape(domain),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "EchoState/1.0")

	client := &http.Client{Timeout: hstsPreloadTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("hsts preload HTTP %d", resp.StatusCode)
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}

	result := map[string]any{
		"domain":         domain,
		"status":         strings.TrimSpace(fmt.Sprint(payload["status"])),
		"preload_status": strings.TrimSpace(fmt.Sprint(payload["preloadStatus"])),
	}
	if preloaded := strings.EqualFold(fmt.Sprint(payload["preloadStatus"]), "preloaded"); preloaded {
		result["preloaded"] = true
	}
	return result, nil
}
