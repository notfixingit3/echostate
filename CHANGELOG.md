# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.0.1-beta.12] - 2026-06-17

### Added

- **HIBP enrichment** — breach checks for WHOIS/crawl emails (up to 3 per snapshot) when `hibp_api_key` is set.
- **RiskIQ / PassiveTotal enrichment** — passive DNS correlation for the scanned host.
- **Graph drift alerting (G13)** — BGP visible-origin, AS-path, RPKI, and traceroute hop changes flow into `change_details` and webhooks.
- **Alert rule builder** — per-rule webhook targeting, graph-drift preset, and unified system-config save.
- **Enrichment intel tab** — Shodan, Censys, HIBP, and RiskIQ results in the snapshot UI.

### Changed

- Settings page wraps all system configuration in one save action (API keys, scheduler, retention, alert rules).

## [0.0.1-beta.11] - 2026-06-17

### Added

- **Async scan jobs** — `POST /api/scan` returns `202` with a job ID; poll `GET /api/scans/:id` until complete.
- **Field-level snapshot diffs** — structured `change_details` (type, severity, summary) stored per snapshot and shown in the Changes intel tab.
- **Richer webhooks** — structured change payloads with configurable alert rules (severity/type filters).
- **Scheduled rescans** — background scheduler re-enqueues stale targets (configurable interval and max age).
- **Snapshot retention** — optional pruning of old snapshots per target after each scan.
- **Blob offload** — screenshot thumbnails stored in `snapshot_blobs` instead of inline `raw_data`.
- **API key auth** — optional `ECHOSTATE_API_KEY` (or Settings UI key) protects write endpoints; read/browse stays open.
- **Enrichment worker** — optional Shodan/Censys lookups keyed off Settings API keys.
- **Deeper passive intel** — SPF/DKIM/DMARC parsing, TLS chain + expiry, CertSpotter CT fallback.
- **Settings expansion** — scan concurrency, per-gatherer timeouts, scheduler, retention, alert rules, and enrichment keys in the UI.

### Changed

- Scan handler no longer blocks on long CT/TLS gathers; work runs in a bounded-concurrency worker pool.
- Dedup hash excludes volatile fields (screenshot thumbnails, scan timestamps).

## [0.0.1-beta.10] - 2026-06-17

### Added

- Certificate transparency gatherer via crt.sh with CT intel tab and click-to-scan subdomains.
- Interactive infrastructure graph at `/graph` with seven views: `infra`, `ct`, `dns`, `cert`, `bgp`, `traceroute`, `peering` (`GET /api/graph?view=`).
- Neo4j graph sync: CT subdomains, DNS dependencies, cert SAN overlap, BGP AS-path enrichment, shared-hop convergence, RPKI/hijack risk edges, infra community clusters, and historical route diff on BGP/traceroute views.
- Multi-vantage traceroute (local + HackerTarget) with vantage filter and divergence panel.
- PeeringDB ASN enrichment with IX nodes and peering graph view.
- Visual screenshot gatherer (chromedp JPEG thumbnails), screenshot timeline intel tab, and `GET /api/targets/:id/screenshots`.
- Hop geo map on traceroute view (RIPEstat geolocation).
- Rescan target button on target detail (re-POST `/api/scan`).
- Query-param static routes `/target?id=` and `/target/snapshots?id=` for Docker static export.
- Per-gatherer scan timeouts (CT 80s, traceroute 40s, screenshot 25s) and 100s scan handler budget.

### Changed

- Scan & Report button on the home page for one-click scan plus PDF queue.
- Brand wordmark corrected to **EchoState** across SVG, UI, and PNG assets.

### Fixed

- Graph page target filter no longer shows raw UUIDs; Neo4j subgraph query rewritten.
- Graph API returns `"edges": []` instead of `null` for empty graphs (fixes CT/DNS tab crashes).
- crt.sh CT requests use 60s HTTP timeout with fast-failure retries (502/503) for slow domains.
- Local Docker version badge reads `ECHOSTATE_VERSION` / root `VERSION` file (beta.10).

## [0.0.1-beta.5] - 2026-06-17

### Added

- **Scan & Report** button on the home page — one-click scan plus PDF report generation.

### Changed

- Brand wordmark corrected to **EchoState** (single word) across SVG logo, UI component, and PNG assets.

