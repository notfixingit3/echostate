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

---

## Easy → Medium

| # | Item | Status | Notes |
|---|------|--------|-------|
| 16 | Pushover notifications | ✅ Done | Webhook type + documented Settings fields |
| 5 | Subdomain enum via crt.sh | ✅ Done | crt.sh gatherer, CT tab, click-to-scan subdomains |
| 7 | Visual screenshots | ✅ Done | chromedp JPEG thumbnails + timeline UI |

---

## Medium

| # | Item | Status | Notes |
|---|------|--------|-------|
| 1 | TLS/SSL certificate analysis | ✅ Done | Leaf cert + TLS tab; *future:* full chain, expiry alerts |
| 2 | DNS security & deep dive | ✅ Done | A/AAAA, MX, NS, TXT, CNAME, DMARC; *future:* SOA, SPF/DKIM parse |
| 3 | Tech stack fingerprinting | ✅ Done | Final-URL headers + HTML CMS hints; *future:* JS analysis, auto-tagging |
| 8 | Favicon MMH3/SHA256 hashing | ✅ Done | Shodan-compatible; *future:* Shodan/Censys correlation |
| 6 | Port/banner grabbing | ❌ Planned | TCP checks on 22, 21, 80, 443, 3389 |
| — | Automated scheduling (cron) | ❌ Later | Tag-based scan policies |

---

## Medium → Hard

| # | Item | Status | Notes |
|---|------|--------|-------|
| 10 | JARM / JA3 fingerprinting | ✅ Done | JARM on 443; *future:* JA3S, graph linking |
| 13 | API key vault | ❌ Planned | Shodan, HIBP, RiskIQ keys in Settings |
| 15 | Storage archival & pruning | ❌ Planned | Retention policies, blob offload |
| 14 | Interactive Neo4j graph UI | ✅ Done | `/graph` force-directed view, `/api/graph`, JARM/favicon/issuer links |

---

## Hard

| # | Item | Status | Notes |
|---|------|--------|-------|
| 12 | BGP hijack risk (RIPEstat) | ✅ Done | `hijack_risk` heuristic + ASN tab; *future:* path profiling |
| — | Custom alerting & rules engine | ❌ Later | Diff rules, webhook filtering |

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
| G9 | Port / banner graph | ❌ Planned | Depends on #6 port/banner gatherer |
| G10 | Community detection | ✅ Done | Infra clusters (2+ shared signals) on `/graph` infrastructure view |
| G11 | Multi-vantage traceroute | ✅ Done | Local + HackerTarget vantages, divergence panel |
| G12 | PeeringDB / IX map | ✅ Done | PeeringDB enrich, `/graph?view=peering`, IX panel |
| G13 | Rules engine on graph changes | ❌ Planned | Webhook on origin/path/graph drift |
| G14 | Screenshot timeline | ✅ Done | `/api/targets/:id/screenshots`, intel timeline tab |

---

## Summary

| | Count |
|---|------|
| **Done** | 27 roadmap items (15 core + 12 graph) |
| **Planned (near-term)** | 3 (#6, #13, #15) + 1 graph (G9) |
| **Platform (later)** | 4 (cron, rules, agents, auth) |

**Suggested next picks (easiest open items):** #6 port banners → G9 port graph → G13 rules engine