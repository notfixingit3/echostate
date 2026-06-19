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

### P2 — Depth (medium effort)

| # | Item | Status | Notes |
|---|------|--------|-------|
| C9 | `humans.txt` + `ads.txt` | 🔲 Next | `/humans.txt`, `/ads.txt`, `/app-ads.txt` — crawl tab + PDF |
| C10 | BIMI DNS | 🔲 Next | `default._bimi` TXT; tie to DMARC posture grade |
| C11 | PTR for IP targets | 🔲 Next | Reverse DNS when scan target is raw IP (ASN tab complement) |
| C12 | Expanded DKIM selectors | 🔲 Next | Beyond 6 hardcoded selectors; optional Settings list |
| C13 | Storage from rendered HTML | 🔲 Next | Bucket regex on Chrome-captured HTML/JS, not only plain `GET /` (SPA leaks) |
| C14 | Meta / OG tags | 🔲 Next | `description`, `generator`, `canonical`, `og:*` from existing Chrome scrape |
| C15 | RDAP WHOIS fallback | 🔲 Next | When classic WHOIS thin/fails; registrant/abuse fields |
| C16 | CDN / email provider labels | 🔲 Next | Summarize CNAME→Cloudflare/Fastly/Akamai, MX→Google/M365 in intel highlights |
| C17 | Provider hints in HTML | 🔲 Next | Firebase, Supabase, CloudFront, Fastly URL patterns in storage/web gatherers |
| C18 | Split A vs AAAA in DNS | 🔲 Next | Store/report v4 and v6 separately (today both land in `A` via `LookupIPAddr`) |

### P3 — Enrichment & async intel

| # | Item | Status | Notes |
|---|------|--------|-------|
| E1 | Shodan host/IP lookup | 🔲 Next | Enrich resolved IP, not only favicon-hash search |
| E2 | Censys host/cert by IP | 🔲 Next | Complement JARM-only search |
| E3 | Enrichment-before-report UX | 🔲 Next | Snapshot UI: “enrichment pending” badge; optional wait/re-queue report after enrichment worker |
| E4 | Enrichment diff + alerts | 🔲 Next | `change_details` when Shodan/Censys/HIBP/RiskIQ results change between snapshots |
| E5 | Wayback/CDX URLs | 🔲 Next | Passive historical URL list for domain (no crawl) |
| E6 | Optional VT passive DNS | 🔲 Next | Third-party key in Settings (lower priority than E1–E2) |

### P4 — TLS, CT, DNS hardening

| # | Item | Status | Notes |
|---|------|--------|-------|
| C19 | CT cert metadata | 🔲 Next | Issuer, serial, not-before/not-after per cert from crt.sh (not just subdomain names) |
| C20 | DNSSEC status | 🔲 Next | DO bit / chain validation summary on DNS tab |
| C21 | OCSP stapling + TLS version | 🔲 Next | Prominent negotiated version/cipher; staple status on TLS gatherer |
| C22 | HSTS preload check | 🔲 Next | Chromium preload list lookup when HSTS header present |
| C23 | Cookie name fingerprint | 🔲 Next | Names only (no values) from document load — session stack hints |

### P5 — Reporting, UI & platform wiring

New gatherers must land in **intel tabs** (`intel.ts` field lists), **PDF renderer**, **snapshot diffs**, **webhooks**, and **Neo4j** where applicable.

| # | Item | Status | Notes |
|---|------|--------|-------|
| R1 | Crawl/DNS intel tab fields | ✅ Done | beta.32 — P1 fields in `intel.ts`, intel tabs, and PDF renderer |
| R2 | PDF enrichment formatting | 🔲 Next | Human-readable Shodan/Censys match summaries (not truncated map dumps) |
| R3 | PDF “pending enrichment” note | 🔲 Next | Footnote when report generated before async enrichment completes |
| R4 | Intel highlights parity | ✅ Done | beta.32 — summary cards for security contact, MTA-STS, CAA, redirect hops |
| R5 | `change_details` for new fields | 🔲 Next | Diff rules for CAA, security.txt, MTA-STS, redirect chain, BIMI, etc. |
| R6 | Webhook payloads | 🔲 Next | Include new change types in structured webhook entries + alert-rule filters |
| R7 | Neo4j sync | 🔲 Next | Graph nodes/edges for security.txt contacts, MTA-STS, CAA, Wayback URLs as needed |
| R8 | README + help copy | 🔲 Next | Document new gatherers in README features list and Settings `help-copy.ts` |
| R9 | Gatherer unit + integration tests | ✅ Done | beta.32 — parser tests for security.txt, MTA-STS, TLS-RPT, CAA, plain sitemaps, emails |
| R10 | E2E report coverage | 🔲 Next | Extend Playwright report spec when new PDF sections ship |

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
| **Done** | 66 items (core + graph + platform + P1 batch through beta.32) |
| **Next** | 22 (P2–P5 backlog: C9–C23, E1–E6, R2–R3, R5–R8, R10) |
| **Skipped** | 4 (#6 port scan, G9 port graph, subdomain brute, vuln scan) |
| **Later / platform** | 2 (distributed agents, OIDC/RBAC) |

**Suggested next picks:** C9 `humans.txt`/`ads.txt` → C13 storage from Chrome HTML → C15 RDAP fallback → E1 Shodan IP lookup → E3 enrichment/report UX