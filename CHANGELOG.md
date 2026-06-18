# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.0.1-beta.28] - 2026-06-18

### Added

- **Richer PDF reports** — Cover branding with EchoState logo, web screenshot thumbnail, snapshot change summaries, intel highlights, and expanded DNS/TLS sections.
- **Report state hydration** — Snapshot detail restores the latest report for that snapshot on page load (pending, running, or completed).
- **Scan result downloads** — Home scan result card now offers **Download PDF** when a report completes (matching snapshot detail).

### Changed

- **Report E2E** — Playwright flow stays on the snapshot page, waits for completion, and downloads via `snapshot-download-report-button`. E2E global setup builds local dev images and recreates volumes when auth is bootstrapped.

## [0.0.1-beta.27] - 2026-06-18

### Fixed

- **Snapshot create report** — Snapshot detail now polls report status, shows in-card progress, and offers **View report** / **Download PDF** when complete (previously only a brief “queued” alert).
- **Report downloads** — Shared `downloadReport()` helper uses credentialed fetch so PDF downloads work behind auth on snapshot, report detail, and reports table pages.

### Changed

- **Reports list** — `/reports?snapshot_id=` pre-fills the snapshot filter from the snapshot page **View reports** link.

## [0.0.1-beta.26] - 2026-06-18

### Changed

- **API Docker image** — Dropped bundled Chromium (~630 MB); web scans continue to use the `browserless/chrome` compose service via `BROWSER_WS_URL`.
- **Node.js 22** — CI and frontend Docker builds use Node 22 LTS (replaces Node 20).

## [0.0.1-beta.25] - 2026-06-18

### Fixed

- **Graph node drag** — Dragged nodes stay pinned instead of snapping back when released.
- **Docker semver on dev** — Dev CI now publishes `ghcr.io/.../echostate(-frontend):<VERSION>` from the root `VERSION` file so semver pins work without waiting on a separate tag workflow.

### Changed

- **docker-compose** — Removed obsolete top-level `version` key (silences compose v2 warning).

## [0.0.1-beta.24] - 2026-06-18

### Added

- **Export all / import bundle** — Admin **Data transfer** page with `GET /api/export` and `POST /api/import` to move targets, snapshots, collections, notes, saved graph views, and optional blobs/settings between installs.
- **Graph canvas sizing** — Graph view now fills its container and re-measures on tab and window resize.

### Fixed

- **Collections selects** — Collection and target pickers show names instead of UUIDs after selection.

## [0.0.1-beta.23] - 2026-06-18

### Added

- **Profile display name** — Edit and save your display name via `PATCH /api/auth/profile`.
- **Device enrollment codes** — **Profile → Issue device code** (`POST /api/auth/device-code`) for signing in on another browser or phone.
- **Per-user theme** — Light/dark/system preference stored on the account; synced on sign-in and from the nav theme toggle (`users.theme`).

### Changed

- **Navigation split** — `/profile` for personal settings; `/admin` for server configuration (admin-only nav link). Legacy `/settings/*` redirects.
- **Passkey rename** — `PUT` alias added alongside `PATCH` for credential nickname updates.

## [0.0.1-beta.22] - 2026-06-18

### Added

- **Profile page** — Nav and Admin **Profile** with timezone preference, passkey register/rename/remove (replaces Account & passkeys).
- **Passkey rename** — `PATCH /api/users/:id/credentials/:credId` and inline rename in Profile.
- **User timezone** — `users.timezone` column and `PATCH /api/auth/profile`.
- **`scripts/docker-clean.sh`** — Prune local EchoState dev images and Docker build cache.

### Changed

- **Login** — Passkey-first sign-in; enrollment/recovery code is secondary. Sign-out no longer clears saved account id (fixes missing passkey button after logout).
- **Login shell** — Hide nav/footer until authenticated; centered login card with theme toggle.
- **CI** — Workflow concurrency, job timeouts, and `go test -timeout 12m` to avoid hung/stuck Actions runs.

## [0.0.1-beta.21] - 2026-06-17

### Added

- **Admin console** — Settings renamed to Admin with sidebar navigation: Account & passkeys, Integrations, System, Authentication, and User management.
- **Account & passkeys** — `/settings/account` for registering and removing passkeys on the signed-in account (nav **Account** link).
- **Login device enrollment** — After code verification, choose **Sign in with passkey** or **Register passkey on this device** instead of auto-attempting passkey login.

### Changed

