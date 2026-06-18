package scanner

import (
	"bufio"
	"context"
	"encoding/xml"
	"fmt"
	"regexp"
	"strings"
)

var sitemapLocPattern = regexp.MustCompile(`(?i)<loc>\s*([^<\s]+)\s*</loc>`)

func gatherCrawl(ctx context.Context, host string) (string, map[string]any, error) {
	host = NormalizeHost(host)
	if host == "" {
		return "crawl", nil, fmt.Errorf("empty host")
	}

	result := map[string]any{}
	var errors []string

	if robots, err := fetchHostResource(ctx, host, "/robots.txt", 256*1024); err == nil {
		result["robots"] = parseRobotsTxt(string(robots))
	} else {
		errors = append(errors, "robots.txt: "+err.Error())
	}

	sitemapURLs := []string{"/sitemap.xml"}
	if robots, ok := result["robots"].(map[string]any); ok {
		if raw, ok := robots["sitemaps"].([]string); ok {
			for _, sm := range raw {
				sitemapURLs = append(sitemapURLs, sm)
			}
		}
	}

	seenSitemaps := make(map[string]struct{})
	var allURLs []string
	for _, sitemapPath := range sitemapURLs {
		sitemapPath = normalizeFaviconPath(host, sitemapPath)
		if sitemapPath == "" {
			continue
		}
		if _, ok := seenSitemaps[sitemapPath]; ok {
			continue
		}
		seenSitemaps[sitemapPath] = struct{}{}

		body, err := fetchHostResource(ctx, host, sitemapPath, 1024*1024)
		if err != nil {
			errors = append(errors, fmt.Sprintf("sitemap %s: %v", sitemapPath, err))
			continue
		}

		urls := parseSitemap(string(body))
		if len(urls) == 0 {
			continue
		}
		allURLs = append(allURLs, urls...)
	}

	if len(allURLs) > 0 {
		if len(allURLs) > 100 {
			result["sitemap_url_count"] = len(allURLs)
			allURLs = allURLs[:100]
		}
		result["sitemap_urls"] = allURLs
	}

	if len(result) == 0 {
		if len(errors) > 0 {
			return "crawl", map[string]any{"errors": errors}, fmt.Errorf("crawl failed for %s", host)
		}
		return "crawl", nil, fmt.Errorf("no crawl data found for %s", host)
	}

	if len(errors) > 0 {
		result["errors"] = errors
	}

	return "crawl", result, nil
}

func parseRobotsTxt(content string) map[string]any {
	disallow := []string{}
	allow := []string{}
	sitemaps := []string{}

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		lower := strings.ToLower(line)
		switch {
		case strings.HasPrefix(lower, "disallow:"):
			value := strings.TrimSpace(line[len("disallow:"):])
			if value != "" {
				disallow = append(disallow, value)
			}
		case strings.HasPrefix(lower, "allow:"):
			value := strings.TrimSpace(line[len("allow:"):])
			if value != "" {
				allow = append(allow, value)
			}
		case strings.HasPrefix(lower, "sitemap:"):
			value := strings.TrimSpace(line[len("sitemap:"):])
			if value != "" {
				sitemaps = append(sitemaps, value)
			}
		}
	}

	return map[string]any{
		"disallow": disallow,
		"allow":    allow,
		"sitemaps": sitemaps,
	}
}

func parseSitemap(content string) []string {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil
	}

	if strings.Contains(content, "<urlset") || strings.Contains(content, "<sitemapindex") {
		if urls := parseSitemapXML(content); len(urls) > 0 {
			return urls
		}
	}

	var urls []string
	for _, match := range sitemapLocPattern.FindAllStringSubmatch(content, -1) {
		if len(match) > 1 {
			urls = append(urls, strings.TrimSpace(match[1]))
		}
	}
	return uniqueStrings(urls)
}

func parseSitemapXML(content string) []string {
	type urlSet struct {
		URLs []struct {
			Loc string `xml:"loc"`
		} `xml:"url"`
	}
	type sitemapIndex struct {
		Sitemaps []struct {
			Loc string `xml:"loc"`
		} `xml:"sitemap"`
	}

	var urls []string

	var set urlSet
	if err := xml.NewDecoder(strings.NewReader(content)).Decode(&set); err == nil {
		for _, item := range set.URLs {
			if item.Loc != "" {
				urls = append(urls, strings.TrimSpace(item.Loc))
			}
		}
		if len(urls) > 0 {
			return uniqueStrings(urls)
		}
	}

	var index sitemapIndex
	if err := xml.NewDecoder(strings.NewReader(content)).Decode(&index); err == nil {
		for _, item := range index.Sitemaps {
			if item.Loc != "" {
				urls = append(urls, strings.TrimSpace(item.Loc))
			}
		}
	}

	return uniqueStrings(urls)
}

func uniqueStrings(items []string) []string {
	seen := make(map[string]struct{})
	var out []string
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}