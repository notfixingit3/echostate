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
	location         string
	headerURL        string
	headers          map[string]string
	htmlHints        []string
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
		)

		webData := map[string]any{
			"title":            capture.title,
			"url":              capture.location,
			"header_url":       headerURL,
			"copyrights":       extractCopyrights(capture.text),
			"headers":          headers,
			"security_headers": securityHeaders,
			"tech_stack":       techStack,
		}
		if plugins := capture.wordpressPlugins; len(plugins) > 0 {
			webData["wordpress_plugins"] = plugins
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
			}
			mu.Unlock()
		}
	})

	var wpIntel wordpressIntel
	err := chromedp.Run(ctx,
		network.Enable(),
		chromedp.Navigate(url),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.Title(&capture.title),
		chromedp.Text("body", &capture.text, chromedp.ByQuery),
		chromedp.Location(&capture.location),
		chromedp.Evaluate(htmlTechHintsJS, &capture.htmlHints),
		chromedp.Evaluate(wordpressIntelJS, &wpIntel),
	)
	if err != nil {
		return err
	}

	capture.wordpressPlugins = normalizeWordPressItems(wpIntel.Plugins)
	capture.wordpressThemes = normalizeWordPressItems(wpIntel.Themes)

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