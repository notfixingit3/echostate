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
  "settings.profile": "Your display name and role in EchoState.",
  "settings.preferences": "Default theme and timezone for your account across devices.",
  "settings.passkeys": "Issue device codes, register passkeys, rename, or remove credentials.",
  "settings.device_code":
    "Single-use enrollment code for signing in on another browser or phone, then registering a passkey there.",
  "page.reports":
    "PDF intelligence reports generated from snapshot data. Reports queue asynchronously after you request them.",

  "scan.host":
    "Hostname, IP address, or full URL. EchoState normalizes input and gathers WHOIS, ASN, DNS, TLS, and web intel concurrently.",
  "scan.submit":
    "Queues an async scan job. Poll until complete, then open the target or snapshot detail page.",
  "scan.status":
    "Shows queued, running, completed, or failed state for the current scan job.",

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
  "intel.favicon_mmh3": "Shodan-compatible MurmurHash3 of the favicon.",
  "intel.jarm": "JARM TLS fingerprint hash from port 443.",
  "intel.ja3s": "JA3S TLS server fingerprint from the port 443 ServerHello.",
  "intel.hijack_risk": "Heuristic BGP hijack risk from RIPEstat visible-origin analysis.",
  "intel.bgp_path_stability":
    "BGP AS-path diversity: stable (one path), diverse (2–4), or volatile (5+).",
  "intel.mail_posture":
    "Composite SPF, DKIM, and DMARC grade (A–F) for outbound email security posture.",
  "intel.js_assets":
    "Passive script and stylesheet URLs with library/CDN hints from page assets.",
  "intel.wp_plugins": "WordPress plugins detected passively from page assets.",
  "intel.wp_themes": "Active WordPress theme slug and version when available.",

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
    "Per-gatherer timeouts in seconds. Increase for slow CT or traceroute targets.",
  "settings.traceroute_redact_scanner_prefix":
    "Strip private LAN hops and the first public ISP hop from local traceroutes before storing results. External vantage paths are unchanged.",
  "settings.traceroute_redact_local_extra_hops":
    "Additional local hops to hide after the automatic scanner-prefix redaction (0 = none).",
  "settings.retention":
    "Per-target snapshot keep-N policy. Older snapshots are pruned after each scan.",
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
  "scan.and_report":
    "Run a scan and queue a PDF report in one step. The result card polls until the report is ready, then offers download.",

  "snapshot.changes":
    "Structured field-level diffs with severity, type, and summary versus the previous snapshot.",
  "snapshot.raw_data": "Full JSON payload returned by all gatherers for this scan.",
}

export function getHelpCopy(id: string): string | undefined {
  return HELP_COPY[id]
}