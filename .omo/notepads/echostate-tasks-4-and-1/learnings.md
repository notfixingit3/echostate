# EchoState Learnings

## WHOIS Gatherer

- `github.com/likexian/whois` provides `whois.Whois(domain)` but no context-aware variant in v1.15.7, so the gatherer wraps the call in a goroutine and selects on `context.WithTimeout(ctx, 10*time.Second).Done()` to honor the required deadline.
- `github.com/likexian/whois-parser` exposes `whoisparser.Parse(raw)` and returns `WhoisInfo` with `Domain` and `Registrar` pointers; fields may be nil for malformed or minimal responses.
- The raw WHOIS text is always parsed in full, then truncated to 8,192 bytes before being stored under the `raw` key so the scanner never stores more than the limit.
- If parsing fails but raw text exists, the gatherer returns both the map (containing the truncated raw text) and an error so the scanner records the failure in `result.Errors`.
- A package-level `whoisLookup` shim mirrors `whois.Whois` and lets mocked tests inject raw responses without network access.
# EchoState Tasks 4 and 1 Learnings

## Web / Copyright Gatherer

- `internal/scanner/web.go` implements `newWebGatherer(browserWSURL)` returning the
  `Gatherer` closure required by `internal/scanner/scanner.go`.
- The gatherer normalizes the target host with `scanner.NormalizeHost`, then tries
  `https://<host>` and falls back to `http://<host>` if the secure request fails.
- Browser connection uses `chromedp.NewRemoteAllocator`; the default endpoint is
  `ws://localhost:3000/` and can be overridden via `config.Config.BrowserWSURL`.
- Each gatherer run is bounded by a 10-second `context.WithTimeout`.
- Page data is extracted with `chromedp.Navigate`, `chromedp.WaitVisible("body")`,
  `chromedp.Title`, `chromedp.Text("body", ...)` and `chromedp.Location`.
- `extractCopyrights(text)` is a pure helper that returns deduplicated visible text
  snippets containing `©`, `Copyright` (case-insensitive), or a 4-digit year
  (`19xx`/`20xx`).
- `TestCopyrightMatcher` unit-tests the helper directly; `TestWebLive` integrates
  with a real browser and skips when the browser endpoint is unavailable.
- `scanner.NewScanner` now accepts the browser WebSocket URL so the handler can
  pass `cfg.BrowserWSURL`; the `Gatherer` function type itself was not changed.
