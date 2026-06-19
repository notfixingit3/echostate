# Third-Party Software & Services

EchoState integrates open-source libraries, container images, and external data sources. This document summarizes **direct dependencies** (libraries EchoState imports) and **external services** (network APIs used at scan or enrichment time).

> **Not legal advice.** Verify current license terms and API pricing on each vendor’s site before production use. Vendor policies change; links below were accurate as of **2026-06-18**.

---

## Project license

EchoState itself is released under the [MIT License](LICENSE).

---

## Go dependencies (direct)

These are the modules listed in `go.mod` under `require (` (production imports).

| Module | License | Role in EchoState |
|--------|---------|-------------------|
| [chromedp/chromedp](https://github.com/chromedp/chromedp) | MIT | Headless Chrome automation (screenshots, rendered HTML) |
| [gin-gonic/gin](https://github.com/gin-gonic/gin) | MIT | HTTP API server |
| [gin-contrib/cors](https://github.com/gin-contrib/cors) | MIT | CORS middleware |
| [jackc/pgx/v5](https://github.com/jackc/pgx) | MIT | PostgreSQL driver |
| [neo4j/neo4j-go-driver/v5](https://github.com/neo4j/neo4j-go-driver) | Apache-2.0 | Neo4j graph sync |
| [johnfercher/maroto/v2](https://github.com/johnfercher/maroto) | MIT | PDF report generation |
| [likexian/whois](https://github.com/likexian/whois) + [whois-parser](https://github.com/likexian/whois-parser) | Apache-2.0 | Classic WHOIS lookups |
| [go-webauthn/webauthn](https://github.com/go-webauthn/webauthn) | MIT | Passkey authentication |
| [georgestarcher/pwhois](https://github.com/georgestarcher/pwhois) | MIT | pWhois client library |
| [hdm/jarm-go](https://github.com/hdm/jarm-go) | MIT | JARM TLS fingerprinting |
| [twmb/murmur3](https://github.com/twmb/murmur3) | MIT | Favicon MMH3 hashing |
| [google/uuid](https://github.com/google/uuid) | BSD-3-Clause | UUID generation |
| [golang.org/x/net](https://go.dev/) | BSD-3-Clause | DNS message parsing, networking |
| [golang.org/x/crypto](https://go.dev/) | BSD-3-Clause | Cryptography helpers |
| [golang.org/x/image](https://go.dev/) | BSD-3-Clause | Image handling (screenshots) |
| [stretchr/testify](https://github.com/stretchr/testify) | MIT | Tests only |

**Indirect dependencies** (pulled in transitively, e.g. PDF/font stacks, Gin validators) are recorded in `go.sum`. To regenerate a full license report locally:

```bash
go install github.com/google/go-licenses@latest
go-licenses report ./...
```

---

## JavaScript dependencies (direct)

Production dependencies from `frontend/package.json`:

| Package | License | Role |
|---------|---------|------|
| [next](https://github.com/vercel/next.js) | MIT | Web UI framework |
| [react](https://github.com/facebook/react) / [react-dom](https://github.com/facebook/react) | MIT | UI runtime |
| [@base-ui/react](https://github.com/mui/base-ui) | MIT | Accessible UI primitives |
| [class-variance-authority](https://github.com/joe-bell/cva) | Apache-2.0 | Component variants |
| [clsx](https://github.com/lukeed/clsx) | MIT | Class name helper |
| [tailwind-merge](https://github.com/dcastil/tailwind-merge) | MIT | Tailwind class merging |
| [lucide-react](https://github.com/lucide-icons/lucide) | ISC | Icons |
| [date-fns](https://github.com/date-fns/date-fns) | MIT | Date formatting |
| [next-themes](https://github.com/pacocoursey/next-themes) | MIT | Light/dark theme |
| [react-diff-viewer-continued](https://github.com/aeolun/react-diff-viewer-continued) | MIT | Snapshot JSON diffs |
| [react-force-graph-2d](https://github.com/vasturiano/react-force-graph) | MIT | Graph canvas |
| [recharts](https://github.com/recharts/recharts) | MIT | Charts |
| [shadcn](https://github.com/shadcn-ui/ui) | MIT | Component tooling (dev-time CLI) |
| [tw-animate-css](https://github.com/Wombosvideo/tw-animate-css) | MIT | Tailwind animations |

To list all transitive npm licenses:

```bash
cd frontend && npx license-checker --production --summary
```

---

## Container images (Docker Compose)

| Image | License / terms | Role |
|-------|-----------------|------|
| [postgres:16-alpine](https://hub.docker.com/_/postgres) | [PostgreSQL License](https://www.postgresql.org/about/licence/) | Primary database |
| [neo4j:5](https://neo4j.com/licensing/) | GPL v3 (community) / commercial for some deployments | Graph database |
| [browserless/chrome](https://www.browserless.io/) | Service-specific; image is third-party | Headless Chrome for screenshots & rendered HTML |
| GHCR `echostate` / `echostate-frontend` | MIT (this repo) | Application images |

Neo4j and browserless may require separate commercial agreements depending on deployment size and use case.

---

## External services — passive scan (no EchoState API key)

Used during every scan unless a gatherer times out or is skipped. Respect each provider’s terms of use and rate limits.

| Service | Cost model | EchoState usage | Notes |
|---------|------------|-----------------|-------|
| [Team Cymru DNS](https://www.team-cymru.com/ip-asn-mapping) | Free (DNS-based) | Origin ASN + AS description | `origin.asn.cymru.com`, `asn.cymru.com` |
| [RIPEstat](https://stat.ripe.net/) | Free API | BGP visible origins, AS paths, RPKI, hijack heuristics | Open data; fair-use limits |
| [PeeringDB](https://www.peeringdb.com/) | Free API | Internet exchange (IX) presence | API terms on peeringdb.com |
| [crt.sh](https://crt.sh/) | Free | Certificate transparency subdomains + cert metadata | Public CT mirror; can be slow |
| [Cert Spotter](https://sslmate.com/certspotter/) | Free tier (rate limits) | CT fallback when crt.sh fails | No API key in EchoState today |
| [RDAP.org](https://about.rdap.org/) | Free | WHOIS fallback for thin registries | `rdap.org/domain/...` |
| [hstspreload.org](https://hstspreload.org/) | Free API | HSTS preload list status | Only queried when HSTS header present |
| [Internet Archive CDX](https://web.archive.org/) | Free | Wayback URL history (enrichment without API key) | Used when no paid enrichment keys configured |
| [HackerTarget](https://hackertarget.com/) | Free tier (limits) | Optional external traceroute vantage | May throttle heavy use |
| Target host DNS resolvers | Configurable (default Google/Cloudflare) | A/AAAA/MX/NS/TXT/CAA/DNSSEC/etc. | Set in Settings → DNS servers |
| Target websites | N/A | HTTP(S) fetch: robots, security.txt, sitemaps, headers, redirects | Passive only; no brute force |

Classic WHOIS uses registrars/RIRs via [likexian/whois](https://github.com/likexian/whois); data is subject to registrar terms.

---

## External services — optional enrichment (API keys)

Configured in **Admin → System** or environment variables. **These are generally paid or quota-limited** — budget accordingly.

| Service | Cost model | EchoState usage | Settings field |
|---------|------------|-----------------|----------------|
| [Shodan](https://account.shodan.io/) | **Paid membership** (API credits) | Host profile by IP; favicon hash search | `shodan_api_key` |
| [Censys](https://search.censys.io/account/api) | **Paid** (free tier very limited) | Host profile by IP; JARM fingerprint search | `censys_api_id`, `censys_api_secret` |
| [Have I Been Pwned](https://haveibeenpwned.com/API/Key) | **Paid API** (per-email pricing) | Breach checks on contact emails | `hibp_api_key` |
| [RiskIQ / PassiveTotal](https://community.riskiq.com/) | **Paid / enterprise** | Passive DNS enrichment | `riskiq_api_user`, `riskiq_api_key` |
| [VirusTotal](https://www.virustotal.com/gui/join-us) | **Freemium** (strict daily limits on free tier) | Passive DNS resolutions for domains | `virustotal_api_key` |

Enrichment runs **asynchronously** after each snapshot. Without keys, Wayback CDX may still run for domain targets (free).

---

## External services — notifications (operator-configured)

| Service | Cost model | EchoState usage |
|---------|------------|-----------------|
| [Pushover](https://pushover.net/) | **Paid app** (one-time per platform) + optional team features | Mobile push via user token in webhook config |
| Slack / Discord / MS Teams | Free incoming webhooks (workspace policies apply) | Snapshot `change_details` alerts |

---

## External services — infrastructure & misc

| Service | Cost model | EchoState usage |
|---------|------------|-----------------|
| [pWhois](https://pwhois.org/) | Public WHOIS-style IP service (see site terms) | Async enrichment of **submitter client IPs** |
| [GitHub API](https://docs.github.com/en/rest) | Free within rate limits | Optional upstream version check (`ECHOSTATE_UPDATE_CHECK`) |

---

## Data handling reminders

- **API keys** are stored in PostgreSQL settings (masked in API responses). Use secrets management in production.
- **Passive recon** does not port-scan targets, but third-party APIs (Shodan/Censys) may reflect **their own** prior observations of the internet.
- **WHOIS/RDAP/pWhois** data often has redistribution restrictions — use for operational security, not bulk republishing.
- **HIBP** requires a paid key and prohibits using breach data for unauthorized purposes.

---

## Updating this document

When adding a gatherer, enrichment provider, npm/go dependency, or Compose service:

1. Add a row to the appropriate table with **cost model**, **usage**, and **configuration**.
2. Link to the vendor’s license/terms page.
3. Mention new optional keys in `.env.example` and `README.md`.