- **User management** — Users and enrollment codes moved from monolithic Settings to **Admin → User management**.
- **Nav** — Top-level **Settings** link renamed **Admin**; **Account** shortcut added to the user menu (desktop and mobile).

## [0.0.1-beta.20] - 2026-06-17

### Added

- **Passkey authentication** — WebAuthn passkeys with HTTP-only session cookies (`echostate_session`) and enrollment cookies (`echostate_enroll`).
- **Enrollment codes** — Configurable-length numeric codes and alphanumeric recovery codes; single-use, HMAC-hashed, with attempt rate limiting.
- **Roles** — `admin` (full access) and `scanner` (scan, reports, read intel); session middleware on protected API routes.
- **Auth settings** — Admin-configurable defaults in Settings: auth toggle, code TTL/length, session TTL, attempt limits, WebAuthn RP ID/origin.
- **User management** — Settings UI to create users and issue enrollment/recovery codes; credential list/delete API.
- **CLI** — `echostate auth bootstrap-admin` and `echostate auth issue-admin-code` for first admin and break-glass recovery.
- **Login page** — `/login` with enrollment code verification, passkey registration, and returning passkey sign-in.

### Changed

- Protected `/api/*` routes use session auth when users exist; legacy API key still works for admin mutations before bootstrap.
- CORS allows `X-API-Key`; frontend `fetchApi` sends cookies (`credentials: include`).

## [0.0.1-beta.19] - 2026-06-17

### Added

- **Saved graph views** — `graph_views` table and CRUD API (`/api/graph/views`) to bookmark graph lenses.
- **Graph view restore** — Save/load view tab, target filter, vantage, snapshot compare pair, pinned node layout, and selected node from the `/graph` toolbar.

## [0.0.1-beta.18] - 2026-06-17

### Added

- **Investigation notes** — `investigation_notes` table and CRUD API with archive, trash, restore, and permanent delete from trash.
- **Note references** — Structured URL, target, snapshot, collection, and graph-node reference chips alongside sanitized HTML links in note bodies.
- **Trash retention** — Trashed notes auto-purge after 30 days; manual **Delete forever** from the trash tab.
- **Notes UI** — `/notes` page plus embedded panels on target detail, graph node inspector, and collections.

## [0.0.1-beta.17] - 2026-06-17

### Added

- **Target collections** — `target_collections` tables, CRUD API (`/api/collections`), member management, and bulk rescan enqueue; `/collections` UI with watchlist manager.
- **Graph export** — PNG (canvas snapshot) and SVG (node positions) download buttons on the graph toolbar.
- **Settings tooltips** — `HelpTip` / `LabelWithHelp` across Settings sections; expanded `help-copy.ts` for collections and export.

### Changed

- **JA3S CI fix** — Removed `dreadl0ck/ja3` / `gopacket` dependency; inline TLS ServerHello parse and JA3S hashing (`tls_server_hello.go`, `ja3s_hash.go`).
- Graph shows a rescan hint when a target has fewer than two snapshots (temporal history needs multiple scans).

## [0.0.1-beta.16] - 2026-06-17

### Changed

- Traceroute external vantage label renamed to **External vantage** (no HackerTarget wording in the UI).

### Added

- **JA3S TLS fingerprint** — ServerHello hash on port 443; synced to Neo4j `JA3S` nodes with `HAS_JA3S` edges on the infrastructure graph.
- **JS asset analysis** — Passive script/stylesheet URL collection with library, CDN, and analytics hints merged into `tech_stack`.
- **Email posture score** — Composite SPF/DKIM/DMARC grade (`MAIL_POSTURE`) with A–F rating and findings in the DNS intel tab.
- **BGP path profiling** — `path_profile` stability label (stable / diverse / volatile) from RIPEstat AS-path diversity.
- **Tech auto-tags** — Targets auto-tag `shopify`, `nextjs`, `drupal`, `react`, `vue`, and `angular` from detected stack signals.

### Changed

- Snapshot diffs alert on JA3S changes, mail posture grade shifts, and BGP path stability transitions.

## [0.0.1-beta.15] - 2026-06-17

### Added

- **Temporal graph topology** — versioned Neo4j edges with `snapshot_id`; history slider reloads the canvas for all 7 views when a target is selected.
- **Graph compare modes** — `vs previous` and `vs latest` toggles with green/red topology diff overlay on nodes and edges.
- **Graph canvas controls** — zoom, fit, reset layout, drag-to-pin nodes, neighbor highlighting, and inspector cross-view jumps.
- **Contextual tooltips** — dense `HelpTip` icons across graph, scan form, intel summary, and page headers (`help-copy.ts` registry).

