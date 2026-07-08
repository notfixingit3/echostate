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

type providerHintPattern struct {
	provider string
	re       *regexp.Regexp
}

var bucketPatterns = []bucketPattern{
	{provider: "aws_s3", re: regexp.MustCompile(`(?i)([a-z0-9][a-z0-9.-]{1,61}[a-z0-9])\.s3(?:[.-][a-z0-9-]+)?\.amazonaws\.com`), nameIdx: 1},
	{provider: "aws_s3", re: regexp.MustCompile(`(?i)s3://([a-z0-9][a-z0-9.-]{1,61}[a-z0-9])`), nameIdx: 1},
	{provider: "azure_blob", re: regexp.MustCompile(`(?i)([a-z0-9][a-z0-9-]{1,61}[a-z0-9])\.blob\.core\.windows\.net`), nameIdx: 1},
	{provider: "gcp_storage", re: regexp.MustCompile(`(?i)storage\.googleapis\.com/([a-z0-9][a-z0-9._-]{1,61}[a-z0-9])`), nameIdx: 1},
	{provider: "gcp_storage", re: regexp.MustCompile(`(?i)([a-z0-9][a-z0-9._-]{1,61}[a-z0-9])\.storage\.googleapis\.com`), nameIdx: 1},
	{provider: "digitalocean_spaces", re: regexp.MustCompile(`(?i)([a-z0-9][a-z0-9-]{1,61}[a-z0-9])\.(?:[a-z0-9-]+\.)?digitaloceanspaces\.com`), nameIdx: 1},
}

var providerHintPatterns = []providerHintPattern{
	{provider: "firebase", re: regexp.MustCompile(`(?i)firebase(app|io)?\.googleapis\.com|firebaseio\.com`)},
	{provider: "supabase", re: regexp.MustCompile(`(?i)supabase\.co|supabase\.in`)},
	{provider: "cloudfront", re: regexp.MustCompile(`(?i)[a-z0-9]+\.cloudfront\.net`)},
	{provider: "fastly", re: regexp.MustCompile(`(?i)[a-z0-9-]+\.fastly\.net|fastlylb\.net`)},
	{provider: "vercel", re: regexp.MustCompile(`(?i)\.vercel\.app|vercel-insights\.com`)},
	{provider: "netlify", re: regexp.MustCompile(`(?i)\.netlify\.app|netlify\.com`)},
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

	return "storage", buildStorageIntel(string(body)), nil
}

func buildStorageIntel(content string) map[string]any {
	result := map[string]any{
		"buckets": []any{},
	}
	buckets := detectBuckets(content)
	if len(buckets) > 0 {
		result["buckets"] = buckets
	}
	if hints := detectProviderHints(content); len(hints) > 0 {
		result["provider_hints"] = hints
	}
	return result
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

func detectProviderHints(content string) []map[string]any {
	seen := make(map[string]struct{})
	var hints []map[string]any

	for _, pattern := range providerHintPatterns {
		match := pattern.re.FindString(content)
		if match == "" {
			continue
		}
		key := pattern.provider + ":" + strings.ToLower(match)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		hints = append(hints, map[string]any{
			"provider": pattern.provider,
			"match":    match,
		})
	}

	return hints
}

func mergeBucketLists(existing, extra []map[string]any) []map[string]any {
	seen := make(map[string]struct{})
	var merged []map[string]any

	add := func(items []map[string]any) {
		for _, item := range items {
			provider := strings.ToLower(strings.TrimSpace(fmt.Sprint(item["provider"])))
			name := strings.ToLower(strings.TrimSpace(fmt.Sprint(item["name"])))
			match := strings.TrimSpace(fmt.Sprint(item["match"]))
			key := provider + ":" + name + ":" + match
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			merged = append(merged, item)
		}
	}

	add(existing)
	add(extra)
	return merged
}

func mergeProviderHints(existing, extra []map[string]any) []map[string]any {
	seen := make(map[string]struct{})
	var merged []map[string]any

	add := func(items []map[string]any) {
		for _, item := range items {
			provider := strings.ToLower(strings.TrimSpace(fmt.Sprint(item["provider"])))
			match := strings.TrimSpace(fmt.Sprint(item["match"]))
			key := provider + ":" + match
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			merged = append(merged, item)
		}
	}

	add(existing)
	add(extra)
	return merged
}
