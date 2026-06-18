package scanner

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

const defaultBrowserWSURL = "ws://localhost:3000/"

// newWebGatherer returns a Gatherer that connects to a browserless Chrome
// instance at browserWSURL and extracts the landing page title, final URL, and
// copyright snippets from visible text.
func newWebGatherer(browserWSURL string) Gatherer {
	if browserWSURL == "" {
		browserWSURL = defaultBrowserWSURL
	}

	return func(ctx context.Context, host string) (string, map[string]any, error) {
		host = NormalizeHost(host)
		if host == "" {
			return "web", map[string]any{"error": "empty host"}, errors.New("empty host")
		}

		// Per-gatherer timeout; the caller may have a wider handler context.
		ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()

		var title, text, url string
		var headers map[string]string
		var securityHeaders map[string]string
		var techStack []string

		// Fetch headers quickly via standard net/http
		client := &http.Client{Timeout: 5 * time.Second}
		if resp, err := client.Get("https://" + host); err == nil {
			headers = make(map[string]string)
			securityHeaders = make(map[string]string)
			for k, v := range resp.Header {
				val := strings.Join(v, ", ")
				headers[k] = val

				kl := strings.ToLower(k)
				if kl == "strict-transport-security" || kl == "content-security-policy" || kl == "x-frame-options" || kl == "x-content-type-options" {
					securityHeaders[kl] = val
				}

				if kl == "server" || kl == "x-powered-by" {
					techStack = append(techStack, fmt.Sprintf("%s: %s", k, val))
				}
			}
			resp.Body.Close()
		} else if resp, err := client.Get("http://" + host); err == nil {
			headers = make(map[string]string)
			securityHeaders = make(map[string]string)
			for k, v := range resp.Header {
				val := strings.Join(v, ", ")
				headers[k] = val

				kl := strings.ToLower(k)
				if kl == "strict-transport-security" || kl == "content-security-policy" || kl == "x-frame-options" || kl == "x-content-type-options" {
					securityHeaders[kl] = val
				}

				if kl == "server" || kl == "x-powered-by" {
					techStack = append(techStack, fmt.Sprintf("%s: %s", k, val))
				}
			}
			resp.Body.Close()
		}

		resolvedWS := resolveBrowserWSURL(ctx, browserWSURL)
		allocCtx, allocCancel := chromedp.NewRemoteAllocator(ctx, resolvedWS)
		defer allocCancel()

		scrapeErr := scrapePage(allocCtx, "https://"+host, &title, &text, &url)
		if scrapeErr != nil {
			title, text, url = "", "", ""
			scrapeErr = scrapePage(allocCtx, "http://"+host, &title, &text, &url)
		}

		webData := map[string]any{
			"title":            title,
			"url":              url,
			"copyrights":       extractCopyrights(text),
			"headers":          headers,
			"security_headers": securityHeaders,
			"tech_stack":       techStack,
		}

		if scrapeErr != nil {
			webData["error"] = scrapeErr.Error()
			return "web", webData, fmt.Errorf("web gather failed for %s: %w", host, scrapeErr)
		}

		return "web", webData, nil
	}
}

// scrapePage navigates to url, waits for the body, and captures the page title,
// visible body text, and final location.
func scrapePage(parent context.Context, url string, title, text, location *string) error {
	ctx, cancel := chromedp.NewContext(parent)
	defer cancel()

	return chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.Title(title),
		chromedp.Text("body", text, chromedp.ByQuery),
		chromedp.Location(location),
	)
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
