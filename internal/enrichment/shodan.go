package enrichment

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
)

func queryShodan(ctx context.Context, apiKey string, payload map[string]any) (map[string]any, error) {
	result := map[string]any{}

	if host, err := queryShodanHost(ctx, apiKey, resolveIP(payload)); err != nil {
		return nil, err
	} else if len(host) > 0 {
		result["host"] = host
	}

	if favicon, err := queryShodanFavicon(ctx, apiKey, payload); err != nil {
		return nil, err
	} else if len(favicon) > 0 {
		result["favicon_search"] = favicon
	}

	if len(result) == 0 {
		return nil, nil
	}
	return result, nil
}

func queryShodanHost(ctx context.Context, apiKey, ip string) (map[string]any, error) {
	ip = strings.TrimSpace(ip)
	if ip == "" || net.ParseIP(ip) == nil {
		return nil, nil
	}

	endpoint := fmt.Sprintf(
		"https://api.shodan.io/shodan/host/%s?key=%s",
		url.PathEscape(ip),
		url.QueryEscape(apiKey),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	resp, err := httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("shodan host HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var parsed map[string]any
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}

	return summarizeShodanHost(parsed), nil
}

func summarizeShodanHost(data map[string]any) map[string]any {
	out := map[string]any{
		"ip":          stringField(data, "ip_str", "ip"),
		"org":         stringField(data, "org"),
		"isp":         stringField(data, "isp"),
		"asn":         stringField(data, "asn"),
		"country":     stringField(data, "country_name"),
		"city":        stringField(data, "city"),
		"hostnames":   data["hostnames"],
		"ports":       data["ports"],
		"tags":        data["tags"],
		"last_update": stringField(data, "last_update"),
	}

	if vulns, ok := data["vulns"].([]any); ok && len(vulns) > 0 {
		out["vuln_count"] = len(vulns)
		out["vulns"] = truncateAny(vulns, 10)
	} else if vulns, ok := data["vulns"].(map[string]any); ok && len(vulns) > 0 {
		names := make([]string, 0, len(vulns))
		for name := range vulns {
			names = append(names, name)
		}
		out["vuln_count"] = len(names)
		if len(names) > 10 {
			names = names[:10]
		}
		out["vulns"] = names
	}

	if services := summarizeShodanServices(data["data"]); len(services) > 0 {
		out["services"] = services
	}
	return out
}

func summarizeShodanServices(raw any) []map[string]any {
	items, ok := raw.([]any)
	if !ok {
		return nil
	}
	var out []map[string]any
	for i, item := range items {
		if i >= 8 {
			break
		}
		service, ok := item.(map[string]any)
		if !ok {
			continue
		}
		entry := map[string]any{
			"port":      service["port"],
			"transport": stringField(service, "transport"),
			"product":   stringField(service, "product"),
			"version":   stringField(service, "version"),
		}
		if shodanMeta, ok := service["_shodan"].(map[string]any); ok {
			entry["module"] = stringField(shodanMeta, "module")
		}
		if banner := stringField(service, "data"); banner != "" {
			if len(banner) > 160 {
				banner = banner[:160] + "…"
			}
			entry["banner"] = banner
		}
		out = append(out, entry)
	}
	return out
}

func queryShodanFavicon(ctx context.Context, apiKey string, payload map[string]any) (map[string]any, error) {
	favicon, _ := payload["favicon"].(map[string]any)
	query := strings.TrimSpace(fmt.Sprint(favicon["mmh3"]))
	if query == "" {
		query = strings.TrimSpace(fmt.Sprint(favicon["shodan"]))
	}
	if query == "" {
		return nil, nil
	}

	endpoint := fmt.Sprintf(
		"https://api.shodan.io/shodan/host/search?key=%s&query=http.favicon.hash:%s",
		url.QueryEscape(apiKey),
		url.QueryEscape(query),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	resp, err := httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("shodan search HTTP %d", resp.StatusCode)
	}

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	return map[string]any{
		"query":   query,
		"total":   body["total"],
		"matches": truncateMatches(body["matches"]),
	}, nil
}

func stringField(data map[string]any, keys ...string) string {
	for _, key := range keys {
		if data == nil {
			return ""
		}
		value := strings.TrimSpace(fmt.Sprint(data[key]))
		if value != "" && value != "<nil>" {
			return value
		}
	}
	return ""
}