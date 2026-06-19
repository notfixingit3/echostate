# EchoState Reconnaissance Roadmap

Items are grouped by **difficulty** (engineering effort). Within each tier, lower numbers were earlier roadmap priorities.

---

## Easy ✓

| # | Item | Status | Notes |
|---|------|--------|-------|
| 17 | App versioning, footer & theme defaults | ✅ Done | `VERSION` file, footer badge, light mode default |
| 4 | HTTP security headers | ✅ Done | HSTS, CSP, X-Frame-Options, X-Content-Type-Options from final document load |
| 9 | `robots.txt` & `sitemap.xml` extraction | ✅ Done | Crawl tab + path summary |
| 11 | Cloud storage bucket detection | ✅ Done | S3, Azure, GCP, DO Spaces hints in HTML |
| 18 | Cert expiry / new-CT intel badges | ✅ Done | Summary cards on target/snapshot detail |
| 19 | Scan job status in UI | ✅ Done | Poll `GET /api/scans/:id`; queued / running / failed |

---

## Easy → Medium

| # | Item | Status | Notes |
|---|------|--------|-------|
| 16 | Pushover notifications | ✅ Done | Webhook type + documented Settings fields |
| 5 | Subdomain enum via crt.sh | ✅ Done | crt.sh + CertSpotter fallback, CT tab, click-to-scan |
| 7 | Visual screenshots | ✅ Done | chromedp JPEG thumbnails + timeline UI |
| 20 | Field-level snapshot diffs | ✅ Done | `change_details` (type, severity, summary); Changes intel tab |
| 21 | Richer webhook payloads | ✅ Done | Structured change entries + alert-rule filtering |
| 22 | SPF / DKIM / DMARC intel | ✅ Done | Parsed mail-security fields in DNS tab |

---

## Medium

| # | Item | Status | Notes |
|---|------|--------|-------|
| 1 | TLS/SSL certificate analysis | ✅ Done | Leaf cert, chain, `days_remaining`, expiry alerts in diff |
| 2 | DNS security & deep dive | ✅ Done | A/AAAA, MX, NS, TXT, CNAME, SOA, DMARC, SPF/DKIM parse |
| 3 | Tech stack fingerprinting | ✅ Done | Headers, HTML hints, JS asset analysis, framework auto-tags |
| 8 | Favicon MMH3/SHA256 hashing | ✅ Done | Shodan-compatible; enrichment worker can correlate |
| 23 | Async scan job queue | ✅ Done | `POST /api/scan` → 202, worker pool, `GET /api/scans/:id` |
| 24 | Scheduled rescans | ✅ Done | Background scheduler for stale targets; configurable in Settings |
| 25 | Snapshot retention / pruning | ✅ Done | Per-target keep-N after scan; blob offload for thumbnails |
| 26 | Minimal API-key auth | ✅ Done | `ECHOSTATE_API_KEY` or Settings key on write routes |
| 27 | API key vault | ✅ Done | Shodan, Censys, HIBP, RiskIQ keys in Settings + async enrichment worker |
| 6 | Port/banner grabbing | ❌ Skipped | Out of scope (passive-only product direction) |

---

## Medium → Hard

| # | Item | Status | Notes |
|---|------|--------|-------|
| 10 | JARM / JA3 fingerprinting | ✅ Done | JARM + JA3S on 443; JA3S linked on infra graph |
| 14 | Interactive Neo4j graph UI | ✅ Done | `/graph` force-directed view, 7 views, JARM/favicon/issuer links |
| 15 | Storage archival & pruning | ✅ Done | `snapshot_blobs`, retention policy, dedup hash excludes volatile fields |
| 28 | Alert rules & gatherer settings | ✅ Done | Severity/type filters, scan concurrency, per-gatherer timeouts in Settings |


---

## Hard

| # | Item | Status | Notes |
|---|------|--------|-------|
| 12 | BGP hijack risk (RIPEstat) | ✅ Done | `hijack_risk` heuristic + `path_profile` stability |
| — | Custom alerting & rules engine | ✅ Done | Snapshot diff rules, graph drift (G13), webhook filters, per-integration targeting |
| — | Graph UX overhaul | ✅ Done | Temporal topology slider, compare modes, canvas toolbar, cross-view inspector, dense tooltips |

---

## Very Hard

| # | Item | Status | Notes |
|---|------|--------|-------|
| — | Distributed scanning agents | ❌ Later | Multi-region agent binary |
| — | Authentication & RBAC | ❌ Later | OIDC/SAML, multi-tenant SOC |

