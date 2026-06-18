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
| 2 | DNS security & deep dive | ✅ Done | A/AAAA, MX, NS, TXT, CNAME, DMARC, SPF/DKIM parse; *future:* SOA |
| 3 | Tech stack fingerprinting | ✅ Done | Final-URL headers + HTML CMS hints; *future:* JS analysis, auto-tagging |
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
| 10 | JARM / JA3 fingerprinting | ✅ Done | JARM on 443; *future:* JA3S, graph linking |
| 14 | Interactive Neo4j graph UI | ✅ Done | `/graph` force-directed view, 7 views, JARM/favicon/issuer links |
| 15 | Storage archival & pruning | ✅ Done | `snapshot_blobs`, retention policy, dedup hash excludes volatile fields |
| 28 | Alert rules & gatherer settings | ✅ Done | Severity/type filters, scan concurrency, per-gatherer timeouts in Settings |


---

## Hard

| # | Item | Status | Notes |
|---|------|--------|-------|
| 12 | BGP hijack risk (RIPEstat) | ✅ Done | `hijack_risk` heuristic + ASN tab; *future:* path profiling |
| — | Custom alerting & rules engine | ✅ Done | Snapshot diff rules, graph drift (G13), webhook filters, per-integration targeting |
| — | Graph UX overhaul | ❌ Later | Maltego-style canvas, node inspector, time slider, cross-view linking — beyond current 7-tab force graph |

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
| G11 | Multi-vantage traceroute | ✅ Done | Local + HackerTarget vantages, divergence panel |
| G12 | PeeringDB / IX map | ✅ Done | PeeringDB enrich, `/graph?view=peering`, IX panel |
| G13 | Rules engine on graph changes | ✅ Done | BGP origin/path, RPKI, traceroute drift in `change_details` + alert rules |
| G14 | Screenshot timeline | ✅ Done | `/api/targets/:id/screenshots`, intel timeline tab |
| G15 | Wire new intel into graph | ❌ Later | Sync `change_details` signals, enrichment hits, DMARC/TLS drift as graph events |

---

## Summary

| | Count |
|---|------|
| **Done** | 41 items (core + graph + platform batch) |
| **Skipped** | 2 (#6 port scan, G9 port graph) |
| **Later / platform** | 5 (graph UX overhaul, G15, SOA, agents, OIDC/RBAC) |

**Suggested next picks:** G15 wire enrichment into Neo4j → graph UX polish (node inspector) → SOA DNS record