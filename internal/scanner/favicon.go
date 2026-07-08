package scanner

import (
	"context"
	"encoding/base64"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/twmb/murmur3"
)

var faviconLinkPattern = regexp.MustCompile(`(?i)<link[^>]+rel=["'](?:shortcut )?icon["'][^>]*>`)
var faviconHrefPattern = regexp.MustCompile(`(?i)href=["']([^"']+)["']`)

func gatherFavicon(ctx context.Context, host string) (string, map[string]any, error) {
	host = NormalizeHost(host)
	if host == "" {
		return "favicon", nil, fmt.Errorf("empty host")
	}

	var candidates []string
	candidates = append(candidates, "/favicon.ico")

	if page, err := fetchHostResource(ctx, host, "/", 256*1024); err == nil {
		if href := extractFaviconHref(string(page)); href != "" {
			candidates = append([]string{href}, candidates...)
		}
	}

	seen := make(map[string]struct{})
	for _, candidate := range candidates {
		candidate = normalizeFaviconPath(host, candidate)
		if candidate == "" {
			continue
		}
		if _, ok := seen[candidate]; ok {
			continue
		}
		seen[candidate] = struct{}{}

		body, err := fetchHostResource(ctx, host, candidate, 512*1024)
		if err != nil || len(body) == 0 {
			continue
		}

		b64 := base64.StdEncoding.EncodeToString(body)
		hash := int32(murmur3.Sum32([]byte(b64)))

		return "favicon", map[string]any{
			"url":    candidate,
			"mmh3":   strconv.Itoa(int(hash)),
			"sha256": hashSHA256(body),
			"size":   len(body),
			"shodan": strconv.Itoa(int(hash)),
		}, nil
	}

	return "favicon", nil, fmt.Errorf("no favicon found for %s", host)
}

func extractFaviconHref(page string) string {
	match := faviconLinkPattern.FindString(page)
	if match == "" {
		return ""
	}
	hrefMatch := faviconHrefPattern.FindStringSubmatch(match)
	if len(hrefMatch) < 2 {
		return ""
	}
	return strings.TrimSpace(hrefMatch[1])
}

func normalizeFaviconPath(host, href string) string {
	href = strings.TrimSpace(href)
	if href == "" {
		return ""
	}
	if strings.HasPrefix(href, "data:") {
		return ""
	}
	if strings.HasPrefix(href, "//") {
		return "https:" + href
	}
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return href
	}
	if strings.HasPrefix(href, "/") {
		return href
	}
	return "/" + href
}

func hashSHA256(data []byte) string {
	return hashBytesSHA256(data)
}