---

## Graph & intel enhancements

| # | Item | Status | Notes |
|---|------|--------|-------|
| G1 | CT subdomain graph | ✅ Done | `Subdomain` nodes, `DISCOVERED_VIA_CT`, `/graph?view=ct` |
| G2 | RPKI / hijack risk on BGP graph | ✅ Done | Risk-colored `VISIBLE_ORIGIN`/`ANNOUNCES` edges, prefix risk labels |
| G3 | Shared-hop convergence | ✅ Done | `SharedHop` nodes, `SHARED_AT` edges, shared hop panel |
| G4 | Historical route diff | ✅ Done | `route_diff` on graph API when target filtered (BGP/traceroute) |
| G5 | DNS dependency graph | ✅ Done | `DNSHost` nodes, NS/MX/CNAME edges, `/graph?view=dns` |
| G6 | Geo map for hops | ✅ Done | RIPEstat geoloc on hops, map on traceroute view |
| G7 | AS-path enrichment | ✅ Done | RIPEstat `bgp-state` paths, BGP graph + panel |
| G8 | Certificate SAN overlap graph | ✅ Done | `CertSAN` nodes, `/graph?view=cert`, shared SAN panel |
| G9 | Port / banner graph | ❌ Skipped | Depends on #6 port/banner gatherer |
| G10 | Community detection | ✅ Done | Infra clusters (2+ shared signals) on `/graph` infrastructure view |
| G11 | Multi-vantage traceroute | ✅ Done | Local + external vantages, divergence panel |
| G12 | PeeringDB / IX map | ✅ Done | PeeringDB enrich, `/graph?view=peering`, IX panel |
| G13 | Rules engine on graph changes | ✅ Done | BGP origin/path, RPKI, traceroute drift in `change_details` + alert rules |
| G14 | Screenshot timeline | ✅ Done | `/api/targets/:id/screenshots`, intel timeline tab |
| G15 | Wire new intel into graph | ✅ Done | `IntelEvent`/`EnrichmentHit` Neo4j sync, DMARC nodes, `/graph` events panel |
| G16 | Graph PNG/SVG export | ✅ Done | Canvas toolbar export buttons + `graph-export.ts` |
| G17 | Saved graph views | ✅ Done | Named lenses: filters, history, pins, layout restore |

---

## Platform & workflows

| # | Item | Status | Notes |
|---|------|--------|-------|
| — | Investigation notes | ✅ Done | Archive/trash/30-day purge, references, `/notes` + embedded panels |
| — | Target collections | ✅ Done | CRUD API, bulk rescan, `/collections` UI |
| — | Settings tooltips | ✅ Done | Dense `HelpTip` coverage on Settings page |
| — | JA3S CI hotfix | ✅ Done | Inline TLS parse; no `gopacket` / `libpcap` on Linux CI |
| — | Snapshot create-report UX | ✅ Done | Poll status, in-card progress, credentialed PDF download (beta.27) |
| — | Report state hydration | ✅ Done | Snapshot page restores latest report on load (beta.28) |
| — | Scan-form report download | ✅ Done | Home scan result card Download PDF parity (beta.28) |
| — | Richer PDF reports | ✅ Done | beta.28 foundation (logo, screenshot, changes, DNS/TLS blocks) |
| — | Report E2E on snapshot page | ✅ Done | Playwright waits on snapshot + `snapshot-download-report-button` (beta.28) |
| — | PDF report overhaul | ✅ Done | beta.30 — wrapping key/value rows, paginated raw WHOIS, full intel sections (DNS w/ mail posture, BGP routing, TLS chain, web/favicon/crawl/storage/CT/traceroute), pWhois from snapshot row |
| — | Sitemap index recursion | ✅ Done | beta.31 — BFS child sitemaps, common path fallbacks, PDF crawl highlights |
| — | PDF enrichment + pWhois record | ✅ Done | beta.31 — `raw_data.enrichment` section, full `pwhois_data`, sitemap/bucket highlight counts |
| — | THIRD_PARTY.md | ✅ Done | beta.37 — Go/npm licenses, Docker image terms, external API cost models |
| — | Node 24 + multi-arch CI | ✅ Done | beta.38 — Node 24 LTS; native cross-compile for linux/amd64 + arm64 Docker images |
| — | PDF report polish | ✅ Done | TOC page, host-date download filenames, cleaner intel layout, GitHub project URL on cover |
| — | Profile + Settings nav | ✅ Done | Top menu Profile (`/profile`); admin Settings → `/admin/system` |
| — | Snapshot & report delete | ✅ Done | List trash actions; `DELETE /api/snapshots/:id` (admin), `DELETE /api/reports/:id` (scanner+) |
| — | Traceroute privacy | ✅ Done | Redact scanner LAN + first public ISP hop on local paths; Settings toggle (Scan Performance) |
| — | PWA / mobile install | ✅ Done | `manifest.webmanifest`, service worker shell cache, safe-area layout, 192/512 icons |

