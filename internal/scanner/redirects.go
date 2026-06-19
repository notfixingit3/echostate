package scanner

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const maxRedirectHops = 10

func captureRedirectChain(ctx context.Context, host string) ([]map[string]any, string) {
	host = NormalizeHost(host)
	if host == "" {
		return nil, ""
	}

	httpsChain, httpsFinal := followRedirectChain(ctx, "https://"+host)
	httpChain, httpFinal := followRedirectChain(ctx, "http://"+host)

	if len(httpsChain) >= len(httpChain) {
		return httpsChain, httpsFinal
	}
	return httpChain, httpFinal
}

func followRedirectChain(ctx context.Context, startURL string) ([]map[string]any, string) {
	client := &http.Client{
		Timeout: 6 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	current := startURL
	var hops []map[string]any

	for hop := 0; hop < maxRedirectHops; hop++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodHead, current, nil)
		if err != nil {
			break
		}
		req.Header.Set("User-Agent", "EchoState/1.0 (+https://github.com/notfixingit3/echostate)")

		resp, err := client.Do(req)
		if err != nil {
			getReq, getErr := http.NewRequestWithContext(ctx, http.MethodGet, current, nil)
			if getErr != nil {
				break
			}
			getReq.Header.Set("User-Agent", "EchoState/1.0 (+https://github.com/notfixingit3/echostate)")
			resp, err = client.Do(getReq)
			if err != nil {
				break
			}
		}

		hops = append(hops, map[string]any{
			"url":    current,
			"status": resp.StatusCode,
		})

		if resp.StatusCode < 300 || resp.StatusCode >= 400 {
			resp.Body.Close()
			return hops, current
		}

		location := strings.TrimSpace(resp.Header.Get("Location"))
		resp.Body.Close()
		if location == "" {
			return hops, current
		}

		next, err := resolveRedirectURL(current, location)
		if err != nil {
			return hops, current
		}
		current = next
	}

	if len(hops) == 0 {
		return nil, ""
	}
	return hops, current
}

func resolveRedirectURL(baseURL, location string) (string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	target, err := url.Parse(location)
	if err != nil {
		return "", err
	}
	return base.ResolveReference(target).String(), nil
}

func redirectHopCount(chain []map[string]any) int {
	if len(chain) <= 1 {
		return 0
	}
	return len(chain) - 1
}

func redirectSummary(chain []map[string]any) string {
	if len(chain) == 0 {
		return ""
	}
	if len(chain) == 1 {
		return fmt.Sprintf("%v", chain[0]["url"])
	}
	first := fmt.Sprint(chain[0]["url"])
	last := fmt.Sprint(chain[len(chain)-1]["url"])
	return fmt.Sprintf("%s → %s (%d hops)", first, last, len(chain)-1)
}