### Changed

- `GET /api/graph` accepts `snapshot_id`, `compare_snapshot_id`, and `compare_mode`; returns `topology_diff`.
- Route diff API supports arbitrary snapshot pairs for BGP/traceroute narrative panels.

## [0.0.1-beta.14] - 2026-06-17

### Added

- **SOA DNS records** — zone authority lookup (walks parent labels on subdomains), parsed fields in the DNS intel tab, snapshot diffs for serial/mname/rname/TTL changes, and `SOAZone` nodes on the DNS graph view.

## [0.0.1-beta.13] - 2026-06-17

### Added

- **Graph intel events (G15)** — snapshot diffs and enrichment hits sync to Neo4j; `/graph` shows an Intel events panel per target.
- **DMARC graph nodes** — `DMARCPolicy` nodes and `HAS_DMARC` edges on the DNS view.
- **WordPress plugin & theme detection** — passive extraction from page assets, intel summary badges, and snapshot diffs.
- **Version update indicator** — branch-aware status light in the footer (`GET /api/version`).
- **Graph UX** — structured node inspector, neighbor navigation, snapshot history bar on BGP/traceroute views.

### Changed

- Targets auto-tag `wordpress` when WP signals are detected on scan.
- Theme versions fetched from `style.css` when asset URLs lack `?ver=`.
- Dev CORS allows any localhost origin against the Docker API.

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

[0.0.1-beta.28]: https://github.com/notfixingit3/echostate/releases/tag/v0.0.1-beta.28
[0.0.1-beta.27]: https://github.com/notfixingit3/echostate/releases/tag/v0.0.1-beta.27
[0.0.1-beta.26]: https://github.com/notfixingit3/echostate/releases/tag/v0.0.1-beta.26
[0.0.1-beta.25]: https://github.com/notfixingit3/echostate/releases/tag/v0.0.1-beta.25
[0.0.1-beta.24]: https://github.com/notfixingit3/echostate/releases/tag/v0.0.1-beta.24
[0.0.1-beta.23]: https://github.com/notfixingit3/echostate/releases/tag/v0.0.1-beta.23
[0.0.1-beta.22]: https://github.com/notfixingit3/echostate/releases/tag/v0.0.1-beta.22
[0.0.1-beta.21]: https://github.com/notfixingit3/echostate/releases/tag/v0.0.1-beta.21
[0.0.1-beta.20]: https://github.com/notfixingit3/echostate/releases/tag/v0.0.1-beta.20
[0.0.1-beta.19]: https://github.com/notfixingit3/echostate/releases/tag/v0.0.1-beta.19
[0.0.1-beta.18]: https://github.com/notfixingit3/echostate/releases/tag/v0.0.1-beta.18
[0.0.1-beta.17]: https://github.com/notfixingit3/echostate/releases/tag/v0.0.1-beta.17
[0.0.1-beta.16]: https://github.com/notfixingit3/echostate/releases/tag/v0.0.1-beta.16
[0.0.1-beta.15]: https://github.com/notfixingit3/echostate/releases/tag/v0.0.1-beta.15
[0.0.1-beta.14]: https://github.com/notfixingit3/echostate/releases/tag/v0.0.1-beta.14
[0.0.1-beta.13]: https://github.com/notfixingit3/echostate/releases/tag/v0.0.1-beta.13
[0.0.1-beta.12]: https://github.com/notfixingit3/echostate/releases/tag/v0.0.1-beta.12
[0.0.1-beta.11]: https://github.com/notfixingit3/echostate/releases/tag/v0.0.1-beta.11
[0.0.1-beta.10]: https://github.com/notfixingit3/echostate/releases/tag/v0.0.1-beta.10
[0.0.1-beta.5]: https://github.com/notfixingit3/echostate/releases/tag/v0.0.1-beta.5
[0.0.1-beta.4]: https://github.com/notfixingit3/echostate/releases/tag/v0.0.1-beta.4
[0.0.1-beta.3]: https://github.com/notfixingit3/echostate/releases/tag/v0.0.1-beta.3
[0.0.1-beta.2]: https://github.com/notfixingit3/echostate/releases/tag/v0.0.1-beta.2
[0.0.1-beta.1]: https://github.com/notfixingit3/echostate/releases/tag/v0.0.1-beta.1
[0.0.1-beta.0]: https://github.com/notfixingit3/echostate/releases/tag/v0.0.1-beta.0