---

## Collection & reporting backlog

Passive recon gaps, wiring fixes, and follow-on UI/PDF/graph work.

### P1 — Quick wins (easy, high signal) ✅

| # | Item | Status | Notes |
|---|------|--------|-------|
| C1 | `security.txt` gatherer | ✅ Done | beta.32 — `/.well-known/security.txt` + `/security.txt`; parsed contacts, policy, canonical, expires |
| C2 | Plain-text sitemaps | ✅ Done | beta.32 — `/sitemap.txt` path + one-URL-per-line parser |
| C3 | MTA-STS + TLS-RPT | ✅ Done | beta.32 — `MTA_STS`, `TLS_RPT` DNS fields; mail posture scoring updated |
| C4 | Fix HIBP email extraction | ✅ Done | beta.32 — reads `security_txt.contacts` and `web.contact_emails` |
| C5 | Emails from web page text | ✅ Done | beta.32 — `web.contact_emails` from visible page text |
| C6 | CAA DNS records | ✅ Done | beta.32 — CAA type 257 lookup with zone walk |
| C7 | Redirect chain | ✅ Done | beta.32 — `web.redirect_chain` via HEAD/GET follow (http + https) |
| C8 | Extra security headers | ✅ Done | beta.32 — Permissions-Policy, Referrer-Policy, Cross-Origin-* |

### P2 — Depth (medium effort) ✅

| # | Item | Status | Notes |
|---|------|--------|-------|
| C9 | `humans.txt` + `ads.txt` | ✅ Done | beta.33 — `/humans.txt`, `/ads.txt`, `/app-ads.txt`; crawl tab + PDF |
| C10 | BIMI DNS | ✅ Done | beta.33 — `default._bimi` TXT; mail posture grade |
| C11 | PTR for IP targets | ✅ Done | beta.33 — reverse DNS when scan target is raw IP |
| C12 | Expanded DKIM selectors | ✅ Done | beta.33 — 15 default selectors + Settings `dkim_selectors` |
| C13 | Storage from rendered HTML | ✅ Done | beta.33 — bucket/provider hints from Chrome HTML merged into storage |
| C14 | Meta / OG tags | ✅ Done | beta.33 — `web.meta` from rendered page |
| C15 | RDAP WHOIS fallback | ✅ Done | beta.33 — RDAP when classic WHOIS thin; registrant/abuse fields |
| C16 | CDN / email provider labels | ✅ Done | beta.33 — `dns.INFRA_LABELS` in intel highlights + PDF |
| C17 | Provider hints in HTML | ✅ Done | beta.33 — Firebase, Supabase, CDN URL patterns in storage/web |
| C18 | Split A vs AAAA in DNS | ✅ Done | beta.33 — separate `A` (v4) and `AAAA` (v6) records |

### P3 — Enrichment & async intel ✅

| # | Item | Status | Notes |
|---|------|--------|-------|
| E1 | Shodan host/IP lookup | ✅ Done | beta.34 — `shodan.host` IP profile + favicon hash search |
| E2 | Censys host/cert by IP | ✅ Done | beta.34 — `censys.host` by IP + JARM search |
| E3 | Enrichment-before-report UX | ✅ Done | beta.34 — pending badge, snapshot poll, `wait_for_enrichment` report option |
| E4 | Enrichment diff + alerts | ✅ Done | beta.34 — `change_details` + webhooks when enrichment changes |
| E5 | Wayback/CDX URLs | ✅ Done | beta.34 — Internet Archive CDX passive URL list |
| E6 | Optional VT passive DNS | ✅ Done | beta.34 — `virustotal_api_key` in Settings |

### P4 — TLS, CT, DNS hardening ✅

