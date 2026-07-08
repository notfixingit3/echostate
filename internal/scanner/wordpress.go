package scanner

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"
)

var (
	wpPluginPathRe      = regexp.MustCompile(`(?i)/wp-content/plugins/([a-z0-9_-]+)/`)
	wpThemePathRe       = regexp.MustCompile(`(?i)/wp-content/themes/([a-z0-9_-]+)/`)
	themeStyleVersionRe = regexp.MustCompile(`(?i)Version:\s*([^\n\r*]+)`)
)

// wordpressIntelJS collects WordPress plugin and theme slugs (plus versions when
// present) from asset URLs and page HTML during the browser capture pass.
const wordpressIntelJS = `(() => {
	const collect = (pathPattern) => {
		const items = new Map();

		const add = (raw) => {
			const text = String(raw || '').trim();
			if (!text) return;
			const match = text.match(pathPattern);
			if (!match) return;

			const slug = match[1].toLowerCase();
			let version = '';
			const verMatch = text.match(/[?&]ver=([^&#'"]+)/i);
			if (verMatch) {
				try {
					version = decodeURIComponent(verMatch[1]).trim();
				} catch {
					version = verMatch[1].trim();
				}
			}

			const existing = items.get(slug);
			if (!existing) {
				items.set(slug, { slug, version });
				return;
			}
			if (!existing.version && version) {
				existing.version = version;
			}
		};

		for (const el of document.querySelectorAll('script[src], link[href], img[src], source[src]')) {
			add(el.src || el.href);
		}

		const html = document.documentElement.innerHTML;
		const re = new RegExp(pathPattern.source, pathPattern.flags + 'g');
		let match;
		while ((match = re.exec(html)) !== null) {
			const slug = match[1].toLowerCase();
			if (!items.has(slug)) {
				items.set(slug, { slug, version: '' });
			}
		}

		return Array.from(items.values()).sort((a, b) => a.slug.localeCompare(b.slug));
	};

	return {
		plugins: collect(/\/wp-content\/plugins\/([a-z0-9_-]+)\//i),
		themes: collect(/\/wp-content\/themes\/([a-z0-9_-]+)\//i),
	};
})()`

type wordpressItem struct {
	Slug    string `json:"slug"`
	Version string `json:"version,omitempty"`
}

type wordpressIntel struct {
	Plugins []wordpressItem `json:"plugins"`
	Themes  []wordpressItem `json:"themes"`
}

func normalizeWordPressItems(hits []wordpressItem) []map[string]any {
	if len(hits) == 0 {
		return nil
	}

	merged := make(map[string]wordpressItem, len(hits))
	for _, hit := range hits {
		slug := strings.ToLower(strings.TrimSpace(hit.Slug))
		if slug == "" {
			continue
		}
		version := strings.TrimSpace(hit.Version)
		existing, ok := merged[slug]
		if !ok {
			merged[slug] = wordpressItem{Slug: slug, Version: version}
			continue
		}
		if existing.Version == "" && version != "" {
			existing.Version = version
			merged[slug] = existing
		}
	}

	out := make([]map[string]any, 0, len(merged))
	for _, item := range merged {
		entry := map[string]any{"slug": item.Slug}
		if item.Version != "" {
			entry["version"] = item.Version
		}
		out = append(out, entry)
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i]["slug"].(string) < out[j]["slug"].(string)
	})

	return out
}

func extractWordPressPluginFromURL(raw string) (slug, version string, ok bool) {
	return extractWordPressItemFromURL(raw, wpPluginPathRe)
}

func extractWordPressThemeFromURL(raw string) (slug, version string, ok bool) {
	return extractWordPressItemFromURL(raw, wpThemePathRe)
}

func extractWordPressItemFromURL(raw string, pattern *regexp.Regexp) (slug, version string, ok bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", "", false
	}

	match := pattern.FindStringSubmatch(raw)
	if match == nil {
		return "", "", false
	}

	slug = strings.ToLower(match[1])
	if u, err := url.Parse(raw); err == nil {
		version = strings.TrimSpace(u.Query().Get("ver"))
	}
	return slug, version, true
}

func enrichWordPressThemeVersions(ctx context.Context, pageURL string, themes []map[string]any) {
	if len(themes) == 0 || strings.TrimSpace(pageURL) == "" {
		return
	}

	parsed, err := url.Parse(pageURL)
	if err != nil || parsed.Host == "" {
		return
	}

	scheme := parsed.Scheme
	if scheme == "" {
		scheme = "https"
	}
	base := scheme + "://" + parsed.Host

	client := &http.Client{Timeout: 5 * time.Second}
	for _, theme := range themes {
		version, _ := theme["version"].(string)
		if strings.TrimSpace(version) != "" {
			continue
		}
		slug, _ := theme["slug"].(string)
		slug = strings.TrimSpace(slug)
		if slug == "" {
			continue
		}

		styleURL := base + "/wp-content/themes/" + slug + "/style.css"
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, styleURL, nil)
		if err != nil {
			continue
		}
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		body, err := io.ReadAll(io.LimitReader(resp.Body, 16<<10))
		resp.Body.Close()
		if err != nil || resp.StatusCode >= 400 {
			continue
		}

		if match := themeStyleVersionRe.FindSubmatch(body); len(match) > 1 {
			theme["version"] = strings.TrimSpace(string(match[1]))
		}
	}
}
