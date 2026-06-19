package enrichment

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func queryRiskIQ(ctx context.Context, apiUser, apiKey string, payload map[string]any) (map[string]any, error) {
	host := strings.TrimSpace(fmt.Sprint(payload["host"]))
	if host == "" {
		return nil, nil
	}

	endpoint := fmt.Sprintf("https://api.passivetotal.org/v2/dns/passive?query=%s",
		url.QueryEscape(host))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(apiUser, apiKey)

	resp, err := httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 256*1024))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("riskiq HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var parsed map[string]any
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}

	results, _ := parsed["results"].([]any)
	return map[string]any{
		"query":   host,
		"total":   parsed["totalRecords"],
		"results": truncateAny(results, 10),
	}, nil
}