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

func queryVirusTotal(ctx context.Context, apiKey string, payload map[string]any) (map[string]any, error) {
	host := resolveHost(payload)
	if host == "" {
		return nil, nil
	}

	endpoint := fmt.Sprintf(
		"https://www.virustotal.com/api/v3/domains/%s/resolutions",
		url.PathEscape(host),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-apikey", apiKey)
	req.Header.Set("User-Agent", "EchoState/1.0")

	resp, err := httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 256*1024))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("virustotal HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var parsed map[string]any
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}

	data, _ := parsed["data"].([]any)
	records := summarizeVTResolutions(data)
	if len(records) == 0 {
		return nil, nil
	}

	return map[string]any{
		"query":   host,
		"total":   len(records),
		"records": records,
	}, nil
}

func summarizeVTResolutions(items []any) []map[string]any {
	var out []map[string]any
	for i, item := range items {
		if i >= 15 {
			break
		}
		row, ok := item.(map[string]any)
		if !ok {
			continue
		}
		attrs, _ := row["attributes"].(map[string]any)
		if len(attrs) == 0 {
			continue
		}
		entry := map[string]any{
			"ip":         stringField(attrs, "ip_address"),
			"date":       stringField(attrs, "date"),
			"resolver":   stringField(attrs, "resolver"),
		}
		if entry["ip"] == "" {
			continue
		}
		out = append(out, entry)
	}
	return out
}