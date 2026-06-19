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

func queryCensys(ctx context.Context, apiID, apiSecret string, payload map[string]any) (map[string]any, error) {
	result := map[string]any{}

	if host, err := queryCensysHost(ctx, apiID, apiSecret, resolveIP(payload)); err != nil {
		return nil, err
	} else if len(host) > 0 {
		result["host"] = host
	}

	if jarm, err := queryCensysJARM(ctx, apiID, apiSecret, payload); err != nil {
		return nil, err
	} else if len(jarm) > 0 {
		result["jarm_search"] = jarm
	}

	if len(result) == 0 {
		return nil, nil
	}
	return result, nil
}

func queryCensysHost(ctx context.Context, apiID, apiSecret, ip string) (map[string]any, error) {
	ip = strings.TrimSpace(ip)
	if ip == "" || net.ParseIP(ip) == nil {
		return nil, nil
	}

	endpoint := fmt.Sprintf("https://search.censys.io/api/v2/hosts/%s", url.PathEscape(ip))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(apiID, apiSecret)

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
		return nil, fmt.Errorf("censys host HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var parsed map[string]any
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}
	return summarizeCensysHost(parsed), nil
}

func summarizeCensysHost(parsed map[string]any) map[string]any {
	result, _ := parsed["result"].(map[string]any)
	if len(result) == 0 {
		return nil
	}

	out := map[string]any{
		"ip":       stringField(result, "ip"),
		"location": result["location"],
		"services": summarizeCensysServices(result["services"]),
	}
	if autonomous, ok := result["autonomous_system"].(map[string]any); ok {
		out["asn"] = stringField(autonomous, "asn")
		out["as_name"] = stringField(autonomous, "name")
	}
	if dns, ok := result["dns"].(map[string]any); ok {
		out["reverse_dns"] = dns["reverse_dns"]
	}
	return out
}

func summarizeCensysServices(raw any) []map[string]any {
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
			"port":        service["port"],
			"transport":   stringField(service, "transport_protocol"),
			"service_name": stringField(service, "service_name"),
		}
		if tls, ok := service["tls"].(map[string]any); ok {
			entry["tls_version"] = stringField(tls, "version")
			if jarm, ok := tls["jarm"].(map[string]any); ok {
				entry["jarm"] = stringField(jarm, "fingerprint")
			}
		}
		if banner := stringField(service, "banner"); banner != "" {
			if len(banner) > 160 {
				banner = banner[:160] + "…"
			}
			entry["banner"] = banner
		}
		out = append(out, entry)
	}
	return out
}

func queryCensysJARM(ctx context.Context, apiID, apiSecret string, payload map[string]any) (map[string]any, error) {
	tlsMap, _ := payload["tls"].(map[string]any)
	jarm := strings.TrimSpace(fmt.Sprint(tlsMap["jarm"]))
	if jarm == "" {
		return nil, nil
	}

	endpoint := "https://search.censys.io/api/v2/hosts/search"
	body := map[string]any{
		"q":        fmt.Sprintf("services.jarm.fingerprint: %s", jarm),
		"per_page": 5,
	}
	encoded, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(string(encoded)))
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(apiID, apiSecret)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("censys search HTTP %d", resp.StatusCode)
	}

	var parsed map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, err
	}
	return map[string]any{
		"query":  jarm,
		"result": parsed["result"],
	}, nil
}