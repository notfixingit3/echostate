package scanner

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

const defaultBrowserWSURL = "ws://localhost:3000/"

// gatherWeb is the default web/copyright gatherer.
// It uses the default browser WebSocket endpoint and is suitable for direct use
// in tests or simple callers. NewScanner builds its own gatherer closure when a
// custom endpoint is configured.
var gatherWeb Gatherer = newWebGatherer("")

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
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		allocCtx, allocCancel := chromedp.NewRemoteAllocator(ctx, browserWSURL)
		defer allocCancel()

		var title, text, url string
		if err := scrapePage(allocCtx, "https://"+host, &title, &text, &url); err != nil {
			title, text, url = "", "", ""
			if err2 := scrapePage(allocCtx, "http://"+host, &title, &text, &url); err2 != nil {
				return "web", map[string]any{"error": err2.Error()}, fmt.Errorf("web gather failed for %s: %w", host, err2)
			}
		}

		return "web", map[string]any{
			"title":      title,
			"url":        url,
			"copyrights": extractCopyrights(text),
		}, nil
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
