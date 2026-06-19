package scanner

import (
	"context"
	"strings"
)

var optionalCrawlPaths = []struct {
	key  string
	path string
}{
	{key: "humans_txt", path: "/humans.txt"},
	{key: "ads_txt", path: "/ads.txt"},
	{key: "app_ads_txt", path: "/app-ads.txt"},
}

func enrichOptionalCrawlFiles(ctx context.Context, host string, result map[string]any) {
	for _, item := range optionalCrawlPaths {
		body, err := fetchHostResource(ctx, host, item.path, 256*1024)
		if err != nil || len(body) == 0 {
			continue
		}
		content := strings.TrimSpace(string(body))
		if content == "" {
			continue
		}
		result[item.key] = map[string]any{
			"source": item.path,
			"lines":  splitNonEmptyLines(content),
			"text":   truncateString(content, 4096),
		}
	}
}

func splitNonEmptyLines(content string) []string {
	var lines []string
	for _, line := range strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		lines = append(lines, line)
	}
	return lines
}