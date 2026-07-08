package scanner

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

const defaultBrowserWSURL = "ws://localhost:3000/"

type pageCapture struct {
	title            string
	text             string
	html             string
	location         string
	headerURL        string
	headers          map[string]string
	htmlHints        []string
	meta             map[string]any
	cookieNames      []string
	jsAssets         []map[string]any
	jsHints          []string
	wordpressPlugins []map[string]any
	wordpressThemes  []map[string]any
}

type documentResponse struct {
	url     string
	headers map[string]string
}

// newWebGatherer returns a Gatherer that connects to a browserless Chrome
// instance at browserWSURL and extracts the landing page title, final URL,
// response headers from the document load, and copyright snippets from visible text.
func newWebGatherer(browserWSURL string) Gatherer {
	if browserWSURL == "" {
		browserWSURL = defaultBrowserWSURL
	}

	return func(ctx context.Context, host string) (string, map[string]any, error) {
		host = NormalizeHost(host)
		if host == "" {
			return "web", map[string]any{"error": "empty host"}, errors.New("empty host")
		}

		ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()

		resolvedWS := resolveBrowserWSURL(ctx, browserWSURL)
		allocCtx, allocCancel := chromedp.NewRemoteAllocator(ctx, resolvedWS)
		defer allocCancel()

		var capture pageCapture
		scrapeErr := scrapePage(allocCtx, "https://"+host, &capture)
		if scrapeErr != nil {
			capture = pageCapture{}
			scrapeErr = scrapePage(allocCtx, "http://"+host, &capture)
		}

		headers := capture.headers
		headerURL := capture.headerURL
		if len(headers) == 0 {
			fallbackURL := capture.location
			if fallbackURL == "" {
				fallbackURL = "https://" + host
			}
			headers, headerURL = fetchDocumentHeaders(ctx, fallbackURL)
			if len(headers) == 0 && !strings.HasPrefix(fallbackURL, "http://") {
				headers, headerURL = fetchDocumentHeaders(ctx, "http://"+host)
			}
		}

		securityHeaders := extractSecurityHeaders(headers)
		techStack := mergeTechStack(
			extractTechStackFromHeaders(headers),
			capture.htmlHints,
			capture.jsHints,
		)

		redirectChain, redirectFinal := captureRedirectChain(ctx, host)
		finalURL := capture.location
		if finalURL == "" {
			finalURL = redirectFinal
		}

		webData := map[string]any{
			"title":            capture.title,
			"url":              finalURL,
			"header_url":       headerURL,
			"copyrights":       extractCopyrights(capture.text),
			"headers":          headers,
			"security_headers": securityHeaders,
			"tech_stack":       techStack,
		}
		if len(redirectChain) > 0 {
			webData["redirect_chain"] = redirectChain
		}
		if emails := extractEmailsFromText(capture.text); len(emails) > 0 {
			webData["contact_emails"] = emails
		}
		if len(capture.meta) > 0 {
			webData["meta"] = capture.meta
		}
		if names := capture.cookieNames; len(names) > 0 {
			webData["cookie_names"] = names
		}
		if hstsHeader := securityHeaders["strict-transport-security"]; strings.TrimSpace(hstsHeader) != "" {
			if preload, err := lookupHSTSPreloadFunc(ctx, host); err == nil && len(preload) > 0 {
				webData["hsts_preload"] = preload
			}
		}
		if capture.html != "" {
			if buckets := detectBuckets(capture.html); len(buckets) > 0 {
				webData["detected_buckets"] = buckets
			}
			if hints := detectProviderHints(capture.html); len(hints) > 0 {
				webData["provider_hints"] = hints
			}
		}
		if plugins := capture.wordpressPlugins; len(plugins) > 0 {
			webData["wordpress_plugins"] = plugins
		}
		if assets := capture.jsAssets; len(assets) > 0 {
			webData["js_assets"] = assets
		}
		if themes := capture.wordpressThemes; len(themes) > 0 {
			themeURL := capture.location
			if themeURL == "" {
				themeURL = headerURL
			}
			if themeURL == "" {
				themeURL = "https://" + host
			}
			enrichWordPressThemeVersions(ctx, themeURL, themes)
			webData["wordpress_themes"] = themes
		}

		if scrapeErr != nil {
			webData["error"] = scrapeErr.Error()
			return "web", webData, fmt.Errorf("web gather failed for %s: %w", host, scrapeErr)
		}

		return "web", webData, nil
	}
}