| # | Item | Status | Notes |
|---|------|--------|-------|
| C19 | CT cert metadata | ✅ Done | beta.35 — issuer, serial, validity per cert from crt.sh (`ct.certificates`) |
| C20 | DNSSEC status | ✅ Done | beta.35 — DNSKEY/DS lookup with DO bit → `dns.DNSSEC` |
| C21 | OCSP stapling + TLS version | ✅ Done | beta.35 — `tls_version`, `negotiated_cipher`, `ocsp_stapled` on TLS gatherer |
| C22 | HSTS preload check | ✅ Done | beta.35 — hstspreload.org API when HSTS header present |
| C23 | Cookie name fingerprint | ✅ Done | beta.35 — `web.cookie_names` from Set-Cookie + document (no values) |

### P5 — Reporting, UI & platform wiring ✅

New gatherers must land in **intel tabs** (`intel.ts` field lists), **PDF renderer**, **snapshot diffs**, **webhooks**, and **Neo4j** where applicable.

| # | Item | Status | Notes |
|---|------|--------|-------|
| R1 | Crawl/DNS intel tab fields | ✅ Done | beta.32 — P1 fields in `intel.ts`, intel tabs, and PDF renderer |
| R2 | PDF enrichment formatting | ✅ Done | beta.36 — Shodan/Censys port lists, vuln IDs, favicon/JARM match host summaries |
| R3 | PDF “pending enrichment” note | ✅ Done | beta.36 — footnote + unit test when enrichment.status is pending |
| R4 | Intel highlights parity | ✅ Done | beta.32 — summary cards for security contact, MTA-STS, CAA, redirect hops |
| R5 | `change_details` for new fields | ✅ Done | beta.36 — CAA, DNSSEC, MTA-STS, TLS-RPT, BIMI, security.txt, redirect, cookies, CT certs, TLS hardening |
| R6 | Webhook payloads | ✅ Done | beta.36 — mail/DNS security alert preset + expanded match_types help |
| R7 | Neo4j sync | ✅ Done | beta.36 — SecurityContact, CAA, MTA-STS, DNSSEC, BIMI, WaybackURL nodes |
| R8 | README + help copy | ✅ Done | beta.36 — README features + intel/help entries for P1–P4 gatherers |
| R9 | Gatherer unit + integration tests | ✅ Done | beta.32 — parser tests for security.txt, MTA-STS, TLS-RPT, CAA, plain sitemaps, emails |
| R10 | E2E report coverage | ✅ Done | beta.36 — Playwright PDF size assertion after full report sections |

### P6 — Mobile & platform follow-ons

| # | Item | Status | Notes |
|---|------|--------|-------|
| P6-1 | PWA install + shell cache | ✅ Done | Standalone manifest, `sw.js` (skips `/api/`), safe-area nav/footer |
| P6-2 | Maskable / adaptive icons | ✅ Done | Dedicated maskable PNGs, refreshed Apple touch icon, `scripts/generate-pwa-icons.py` |
| P6-3 | Offline snapshot read cache | ❌ Later | SW cache visited snapshot pages or key GET responses for flaky mobile links |
| P6-4 | Bulk delete on list pages | ✅ Done | Row checkboxes, select-all, bulk action bar on snapshots (admin) and reports (scanner+) |
| P6-5 | E2E auth + delete + CI | ✅ Done | Virtual WebAuthn setup, auth/delete/report specs, GitHub Actions `e2e` job |
| P6-6 | DEPLOY.md + release checklist | ✅ Done | Production secrets, backup, upgrade path, pre-tag test gate |

### Explicitly out of scope (unchanged)

| # | Item | Status | Notes |
|---|------|--------|-------|
| — | Port/banner grabbing | ❌ Skipped | Active probing; see #6 |
| — | Subdomain brute force | ❌ Skipped | Active DNS guessing |
| — | Vulnerability scanning | ❌ Skipped | Different product category |
| — | Port / banner graph (G9) | ❌ Skipped | Depends on port gatherer |

---

## Summary

| | Count |
|---|------|
| **Done** | 106 items (core + graph + platform + P1–P6 through beta.43) |
| **Next** | 0 (P6 polish complete) |
| **Skipped** | 4 (#6 port scan, G9 port graph, subdomain brute, vuln scan) |
| **Later / platform** | 3 (distributed agents, OIDC/RBAC, P6-3 offline cache) |

**Suggested next picks:** P6-3 offline cache (mobile) or distributed agents / OIDC/RBAC (Later tier)