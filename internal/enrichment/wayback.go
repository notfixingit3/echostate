package enrichment

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/notfixingit3/echostate/internal/config"
)

func queryWayback(ctx context.Context, payload map[string]any) (map[string]any, error) {
	host := resolveHost(payload)
	if host == "" {
		return nil, nil
	}

	endpoint := fmt.Sprintf(
		"https://web.archive.org/cdx/search/cdx?url=%s&output=json&limit=30&collapse=urlkey&fl=timestamp,original,statuscode,mimetype",
		url.QueryEscape(host+"/*"),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "EchoState/1.0 (+https://github.com/notfixingit3/echostate)")

	client := &http.Client{Timeout: config.WaybackHTTPTimeout()}
	resp, err := client.Do(req)
	if err != nil {
		return nil, friendlyServiceError("wayback", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("wayback HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var rows [][]any
	if err := json.Unmarshal(body, &rows); err != nil {
		return nil, err
	}
	if len(rows) <= 1 {
		return nil, nil
	}

	var urls []map[string]any
	for i, row := range rows[1:] {
		if i >= 25 {
			break
		}
		if len(row) < 2 {
			continue
		}
		entry := map[string]any{
			"timestamp": fmt.Sprint(row[0]),
			"url":       fmt.Sprint(row[1]),
		}
		if len(row) > 2 {
			entry["status"] = fmt.Sprint(row[2])
		}
		if len(row) > 3 {
			entry["mimetype"] = fmt.Sprint(row[3])
		}
		urls = append(urls, entry)
	}
	if len(urls) == 0 {
		return nil, nil
	}

	return map[string]any{
		"query": host,
		"total": len(urls),
		"urls":  urls,
	}, nil
}