// scrapePage navigates to url, captures the final document response headers,
// and extracts page title, visible body text, final location, and HTML hints.
func scrapePage(parent context.Context, url string, capture *pageCapture) error {
	ctx, cancel := chromedp.NewContext(parent)
	defer cancel()

	var mu sync.Mutex
	var lastDoc documentResponse
	docRequestURLs := make(map[network.RequestID]string)

	chromedp.ListenTarget(ctx, func(ev any) {
		switch e := ev.(type) {
		case *network.EventResponseReceived:
			if e.Type != network.ResourceTypeDocument {
				return
			}
			mu.Lock()
			docRequestURLs[e.RequestID] = e.Response.URL
			if hdrs := headersFromNetwork(e.Response.Headers); len(hdrs) > 0 {
				lastDoc = documentResponse{url: e.Response.URL, headers: hdrs}
			} else {
				lastDoc = documentResponse{url: e.Response.URL}
			}
			mu.Unlock()
		case *network.EventResponseReceivedExtraInfo:
			mu.Lock()
			if docURL, ok := docRequestURLs[e.RequestID]; ok {
				lastDoc = documentResponse{
					url:     docURL,
					headers: headersFromNetwork(e.Headers),
				}
				for key, value := range e.Headers {
					if strings.EqualFold(key, "set-cookie") {
						capture.cookieNames = mergeCookieNames(capture.cookieNames, parseSetCookieNames(value))
					}
				}
			}
			mu.Unlock()
		}
	})

	var wpIntel wordpressIntel
	var jsIntel jsAssetIntel
	var metaIntel map[string]any
	var documentCookies []string
	err := chromedp.Run(ctx,
		network.Enable(),
		chromedp.Navigate(url),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.Title(&capture.title),
		chromedp.Text("body", &capture.text, chromedp.ByQuery),
		chromedp.OuterHTML("html", &capture.html, chromedp.ByQuery),
		chromedp.Location(&capture.location),
		chromedp.Evaluate(htmlTechHintsJS, &capture.htmlHints),
		chromedp.Evaluate(metaIntelJS, &metaIntel),
		chromedp.Evaluate(jsAssetIntelJS, &jsIntel),
		chromedp.Evaluate(wordpressIntelJS, &wpIntel),
		chromedp.Evaluate(cookieIntelJS, &documentCookies),
	)
	if err != nil {
		return err
	}

	capture.jsAssets, capture.jsHints = normalizeJSAssets(jsIntel)
	capture.wordpressPlugins = normalizeWordPressItems(wpIntel.Plugins)
	capture.wordpressThemes = normalizeWordPressItems(wpIntel.Themes)
	if len(metaIntel) > 0 {
		capture.meta = metaIntel
	}
	capture.cookieNames = mergeCookieNames(capture.cookieNames, documentCookies)
	if len(capture.html) > 1024*1024 {
		capture.html = capture.html[:1024*1024]
	}

	mu.Lock()
	capture.headers = lastDoc.headers
	capture.headerURL = lastDoc.url
	mu.Unlock()

	return nil
}

func fetchDocumentHeaders(ctx context.Context, rawURL string) (map[string]string, string) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, ""
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, ""
	}
	defer resp.Body.Close()

	headers := make(map[string]string, len(resp.Header))
	for key, values := range resp.Header {
		headers[key] = strings.Join(values, ", ")
	}

	return headers, resp.Request.URL.String()
}

var (
	copyrightSymbol = regexp.MustCompile(`©`)
	copyrightWord   = regexp.MustCompile(`(?i)\bcopyright\b`)
	yearPattern     = regexp.MustCompile(`\b(19|20)\d{2}\b`)
)

// extractCopyrights returns deduplicated visible text snippets that contain a
// copyright symbol, the word "Copyright" (case-insensitive), or a 4-digit year
// in the 19xx/20xx range.
func extractCopyrights(text string) []string {
	seen := make(map[string]struct{})
	var out []string

	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !copyrightSymbol.MatchString(line) &&
			!copyrightWord.MatchString(line) &&
			!yearPattern.MatchString(line) {
			continue
		}
		if _, ok := seen[line]; ok {
			continue
		}
		seen[line] = struct{}{}
		out = append(out, line)
	}

	return out
}
