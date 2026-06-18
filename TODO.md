# EchoState Reconnaissance Roadmap

## Completed

### 1. TLS/SSL Certificate Analysis ✓
- Leaf certificate: issuer, subject, SANs, validity window, signature algorithm.
- Intel UI: TLS tab + cert expiry/issuer summary cards.
- *Future:* full chain fetch, expiry webhook rules.

### 2. DNS Security & Deep Dive ✓
- Records: A/AAAA, MX, NS, TXT, CNAME, DMARC (`_dmarc.` lookup).
- Configurable resolvers via Settings UI.
- Intel UI: DNS tab + A-record summary card.
- *Future:* SOA, parsed SPF/DKIM fields.

### 3. Tech Stack Fingerprinting ✓
- `Server` and `X-Powered-By` response headers captured in web gatherer.
- Displayed in the Web intel tab (`tech_stack` field).
- *Future:* `<meta>` tags, linked JS analysis, automatic tagging.

### 4. HTTP Security Headers Analysis ✓
- HSTS, CSP, `X-Frame-Options`, `X-Content-Type-Options` extracted into `security_headers`.
- Displayed in the Web intel tab alongside full response headers.
- *Future:* hygiene scoring and downgrade alerting.

### 8. Favicon Hashing (Infrastructure Tracking) ✓
- Downloads favicon (HTML hint or `/favicon.ico`) and computes Shodan-compatible MMH3 + SHA256.
- Intel UI: Favicon tab + MMH3 summary card.
- *Future:* Shodan/Censys correlation from Settings API vault.

### 9. `robots.txt` & `sitemap.xml` Extraction ✓
- Parses `robots.txt` disallow/allow/sitemap directives and sitemap URL lists.
- Intel UI: Crawl tab + crawl path summary card.
- *Future:* Track hidden path changes over time with dedicated diff rules.

### 10. JARM / JA3 Fingerprinting ✓
- JARM TLS server fingerprint via active probes on port 443.
- Intel UI: TLS tab + JARM summary card.
- *Future:* JA3S extraction and cross-target graph linking.

### 11. Cloud Storage Bucket Detection ✓
- Detects AWS S3, Azure Blob, GCP Storage, and DigitalOcean Spaces references in homepage HTML.
- Intel UI: Storage tab + bucket count summary card.
- *Future:* passive public-access checks for discovered buckets.

### 12. BGP Route Hijacking Detection ✓
- RIPEstat routing visibility + RPKI validation compared against Team Cymru origin data.
- `hijack_risk` heuristic (`low` / `medium` / `high`) with explanatory notes in ASN routing block.
- Intel UI: BGP hijack risk summary card + routing details in ASN tab.
- *Future:* historical path profiling and alerting rules.

---

## Planned

### 5. Subdomain Enumeration via Certificate Transparency
- Query public APIs (e.g., `crt.sh`) to discover newly issued certificates for the target domain.
- **Use Case:** Automatically discover endpoints like `dev.example.com` or `vpn.example.com` and suggest adding them as tracking targets.

### 6. Lightweight Port/Banner Grabbing
- Perform non-intrusive TCP connection checks on common ports (22, 21, 80, 443, 3389).
- Extract service banners.
- **Use Case:** Quickly identify accidental exposure of internal services (like SSH or RDP) to the public internet.

### 7. Visual Screenshots
- Leverage `chromedp` to capture a `.png` or `.jpeg` screenshot of the webpage upon scanning.
- **Use Case:** Provide immediate visual context in the dashboard to see how a site's UI changes over time.

### 13. API Key Vault (External Integrations)
- Create a secure vault in the Settings UI to store third-party API keys (e.g., Shodan, HaveIBeenPwned, RiskIQ).
- **Use Case:** Unlock deeper layers of intelligence seamlessly during the automated scan process.

### 14. Interactive Node Graph UI
- Expose the underlying Neo4j relationship data to the frontend via an interactive, drag-and-drop Maltego-style graph.
- **Use Case:** Visually identify lateral infrastructure links (e.g., Target A and Target B share the same JARM fingerprint or certificate).

### 15. Storage Archival & Pruning
- Develop automated database retention policies to offload or delete older snapshots.
- **Use Case:** Prevent database bloat by shifting full HTML scrapes and large JSON blobs to cheap blob storage (AWS S3, MinIO) after 30 days.

---

# To Do Later (Platform Architecture)

## Custom Alerting & Rules Engine
- Build an evaluation engine to evaluate snapshot diffs against user-defined rules.
- **Use Case:** Prevent webhook fatigue by only alerting on critical specific events (e.g., "Alert only if Port 22 opens" or "Alert if new S3 bucket is found").

## Distributed Scanning Agents
- Create a lightweight `EchoState Agent` binary for deploying in geographically diverse cloud regions.
- **Use Case:** Distribute the scanning workload to prevent IP bans from WAFs (like Cloudflare) and get a true global view of the target.

## Automated Scheduling Engine (Cron)
- Build an internal task scheduler to trigger scans automatically based on tagging policies.
- **Use Case:** Automate continuous monitoring without manual intervention (e.g., "Scan 'High-Risk' targets every 6 hours").

## Authentication & RBAC
- Implement Single Sign-On (OIDC/SAML), user accounts, and Role-Based Access Control.
- **Use Case:** Evolve the tool from a single-user utility to a multi-tenant Security Operations Center (SOC) platform.
