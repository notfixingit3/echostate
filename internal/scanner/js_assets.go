package scanner

import (
	"net/url"
	"regexp"
	"sort"
	"strings"
)

type jsAsset struct {
	URL  string `json:"url"`
	Hint string `json:"hint,omitempty"`
}

type jsAssetIntel struct {
	Assets []jsAsset `json:"assets"`
	Hints  []string  `json:"hints"`
}

// jsAssetIntelJS collects script and stylesheet URLs plus passive library hints.
const jsAssetIntelJS = `(() => {
	const assets = [];
	const hints = new Set();
	const seen = new Set();

	const addAsset = (raw) => {
		const url = String(raw || '').trim();
		if (!url || seen.has(url)) return;
		seen.add(url);
		assets.push({ url });
	};

	const addHint = (value) => {
		const trimmed = String(value || '').trim();
		if (trimmed) hints.add(trimmed);
	};

	const patterns = [
		[/jquery[.-]?(\d[\d.]*)/i, 'lib: jQuery'],
		[/react(?:\.production)?(?:\.min)?\.js/i, 'framework: React'],
		[/vue(?:\.runtime)?(?:\.global)?/i, 'framework: Vue.js'],
		[/angular(?:\.min)?\.js/i, 'framework: Angular'],
		[/bootstrap(?:\.min)?\.(?:js|css)/i, 'ui: Bootstrap'],
		[/tailwind/i, 'ui: Tailwind CSS'],
		[/lodash/i, 'lib: Lodash'],
		[/moment(?:\.min)?\.js/i, 'lib: Moment.js'],
		[/chart\.js/i, 'lib: Chart.js'],
		[/d3(?:\.v\d+)?(?:\.min)?\.js/i, 'lib: D3.js'],
		[/gsap/i, 'lib: GSAP'],
		[/webpack/i, 'bundler: Webpack'],
		[/vite/i, 'bundler: Vite'],
		[/gatsby/i, 'framework: Gatsby'],
		[/nuxt/i, 'framework: Nuxt'],
		[/cdn\.jsdelivr\.net/i, 'cdn: jsDelivr'],
		[/cdnjs\.cloudflare\.com/i, 'cdn: cdnjs'],
		[/unpkg\.com/i, 'cdn: unpkg'],
		[/googleapis\.com/i, 'cdn: Google APIs'],
		[/cloudflare\.com/i, 'cdn: Cloudflare'],
		[/hotjar/i, 'analytics: Hotjar'],
		[/googletagmanager/i, 'analytics: Google Tag Manager'],
		[/google-analytics/i, 'analytics: Google Analytics'],
		[/segment\.com/i, 'analytics: Segment'],
		[/hubspot/i, 'marketing: HubSpot'],
		[/stripe/i, 'payments: Stripe'],
		[/recaptcha/i, 'security: reCAPTCHA'],
	];

	for (const el of document.querySelectorAll('script[src], link[rel="stylesheet"][href]')) {
		const raw = el.src || el.href;
		addAsset(raw);
		const lower = String(raw || '').toLowerCase();
		for (const [pattern, label] of patterns) {
			if (pattern.test(lower)) addHint(label);
		}
	}

	return {
		assets: assets.sort((a, b) => a.url.localeCompare(b.url)),
		hints: Array.from(hints).sort(),
	};
})()`

var jsAssetPathPatterns = []struct {
	re    *regexp.Regexp
	label string
}{
	{regexp.MustCompile(`(?i)jquery[.-]?(\d[\d.]*)`), "lib: jQuery"},
	{regexp.MustCompile(`(?i)react(?:\.production)?(?:\.min)?\.js`), "framework: React"},
	{regexp.MustCompile(`(?i)vue(?:\.runtime)?`), "framework: Vue.js"},
	{regexp.MustCompile(`(?i)angular(?:\.min)?\.js`), "framework: Angular"},
	{regexp.MustCompile(`(?i)bootstrap(?:\.min)?\.(?:js|css)`), "ui: Bootstrap"},
	{regexp.MustCompile(`(?i)webpack`), "bundler: Webpack"},
	{regexp.MustCompile(`(?i)/_next/static/`), "framework: Next.js"},
	{regexp.MustCompile(`(?i)cdn\.shopify\.com`), "platform: Shopify"},
}

func normalizeJSAssets(intel jsAssetIntel) ([]map[string]any, []string) {
	seenHints := make(map[string]struct{})
	for _, hint := range intel.Hints {
		hint = strings.TrimSpace(hint)
		if hint != "" {
			seenHints[hint] = struct{}{}
		}
	}

	var assets []map[string]any
	seenURL := make(map[string]struct{})

	for _, item := range intel.Assets {
		raw := strings.TrimSpace(item.URL)
		if raw == "" {
			continue
		}
		if _, ok := seenURL[raw]; ok {
			continue
		}
		seenURL[raw] = struct{}{}

		entry := map[string]any{"url": raw}
		if hint := hintForAssetURL(raw); hint != "" {
			entry["hint"] = hint
			seenHints[hint] = struct{}{}
		}
		assets = append(assets, entry)
	}

	hints := make([]string, 0, len(seenHints))
	for hint := range seenHints {
		hints = append(hints, hint)
	}
	sort.Strings(hints)

	if len(assets) > 40 {
		assets = assets[:40]
	}

	return assets, hints
}

func hintForAssetURL(raw string) string {
	lower := strings.ToLower(raw)
	if parsed, err := url.Parse(raw); err == nil && parsed.Host != "" {
		lower += " " + strings.ToLower(parsed.Host) + strings.ToLower(parsed.Path)
	}
	for _, pattern := range jsAssetPathPatterns {
		if pattern.re.MatchString(lower) {
			return pattern.label
		}
	}
	return ""
}