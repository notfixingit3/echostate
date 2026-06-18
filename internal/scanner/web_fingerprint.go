package scanner

import (
	"fmt"
	"strings"

	"github.com/chromedp/cdproto/network"
)

var securityHeaderKeys = map[string]struct{}{
	"strict-transport-security": {},
	"content-security-policy":     {},
	"x-frame-options":             {},
	"x-content-type-options":      {},
}

// techHeaderKeys lists response headers promoted into tech_stack fingerprints.
var techHeaderKeys = map[string]struct{}{
	"server":                   {},
	"x-powered-by":             {},
	"x-aspnet-version":         {},
	"x-aspnetmvc-version":      {},
	"x-generator":              {},
	"x-drupal-cache":           {},
	"x-drupal-dynamic-cache":   {},
	"x-varnish":                {},
	"x-runtime":                {},
	"x-version":                {},
	"via":                      {},
	"x-served-by":              {},
	"x-cache":                  {},
}

func headersFromNetwork(raw network.Headers) map[string]string {
	if len(raw) == 0 {
		return nil
	}

	out := make(map[string]string, len(raw))
	for key, value := range raw {
		out[key] = headerValueString(value)
	}
	return out
}

func headerValueString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case []any:
		parts := make([]string, 0, len(v))
		for _, item := range v {
			parts = append(parts, fmt.Sprint(item))
		}
		return strings.Join(parts, ", ")
	default:
		return fmt.Sprint(v)
	}
}

func extractSecurityHeaders(headers map[string]string) map[string]string {
	if len(headers) == 0 {
		return nil
	}

	out := make(map[string]string)
	for key, value := range headers {
		if _, ok := securityHeaderKeys[strings.ToLower(key)]; ok {
			out[strings.ToLower(key)] = value
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func extractTechStackFromHeaders(headers map[string]string) []string {
	if len(headers) == 0 {
		return nil
	}

	seen := make(map[string]struct{})
	var out []string

	for key, value := range headers {
		kl := strings.ToLower(key)
		if _, ok := techHeaderKeys[kl]; ok {
			entry := fmt.Sprintf("%s: %s", key, value)
			if _, dup := seen[entry]; dup {
				continue
			}
			seen[entry] = struct{}{}
			out = append(out, entry)
			continue
		}

		switch kl {
		case "cf-ray":
			entry := "CDN: Cloudflare"
			if _, dup := seen[entry]; !dup {
				seen[entry] = struct{}{}
				out = append(out, entry)
			}
		}
	}

	return out
}

func mergeTechStack(parts ...[]string) []string {
	seen := make(map[string]struct{})
	var out []string

	for _, group := range parts {
		for _, item := range group {
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
	}

	return out
}

const htmlTechHintsJS = `(() => {
	const hints = [];
	const add = (value) => {
		const trimmed = String(value || '').trim();
		if (trimmed) hints.push(trimmed);
	};

	const generator = document.querySelector('meta[name="generator"]');
	if (generator && generator.content) {
		add('generator: ' + generator.content);
	}

	const html = document.documentElement.innerHTML;
	if (html.includes('/wp-content/') || html.includes('/wp-includes/')) {
		add('cms: WordPress');
	}
	if (html.includes('Drupal.settings') || html.includes('/sites/default/files')) {
		add('cms: Drupal');
	}
	if (html.includes('cdn.shopify.com')) {
		add('platform: Shopify');
	}
	if (html.includes('/_next/static/') || html.includes('__NEXT_DATA__')) {
		add('framework: Next.js');
	}

	return hints;
})()`