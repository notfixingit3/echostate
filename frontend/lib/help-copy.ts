export const HELP_COPY: Record<string, string> = {
  "page.home":
    "Enter a hostname, IP, or URL to start a passive recon scan. Results are stored as snapshots you can diff over time.",
  "page.targets":
    "All scanned hosts EchoState has tracked. Tags and snapshot counts summarize how each target has evolved.",
  "page.snapshots":
    "Every scan result across all targets, newest first. Identical rescans update last_seen instead of creating duplicates.",
  "page.graph":
    "Interactive relationship maps synced from Neo4j. Filter by target and scrub snapshot history to see how topology changed.",
  "page.settings":
    "Runtime configuration for DNS, enrichment keys, scan workers, retention, alert rules, and outbound webhooks.",
  "page.profile": "Your personal settings: theme, timezone, passkeys, and registered devices.",
  "page.administration":
    "Server-wide configuration: integrations, system tuning, authentication defaults, and user management.",
  "settings.nav.data": "Export or import targets, snapshots, collections, notes, and graph views between installs.",
  "settings.export_all":
    "Downloads a JSON bundle with investigation data. Optional blobs and server settings can be included.",
  "settings.import_bundle":
    "Upload a bundle from another EchoState instance. Choose overwrite or skip when records already exist.",
  "settings.nav.integrations": "Slack, Discord, Teams, and Pushover webhook integrations.",
  "settings.nav.system": "DNS, scan workers, scheduler, retention, enrichment API keys, and alert rules.",
  "settings.nav.authentication": "Passkey enrollment defaults, session lifetime, and WebAuthn relying party settings.",
  "settings.nav.users": "Create users and issue enrollment or recovery codes.",
  "settings.nav.audit":
    "Read-only trail of sign-ins, deletes, settings changes, imports/exports, and user management.",
  "settings.profile": "Your display name and role in EchoState.",
  "settings.preferences": "Default theme and timezone for your account across devices.",
  "settings.passkeys": "Issue device codes, register passkeys, rename, or remove credentials.",
  "settings.device_code":
    "Single-use enrollment code for signing in on another browser or phone, then registering a passkey there.",
  "page.reports":
    "PDF intelligence reports generated from snapshot data. Reports queue asynchronously after you request them.",

  "scan.host":
    "Hostname, IP address, or full URL. Hostnames must include a domain suffix (e.g. example.com). EchoState normalizes input and gathers WHOIS, ASN, DNS, TLS, and web intel concurrently.",
  "scan.submit":
    "Queues an async scan job. Poll until complete, then open the target or snapshot detail page.",
  "scan.status":
    "Shows queued, running, completed, or failed state for the current scan job.",

  "targets.add":
    "Add a host or IP to surveillance without running a scan. Run a scan from Home or the target detail page when you want intel.",
  "targets.host": "Canonical normalized hostname for this target.",
  "targets.tags": "Labels you assign plus auto-tags such as wordpress when signals are detected.",
  "targets.snapshots": "Number of stored scans for this host.",
  "targets.last_seen": "When this target was last observed in a scan.",
  "targets.created": "When this target was first added.",

  "snapshots.host": "Target hostname for this scan.",
  "snapshots.scanned_at": "When this snapshot was captured.",
  "snapshots.changes": "Human-readable diff summaries versus the previous snapshot.",
  "snapshots.client_ip": "Submitter IP enriched asynchronously via pWhois.",
  "snapshots.hash": "Content hash used to detect identical rescans.",

  "graph.view.infra":
    "Shared infrastructure links: ASN, IP, JARM, JA3S, favicon hash, and certificate issuer.",
  "graph.view.bgp":
    "BGP origin, announced prefix, visible origins, and RIPEstat AS paths with risk coloring.",
  "graph.view.traceroute":
    "Hop-by-hop paths from local and external vantages, including shared transit hops.",
  "graph.view.peering":
    "Internet exchange presence for each target ASN from PeeringDB.",
  "graph.view.ct": "Certificate transparency subdomains discovered for the target.",
  "graph.view.dns":
    "Authoritative DNS dependencies: NS, MX, CNAME, SOA zone, and DMARC policy.",
  "graph.view.cert": "TLS subject alternative names and cross-target SAN overlap.",
  "graph.target_filter":
    "Limit the graph to one target. Required for snapshot history and time-travel topology.",
  "graph.vantage_filter": "Show traceroute paths from a single vantage or all vantages.",
  "graph.refresh": "Reload graph data from the API without changing filters.",
  "graph.history_slider":
    "Scrub across snapshots for the selected target. The canvas reflects that point in time.",
  "graph.compare_previous":
    "Highlight nodes and edges that changed versus the next-older snapshot.",
  "graph.compare_latest":
    "Highlight drift versus the most recent scan while viewing an older snapshot.",
  "graph.toolbar.zoom_in": "Zoom into the canvas.",
  "graph.toolbar.zoom_out": "Zoom out of the canvas.",
  "graph.toolbar.fit": "Fit the full graph into view.",
  "graph.toolbar.reset": "Clear pinned node positions and re-run the layout simulation.",
  "graph.toolbar.fullscreen": "Expand the graph canvas to fullscreen.",
  "graph.toolbar.export_png":
    "Download the current canvas view as a PNG image, including zoom and pan state.",
  "graph.toolbar.export_svg":
    "Download node positions and edges as a lightweight SVG for reports or slides.",
  "graph.saved_views":
    "Save the current graph lens — view tab, target filter, snapshot compare, vantage, pinned layout, and selection.",
  "graph.rescan_hint":
    "Temporal graph edges are versioned per snapshot. Rescan to accumulate history for the slider and compare modes.",

  "page.collections":
    "Watchlists of related targets. Add members and queue bulk rescans from one place.",
  "collections.create": "Name and optional description for a new target collection.",
  "collections.list": "Select a collection to add or remove targets and queue bulk rescans.",
  "collections.rescan":
    "Enqueue an async scan job for every target in the selected collection.",

  "page.notes":
    "Investigation notes across the platform. Filter by status, archive finished work, or recover from trash within 30 days.",
  "notes.panel":
    "Human context on top of passive intel — findings, hypotheses, and links to external references.",
  "notes.compose":
    "Markdown-friendly body with optional HTML anchors and structured reference chips.",
  "notes.references":
    "Attach URLs, targets, snapshots, collections, or graph node IDs as clickable reference chips.",
  "notes.trash":
    "Trashed notes are permanently deleted after 30 days, or immediately via Delete forever.",

  "intel.cert_expires": "TLS certificate not-after date from the leaf cert.",
  "intel.cert_issuer": "Certificate authority that signed the active TLS certificate.",
  "intel.dns_a": "IPv4 addresses returned by DNS resolution at scan time.",
  "intel.dns_aaaa": "IPv6 addresses returned by DNS resolution at scan time.",
  "intel.dns_ptr": "Reverse DNS (PTR) names when the scan target is a raw IP address.",
  "intel.dns_soa": "Authoritative SOA zone and serial for the target.",
  "intel.cdn_provider": "CDN inferred from CNAME target (Cloudflare, Fastly, Akamai, etc.).",
  "intel.mail_provider": "Hosted email provider inferred from MX records.",
  "intel.bimi": "BIMI brand-indicator TXT at default._bimi; included in mail posture grade.",
  "intel.meta_description": "HTML meta description and Open Graph tags from the rendered page.",
  "intel.provider_hints": "Third-party service URL patterns (Firebase, Supabase, CDN hosts) in page HTML.",
  "intel.favicon_mmh3":
    "MurmurHash3 of the site's favicon (Shodan-compatible). Identical hashes often mean the same default icon, admin panel, or product across unrelated hosts.",
  "intel.jarm":
    "JARM fingerprints how a TLS server on port 443 responds to a standard probe sequence. Hosts behind the same load balancer, CDN config, or product often share a JARM hash — useful for linking infrastructure.",
  "intel.ja3s":
    "JA3S fingerprints the server's TLS handshake (cipher suites, extensions, curves). Similar servers publish the same JA3S string; compare it across targets to spot shared TLS stacks.",
  "intel.hijack_risk":
    "Estimated BGP hijack or route-leak risk from RIPEstat visible-origin data. High values mean unexpected ASNs are announcing this prefix.",
  "intel.bgp_path_stability":
    "BGP AS-path diversity: stable (one path), diverse (2–4), or volatile (5+).",
  "intel.mail_posture":
    "Composite SPF, DKIM, and DMARC grade (A–F) for outbound email security posture.",
  "intel.js_assets":
    "Passive script and stylesheet URLs with library/CDN hints from page assets.",
  "intel.wp_plugins": "WordPress plugins detected passively from page assets.",
  "intel.wp_themes": "Active WordPress theme slug and version when available.",

  "feature.whois":
    "Domain registration data: registrar, nameservers, expiration, and contact fields from WHOIS or RDAP.",
  "feature.asn_bgp":
    "Routing context for the resolved IP: origin ASN, announced prefix, country, and allocation registry.",
  "feature.web_intel":
    "Passive webpage signals: title, final URL, copyright text, and related metadata from the landing page.",

  "field.jarm":
    "JARM fingerprints how a TLS server on port 443 responds to a standard probe sequence. Shared hashes often indicate the same load balancer, CDN, or product.",
  "field.ja3s":
    "JA3S fingerprints the TLS ServerHello (ciphers, extensions, curves). Identical JA3S strings suggest the same TLS software or configuration.",
  "field.mmh3":
    "MurmurHash3 of the favicon image. Used by Shodan and EchoState to match hosts that serve the same icon.",
  "field.asn":
    "Autonomous System Number — the BGP routing identity of the network that announces this IP prefix.",
  "field.as_name": "Registered name of the ASN holder (e.g. CLOUDFLARENET).",
  "field.prefix":
    "BGP-announced CIDR block containing the target IP (e.g. 192.0.2.0/24).",
  "field.hijack_risk":
    "Heuristic hijack/leak score from RIPEstat visible origins for this prefix.",
  "field.visible_origins":
    "ASNs currently seen announcing this prefix in global BGP, per RIPEstat.",
  "field.path_profile":
    "Summary of AS-path diversity toward this prefix: stable, diverse, or volatile.",
  "field.rpki_status":
    "Resource Public Key Infrastructure validation state for the route (valid, invalid, or unknown).",
  "field.ocsp_stapled":
    "Whether the server attaches a fresh OCSP revocation response in the TLS handshake.",
  "field.tls_version": "Negotiated TLS protocol version (e.g. TLS 1.3).",
  "field.negotiated_cipher": "Cipher suite chosen for the TLS session.",
  "field.dns_names": "Subject Alternative Names (SANs) listed on the TLS certificate.",
  "field.ip_sans": "IP addresses listed as SANs on the TLS certificate.",
  "field.dnssec_status":
    "Whether the zone publishes DNSSEC keys (signed) or not (unsigned).",
  "field.dmarc":
    "DMARC email authentication policy published in DNS (_dmarc TXT).",
  "field.spf": "SPF record listing senders allowed to mail for this domain.",
  "field.dkim": "DKIM public keys discovered for common selectors.",
  "field.mta_sts":
    "MTA-STS policy mode: none, testing, or enforce — controls strict SMTP TLS.",
  "field.bimi":
    "BIMI record for brand logo display in supporting mail clients.",
  "field.caa":
    "Certificate Authority Authorization — DNS rules for which CAs may issue certs.",
  "field.hsts_preload":
    "Whether the domain appears on the browser HSTS preload list.",
  "field.security_txt":
    "Parsed /.well-known/security.txt contacts and policy URLs.",
  "field.humans_txt":
    "humans.txt credits file at /humans.txt — team names and contact hints for the site.",
  "field.ads_txt":
    "ads.txt authorized digital sellers list for programmatic ad inventory.",
  "intel.hibp":
    "Have I Been Pwned — checks whether WHOIS contact emails appear in known public breach databases.",
  "intel.virustotal_pdns":
    "VirusTotal passive DNS — historical hostname resolutions for this domain.",
  "intel.ct_subdomains":
    "Subdomains discovered in certificate transparency logs that were not seen in earlier snapshots.",
  "intel.submitter_ip":
    "IP address of the client that submitted this scan, enriched asynchronously via pWhois.",
  "field.subdomains":
    "Hostnames discovered via certificate transparency logs (crt.sh).",
  "field.certificate_count": "Number of CT log entries found for this domain.",
  "field.mmh3_shodan": "Whether the favicon hash matches a known Shodan favicon entry.",
  "field.origin_as": "ASN of the submitter's IP from pWhois enrichment.",
  "field.org_name": "Organization name for the submitter's IP.",
  "field.rdap_source": "Registry source used for WHOIS (RDAP endpoint or classic WHOIS).",
  "field.tech_stack":
    "Detected frameworks and CMS hints from page HTML and response headers.",
  "field.provider_hints":
    "Third-party service URLs embedded in the page (Firebase, Stripe, etc.).",
  "field.detected_buckets":
    "Public cloud storage URLs referenced in page content.",
  "field.sitemap_urls": "URLs discovered from robots.txt and sitemap files.",
  "field.redirect_chain": "HTTP redirects followed from the initial request to the final URL.",
  "field.js_assets": "JavaScript and CSS URLs loaded by the page.",
  "field.cookie_names": "Cookie names set on the response (values are not stored).",
  "field.peeringdb": "PeeringDB profile: IX presence, facilities, and network name.",
  "field.ix_count": "Number of internet exchanges where this ASN has a presence.",
  "field.vantages":
    "Traceroute vantage points (local scanner vs external probes) used for path data.",

  "graph.node.JARM":
    "TLS server fingerprint node. Targets linked here share the same JARM hash — often the same TLS stack or appliance.",
  "graph.node.Favicon":
    "Favicon hash node. Targets sharing it serve the same favicon bytes (common for shared admin UIs or defaults).",
  "graph.node.CertIssuer":
    "Certificate authority that issued certs for linked targets.",
  "graph.node.CertSAN":
    "Certificate Subject Alternative Name shared across targets.",
  "graph.node.ASN": "Autonomous System — BGP routing identity for linked IPs or prefixes.",
  "graph.node.Prefix": "BGP prefix (CIDR) announced for a target's IP.",
  "graph.node.DMARCPolicy": "DMARC policy record shared or referenced by targets.",
  "graph.node.SOAZone": "DNS zone of authority (SOA) dependency.",
  "graph.node.IX": "Internet exchange point where an ASN peers.",
  "graph.node.SharedHop":
    "Traceroute hop shared by multiple paths — often a common transit provider.",
  "graph.node.Subdomain": "Hostname discovered via certificate transparency.",
  "graph.node.DNSHost": "DNS hostname dependency (NS, MX, CNAME target, etc.).",

  "settings.webhooks":
    "Outbound Slack, Discord, MS Teams, or Pushover integrations for snapshot diff alerts.",
  "settings.webhook_url":
    "Incoming webhook URL from your chat platform. EchoState posts change summaries with target links.",
  "settings.dns_servers":
    "Comma-separated DNS resolvers used by gatherers (default 8.8.8.8, 1.1.1.1).",
  "settings.dkim_selectors":
    "Extra DKIM selector hostnames to probe beyond built-in defaults (comma-separated).",
  "settings.pwhois_server": "pWhois server hostname for async client-IP ASN and org enrichment.",
  "settings.rate_limit": "Maximum API requests per minute per client IP.",
  "settings.scan_concurrency": "Maximum scan jobs processed in parallel by the API worker pool.",
  "settings.scan_timeouts":
    "Wall-clock limits for each scan gatherer in seconds. Applied on the next scan after you save.",
  "settings.timeout_scan_job":
    "Maximum time for a full scan job before the worker cancels it. Should exceed your slowest gatherer timeouts combined.",
  "settings.timeout_gatherer":
    "Default cap for WHOIS, DNS, TLS, web, crawl, and similar gatherers.",
  "settings.timeout_ct": "HTTP timeout budget for certificate transparency log queries.",
  "settings.timeout_traceroute": "Time allowed for local traceroute to finish.",
  "settings.timeout_screenshot": "Time allowed for the headless browser screenshot step.",
  "settings.enrichment_timeouts":
    "HTTP timeouts for async enrichment (Shodan, Wayback, VirusTotal, etc.). Increase on slow or high-latency networks.",
  "settings.timeout_scanner_http":
    "HTTP timeout for in-scan fetches such as robots.txt, security.txt, and header probes.",
  "settings.timeout_enrichment_http":
    "Default HTTP timeout for Shodan, Censys, HIBP, VirusTotal, and other enrichment APIs.",
  "settings.timeout_wayback":
    "Separate timeout for Internet Archive CDX lookups, which are often slower than other enrichment calls.",
  "settings.traceroute_redact_scanner_prefix":
    "Strip private LAN hops and the first public ISP hop from local traceroutes before storing results. External vantage paths are unchanged.",
  "settings.traceroute_redact_local_extra_hops":
    "Additional local hops to hide after the automatic scanner-prefix redaction (0 = none).",
  "settings.retention":
    "Per-target snapshot keep-N policy. Older snapshots are pruned after each scan.",
  "settings.audit_retention":
    "Delete audit log entries older than this many days. 0 keeps all events. Purge runs hourly and when settings are saved.",
  "settings.scheduler":
    "Background job that rescans stale targets on a configurable interval.",
  "settings.schedule_tags":
    "Optional comma-separated tag filter. Only matching targets are scheduled for rescan.",
  "settings.api_key": "Optional write API key required for scan and mutation endpoints.",
  "settings.shodan_key": "Shodan API key for host enrichment after each snapshot.",
  "settings.censys_keys": "Censys API ID and secret for certificate and host correlation.",
  "settings.hibp_key": "Have I Been Pwned API key for breach checks on WHOIS contact emails.",
  "settings.riskiq_keys":
    "RiskIQ / PassiveTotal credentials for passive DNS enrichment on scanned hosts.",
  "settings.virustotal_key":
    "VirusTotal API key for passive DNS resolution history on scanned domains.",
  "intel.enrichment_status":
    "Async third-party enrichment status: pending while Shodan/Censys/HIBP/Wayback run after the scan.",
  "intel.shodan_host": "Shodan host profile for the resolved scan IP (ports, org, vulns).",
  "intel.wayback_urls": "Historical URLs from the Internet Archive CDX index (passive, no crawl).",
  "intel.dnssec":
    "DNSSEC signing status from DNSKEY/DS lookups (signed vs unsigned).",
  "intel.caa":
    "Certificate Authority Authorization records restricting which CAs may issue certs.",
  "intel.security_txt":
    "Parsed security.txt contacts, policy URLs, and expiry from /.well-known/security.txt.",
  "intel.mta_sts":
    "MTA-STS SMTP transport policy (none, testing, or enforce) from _mta-sts DNS.",
  "intel.hsts_preload":
    "hstspreload.org preload list status when Strict-Transport-Security is present.",
  "intel.cookie_names":
    "Cookie names observed on the page (values are not stored).",
  "intel.ct_certificates":
    "Recent certificate transparency issuances from crt.sh with issuer, serial, and validity.",
  "settings.alert_rules":
    "Filter which change_details events trigger webhook notifications.",
  "settings.alert_match_types":
    "Comma-separated change types (e.g. caa_added, dnssec_status, security_contact_added). Leave empty to match all severities above the minimum.",
  "settings.alert_webhooks":
    "Limit delivery to specific integrations. Empty means all enabled webhooks.",
  "settings.auth":
    "Passkey authentication defaults: enrollment codes, session lifetime, rate limits, and WebAuthn relying party.",
  "settings.auth_enabled":
    "When enabled and at least one user exists, the API requires a signed-in session for protected routes.",
  "settings.enrollment_code_ttl": "Hours before an issued enrollment or recovery code expires.",
  "settings.enrollment_code_length": "Digit count for standard enrollment codes (default 8).",
  "settings.recovery_code_length": "Character count for alphanumeric recovery codes.",
  "settings.session_ttl": "Hours before an HTTP session cookie expires.",
  "settings.max_code_attempts": "Failed verification attempts allowed within the attempt window.",
  "settings.code_attempt_window": "Minutes over which failed code attempts are counted.",
  "settings.webauthn_rp_id": "Relying party ID hostname for passkeys. Defaults from FRONTEND_URL when blank.",
  "settings.webauthn_rp_origin": "Allowed WebAuthn origin URL. Defaults to FRONTEND_URL when blank.",
  "settings.auth_users":
    "Create scanner or admin accounts and issue single-use enrollment or recovery codes.",

  "reports.status": "queued, running, completed, or failed PDF generation state.",
  "reports.snapshot": "Snapshot used as the source data for this report.",
  "reports.download": "Download the generated PDF when status is completed.",

  "snapshot.create_report":
    "Queue an async PDF report for this snapshot. Status updates automatically; download when complete. Revisiting the page restores the latest report.",
  "snapshot.wait_for_enrichment":
    "Delay PDF generation until async Shodan, Censys, HIBP, and Wayback enrichment finishes so the report includes third-party data.",
  "snapshot.scanned_at": "When this snapshot was first captured by a scan job.",
  "snapshot.last_seen":
    "When this identical result was last observed. Rescans with the same content hash update last seen instead of creating a duplicate row.",
  "snapshot.target": "Parent target record for this host — open to see full snapshot history.",
  "snapshot.data_hash":
    "Fingerprint of normalized gatherer output. Identical rescans produce the same hash.",
  "snapshot.intel_summary":
    "Condensed highlights from the gatherer payload. Open the tabs below for full WHOIS, TLS, DNS, and enrichment detail.",
  "scan.and_report":
    "Run a scan and queue a PDF report in one step. The result card polls until the report is ready, then offers download.",

  "snapshot.changes":
    "Diffs versus the previous snapshot. Type names the gatherer field (e.g. tls jarm, dns A). Severity: critical = security or cert regression; warning = notable drift; info = minor or expected change.",
  "snapshot.raw_data": "Full JSON payload returned by all gatherers for this scan.",

  "field.host": "Canonical hostname after EchoState normalizes the scan input.",
  "field.resolved_ip": "IP address DNS returned when the scan ran.",
  "field.country":
    "Country associated with the BGP prefix or geolocation data for the resolved IP.",
  "field.registrar": "Domain registrar from WHOIS or RDAP registry data.",
  "field.web_title": "HTML page title from the rendered landing page.",

  "intel.tab.overview":
    "Quick cross-section of WHOIS, routing, web, and TLS highlights from this snapshot.",
  "intel.tab.whois":
    "Domain registration: registrar, contacts, nameservers, expiration, and raw registry text.",
  "intel.tab.asn":
    "BGP routing context — origin ASN, announced prefix, hijack-risk heuristics, and PeeringDB data.",
  "intel.tab.dns":
    "Authoritative and recursive DNS answers: A/AAAA, MX, NS, TXT, DMARC, SPF, DKIM, DNSSEC, and more.",
  "intel.tab.tls":
    "TLS certificate chain, expiry, negotiated version/cipher, and fingerprints (JARM, JA3S) from port 443.",
  "intel.tab.web":
    "HTTP response intel: title, redirects, security headers, tech stack, cookies, and embedded provider URLs.",
  "intel.tab.favicon":
    "Favicon URL and MurmurHash3 (MMH3) used to match hosts serving the same icon.",
  "intel.tab.crawl":
    "Passive fetches of robots.txt, security.txt, humans.txt, ads.txt, and sitemap files.",
  "intel.tab.ct":
    "Certificate Transparency — subdomains and issuances discovered via crt.sh logs.",
  "intel.tab.traceroute":
    "Network paths toward the destination from the scanner and external vantage points.",
  "intel.tab.screenshots":
    "JPEG thumbnails captured by headless Chrome during the scan, plus historical shots for the target.",
  "intel.tab.storage":
    "Cloud storage bucket URLs referenced in page content (S3, GCS, Azure, etc.).",
  "intel.tab.enrichment":
    "Third-party API results: Shodan host profile, Censys, HIBP breaches, Wayback URLs, VirusTotal passive DNS.",
  "intel.tab.pwhois":
    "Async enrichment of the client IP that submitted this scan (ASN, org, city).",
  "intel.tab.errors":
    "Gatherers that timed out or failed — the snapshot may be partial when errors are present.",
  "intel.tab.changes":
    "Structured diffs versus the prior snapshot. Compare type badges (which field) and severity (how important the drift is).",
  "intel.tab.raw_diff":
    "Side-by-side JSON diff of raw gatherer output versus the previous snapshot.",
}

