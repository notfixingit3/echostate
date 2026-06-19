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

const rdapHTTPTimeout = 8 * time.Second

var rdapFetchFunc = fetchRDAPDomain

func fetchRDAPDomain(ctx context.Context, domain string) (map[string]any, error) {
	domain = NormalizeHost(domain)
	if domain == "" || net.ParseIP(domain) != nil {
		return nil, fmt.Errorf("rdap applies to domain names")
	}

	endpoint := fmt.Sprintf("https://rdap.org/domain/%s", url.PathEscape(domain))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/rdap+json")
	req.Header.Set("User-Agent", "EchoState/1.0 (+https://github.com/notfixingit3/echostate)")

	client := &http.Client{Timeout: rdapHTTPTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("rdap HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	if err != nil {
		return nil, err
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}

	return parseRDAPDomain(payload), nil
}

func parseRDAPDomain(payload map[string]any) map[string]any {
	result := map[string]any{
		"source": "rdap",
	}

	if ldh, ok := payload["ldhName"].(string); ok && ldh != "" {
		result["domain"] = strings.ToLower(ldh)
	}
	if handle, ok := payload["handle"].(string); ok && handle != "" {
		result["handle"] = handle
	}
	if status, ok := payload["status"].([]any); ok && len(status) > 0 {
		result["status"] = stringifyAnyList(status)
	}
	if events, ok := payload["events"].([]any); ok {
		for _, item := range events {
			event, ok := item.(map[string]any)
			if !ok {
				continue
			}
			action := strings.ToLower(strings.TrimSpace(fmt.Sprint(event["eventAction"])))
			date := strings.TrimSpace(fmt.Sprint(event["eventDate"]))
			if action == "expiration" && date != "" {
				result["expiration_date"] = date
			}
			if action == "registration" && date != "" {
				result["registration_date"] = date
			}
		}
	}
	if nameservers, ok := payload["nameservers"].([]any); ok {
		var hosts []string
		for _, item := range nameservers {
			ns, ok := item.(map[string]any)
			if !ok {
				continue
			}
			if ldh, ok := ns["ldhName"].(string); ok && ldh != "" {
				hosts = append(hosts, strings.TrimSuffix(ldh, "."))
			}
		}
		if len(hosts) > 0 {
			result["name_servers"] = hosts
		}
	}

	for _, entity := range rdapEntities(payload["entities"]) {
		roles := stringifyAnyList(entity["roles"])
		roleJoined := strings.ToLower(strings.Join(roles, ","))
		vcard := rdapVcardMap(entity["vcardArray"])
		if strings.Contains(roleJoined, "registrar") {
			if org := vcard["org"]; org != "" {
				result["registrar"] = org
			}
			if email := vcard["email"]; email != "" {
				result["registrar_email"] = email
			}
		}
		if strings.Contains(roleJoined, "registrant") || strings.Contains(roleJoined, "administrative") {
			if org := vcard["org"]; org != "" {
				result["registrant_org"] = org
			}
			if email := vcard["email"]; email != "" {
				result["registrant_email"] = email
			}
		}
		if strings.Contains(roleJoined, "abuse") {
			if email := vcard["email"]; email != "" {
				result["abuse_email"] = email
			}
		}
	}

	if len(result) <= 1 {
		return nil
	}
	return result
}

func rdapEntities(raw any) []map[string]any {
	items, ok := raw.([]any)
	if !ok {
		return nil
	}
	var out []map[string]any
	for _, item := range items {
		if entity, ok := item.(map[string]any); ok {
			out = append(out, entity)
		}
	}
	return out
}

func rdapVcardMap(raw any) map[string]string {
	arr, ok := raw.([]any)
	if !ok || len(arr) < 2 {
		return nil
	}
	rows, ok := arr[1].([]any)
	if !ok {
		return nil
	}

	out := map[string]string{}
	for _, row := range rows {
		cols, ok := row.([]any)
		if !ok || len(cols) < 4 {
			continue
		}
		name := strings.ToLower(strings.TrimSpace(fmt.Sprint(cols[0])))
		value := strings.TrimSpace(fmt.Sprint(cols[3]))
		if value == "" {
			continue
		}
		switch name {
		case "fn", "org":
			if out["org"] == "" {
				out["org"] = value
			}
		case "email":
			out["email"] = strings.TrimPrefix(strings.ToLower(value), "mailto:")
		}
	}
	return out
}

func stringifyAnyList(raw any) []string {
	items, ok := raw.([]any)
	if !ok {
		return nil
	}
	var out []string
	for _, item := range items {
		text := strings.TrimSpace(fmt.Sprint(item))
		if text != "" {
			out = append(out, text)
		}
	}
	return out
}

func whoisNeedsRdapFallback(data map[string]any) bool {
	if len(data) == 0 {
		return true
	}
	registrar := strings.TrimSpace(fmt.Sprint(data["registrar"]))
	domain := strings.TrimSpace(fmt.Sprint(data["domain"]))
	expiration := strings.TrimSpace(fmt.Sprint(data["expiration_date"]))
	if registrar == "" || registrar == "<nil>" {
		return true
	}
	if domain == "" && expiration == "" {
		return true
	}
	return false
}

func mergeWhoisRdap(dst, rdap map[string]any) {
	if len(rdap) == 0 {
		return
	}
	for key, value := range rdap {
		if key == "source" {
			dst["rdap_source"] = value
			continue
		}
		current := strings.TrimSpace(fmt.Sprint(dst[key]))
		if current == "" || current == "<nil>" {
			dst[key] = value
		}
	}
	dst["rdap"] = rdap
}