package scanner

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

type bucketPattern struct {
	provider string
	re       *regexp.Regexp
	nameIdx  int
}

var bucketPatterns = []bucketPattern{
	{provider: "aws_s3", re: regexp.MustCompile(`(?i)([a-z0-9][a-z0-9.-]{1,61}[a-z0-9])\.s3(?:[.-][a-z0-9-]+)?\.amazonaws\.com`), nameIdx: 1},
	{provider: "aws_s3", re: regexp.MustCompile(`(?i)s3://([a-z0-9][a-z0-9.-]{1,61}[a-z0-9])`), nameIdx: 1},
	{provider: "azure_blob", re: regexp.MustCompile(`(?i)([a-z0-9][a-z0-9-]{1,61}[a-z0-9])\.blob\.core\.windows\.net`), nameIdx: 1},
	{provider: "gcp_storage", re: regexp.MustCompile(`(?i)storage\.googleapis\.com/([a-z0-9][a-z0-9._-]{1,61}[a-z0-9])`), nameIdx: 1},
	{provider: "gcp_storage", re: regexp.MustCompile(`(?i)([a-z0-9][a-z0-9._-]{1,61}[a-z0-9])\.storage\.googleapis\.com`), nameIdx: 1},
	{provider: "digitalocean_spaces", re: regexp.MustCompile(`(?i)([a-z0-9][a-z0-9-]{1,61}[a-z0-9])\.(?:[a-z0-9-]+\.)?digitaloceanspaces\.com`), nameIdx: 1},
}

func gatherStorage(ctx context.Context, host string) (string, map[string]any, error) {
	host = NormalizeHost(host)
	if host == "" {
		return "storage", nil, fmt.Errorf("empty host")
	}

	body, err := fetchHostResource(ctx, host, "/", 1024*1024)
	if err != nil {
		return "storage", nil, fmt.Errorf("fetch homepage: %w", err)
	}

	buckets := detectBuckets(string(body))
	if len(buckets) == 0 {
		return "storage", map[string]any{"buckets": []any{}}, nil
	}

	return "storage", map[string]any{"buckets": buckets}, nil
}

func detectBuckets(content string) []map[string]any {
	seen := make(map[string]struct{})
	var buckets []map[string]any

	for _, pattern := range bucketPatterns {
		for _, match := range pattern.re.FindAllStringSubmatch(content, -1) {
			if len(match) <= pattern.nameIdx {
				continue
			}
			name := strings.ToLower(strings.TrimSpace(match[pattern.nameIdx]))
			if name == "" {
				continue
			}
			key := pattern.provider + ":" + name
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}

			buckets = append(buckets, map[string]any{
				"provider": pattern.provider,
				"name":     name,
				"match":    match[0],
			})
		}
	}

	return buckets
}