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

func queryHIBP(ctx context.Context, apiKey string, payload map[string]any) (map[string]any, error) {
	emails := extractEmails(payload)
	if len(emails) == 0 {
		return nil, nil
	}

	results := make([]map[string]any, 0, len(emails))
	for _, email := range emails {
		breaches, err := hibpBreaches(ctx, apiKey, email)
		if err != nil {
			return nil, err
		}
		results = append(results, map[string]any{
			"email":    email,
			"breaches": breaches,
			"pwned":    len(breaches) > 0,
		})
	}

	return map[string]any{
		"checked": len(results),
		"results": results,
	}, nil
}

func hibpBreaches(ctx context.Context, apiKey, email string) ([]map[string]any, error) {
	endpoint := fmt.Sprintf("https://haveibeenpwned.com/api/v3/breachedaccount/%s?truncateResponse=true",
		url.PathEscape(email))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("hibp-api-key", apiKey)
	req.Header.Set("User-Agent", "EchoState/1.0")

	resp, err := httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("hibp HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var breaches []map[string]any
	if err := json.Unmarshal(body, &breaches); err != nil {
		return nil, err
	}
	return breaches, nil
}