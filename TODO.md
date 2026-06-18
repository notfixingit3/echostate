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

### 8. Favicon Hashing (Infrastructure Tracking)
- Download the site's `favicon.ico` and compute its MurmurHash3 (MMH3) fingerprint.
- **Use Case:** Cross-reference the hash with search engines like Shodan or Censys to discover hidden internal subdomains or detect phishing sites stealing corporate logos.

### 9. `robots.txt` & `sitemap.xml` Extraction
- Automatically pull and parse standard web crawler files during the web scrape.
- **Use Case:** Reveal sensitive admin panels or internal API routes that admins attempt to hide via `Disallow` rules. Track these hidden paths over time.

### 10. JARM / JA3 Fingerprinting
- Collect the unique TLS handshake fingerprint (JARM) of the server.
- **Use Case:** Track infrastructure across the internet even if IPs or domains completely change, effectively identifying backend server configurations or malicious C2 servers.

### 11. Cloud Storage Bucket Detection
- Analyze the web scrape to detect references to AWS S3, Azure Blob Storage, or GCP Buckets.
- **Use Case:** Automatically check if discovered buckets are accidentally configured for "public list" access, catching potential data leaks early.

### 12. BGP Route Hijacking Detection
- Actively monitor and profile the ASN Path data collected during scans.
- **Use Case:** Alert when traffic intended for a target suddenly begins routing through an anomalous or unverified ASN, which is a strong indicator of a BGP hijack attack.

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