### Fixed

- Graph page target filter no longer shows raw UUIDs or fails to load; Neo4j subgraph query rewritten.

## [0.0.1-beta.4] - 2026-06-17

### Added

- Certificate transparency subdomain enumeration via crt.sh (`ct` gatherer, CT intel tab).
- Clickable CT subdomain badges prefill the scan form (`/?scan=hostname`).
- Interactive infrastructure graph UI at `/graph` with drag-and-drop force layout.
- `GET /api/graph` exposes Neo4j nodes and edges (`?target_id=` for subgraph).
- Neo4j sync links targets to shared JARM hashes, favicon MMH3, and certificate issuers.

## [0.0.1-beta.3] - 2026-06-17

### Added

- Central `VERSION` file and build-time injection for API (`/health`) and frontend footer.
- Redesigned footer with brand mark, capability badges, changelog link, and version badge.
- Pushover notification integration with configurable priority, sound, device, title, and emergency retry/expire settings (documented in Settings UI).
- Favicon MMH3/SHA256 fingerprinting (Shodan-compatible hash).
- `robots.txt` and `sitemap.xml` crawl extraction.
- JARM TLS server fingerprinting on port 443.
- Cloud storage bucket reference detection (S3, Azure, GCP, DO Spaces).
- BGP hijack risk heuristics via RIPEstat routing visibility and RPKI validation.

### Changed

- Light mode is now the default theme (system preference still available via toggle).
- Web gatherer captures HTTP response headers from the browser's final document load (`header_url`), not a separate root `GET`.
- Expanded `tech_stack` heuristics: infra headers (`Server`, `X-Powered-By`, `Via`, ASP.NET, Varnish, Cloudflare, etc.) plus HTML hints (`meta generator`, WordPress, Drupal, Shopify, Next.js).

## [0.0.1-beta.2] - 2026-06-17

### Changed

- New EchoState brand identity — radar echo mark with theme-aware SVG logo (light/dark), refreshed favicon, and updated README assets.

### Added

- DNS gatherer — A/AAAA, MX, NS, TXT, CNAME, and DMARC records via configurable resolvers.
- TLS gatherer — certificate issuer, SANs, validity window, and signature algorithm.
- Web gatherer enhancements — HTTP security headers (HSTS, CSP, X-Frame-Options, X-Content-Type-Options) and `Server`/`X-Powered-By` tech-stack hints.
- Intel UI — DNS and TLS tabs, cert expiry and A-record summary cards; security headers and tech stack in the Web tab.
- `TODO.md` roadmap — items 1–4 (TLS, DNS, tech fingerprinting, security headers) marked complete.

### Fixed

- `storeSnapshot` no longer panics when handler `config` is nil (CI test failure).
- Release workflow cleans up stale GitHub releases before retagging.

### Removed

- `.omo/` agent artifacts removed from the repository.

## [0.0.1-beta.1] - 2026-06-17

### Added

- Target tags, snapshot diff viewer, dynamic settings UI, and webhook notifications.

## [0.0.1-beta.0] - 2026-06-17

### Added

- Initial EchoState release — Go API, Next.js UI, PostgreSQL snapshots, PDF reports, WHOIS/ASN/web gatherers, Docker Compose stack.

[0.0.1-beta.12]: https://github.com/notfixingit3/echostate/releases/tag/v0.0.1-beta.12
[0.0.1-beta.11]: https://github.com/notfixingit3/echostate/releases/tag/v0.0.1-beta.11
[0.0.1-beta.10]: https://github.com/notfixingit3/echostate/releases/tag/v0.0.1-beta.10
[0.0.1-beta.5]: https://github.com/notfixingit3/echostate/releases/tag/v0.0.1-beta.5
[0.0.1-beta.4]: https://github.com/notfixingit3/echostate/releases/tag/v0.0.1-beta.4
[0.0.1-beta.3]: https://github.com/notfixingit3/echostate/releases/tag/v0.0.1-beta.3
[0.0.1-beta.2]: https://github.com/notfixingit3/echostate/releases/tag/v0.0.1-beta.2
[0.0.1-beta.1]: https://github.com/notfixingit3/echostate/releases/tag/v0.0.1-beta.1
[0.0.1-beta.0]: https://github.com/notfixingit3/echostate/releases/tag/v0.0.1-beta.0