export function getHelpCopy(id: string): string | undefined {
  return HELP_COPY[id]
}

/** Maps raw intel field keys to help-copy IDs for inline tooltips. */
export const FIELD_HELP_IDS: Record<string, string> = {
  jarm: "field.jarm",
  ja3s: "field.ja3s",
  mmh3: "field.mmh3",
  asn: "field.asn",
  as_name: "field.as_name",
  prefix: "field.prefix",
  hijack_risk: "field.hijack_risk",
  visible_origins: "field.visible_origins",
  path_profile: "field.path_profile",
  rpki_status: "field.rpki_status",
  ocsp_stapled: "field.ocsp_stapled",
  tls_version: "field.tls_version",
  negotiated_cipher: "field.negotiated_cipher",
  dns_names: "field.dns_names",
  ip_sans: "field.ip_sans",
  dnssec_status: "field.dnssec_status",
  DMARC: "field.dmarc",
  spf: "field.spf",
  dkim: "field.dkim",
  mta_sts: "field.mta_sts",
  bimi: "field.bimi",
  caa: "field.caa",
  hsts_preload: "field.hsts_preload",
  security_txt: "field.security_txt",
  subdomains: "field.subdomains",
  certificate_count: "field.certificate_count",
  shodan: "field.mmh3_shodan",
  origin_as: "field.origin_as",
  org_name: "field.org_name",
  rdap_source: "field.rdap_source",
  tech_stack: "field.tech_stack",
  provider_hints: "field.provider_hints",
  detected_buckets: "field.detected_buckets",
  sitemap_urls: "field.sitemap_urls",
  redirect_chain: "field.redirect_chain",
  js_assets: "field.js_assets",
  cookie_names: "field.cookie_names",
  peeringdb: "field.peeringdb",
  ix_count: "field.ix_count",
  vantages: "field.vantages",
}

export function getFieldHelpId(key: string): string | undefined {
  return FIELD_HELP_IDS[key] ?? FIELD_HELP_IDS[key.toLowerCase()]
}

export function getGraphNodeHelpId(type: string): string | undefined {
  const id = `graph.node.${type}`
  return HELP_COPY[id] ? id : undefined
}