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
  "graph.rescan_hint":
    "Temporal graph edges are versioned per snapshot. Rescan to accumulate history for the slider and compare modes.",

  "page.collections":
    "Watchlists of related targets. Add members and queue bulk rescans from one place.",
  "collections.create": "Name and optional description for a new target collection.",
  "collections.list": "Select a collection to add or remove targets and queue bulk rescans.",
  "collections.rescan":
    "Enqueue an async scan job for every target in the selected collection.",

  "intel.cert_expires": "TLS certificate not-after date from the leaf cert.",
  "intel.cert_issuer": "Certificate authority that signed the active TLS certificate.",
  "intel.dns_a": "IPv4/IPv6 addresses returned by DNS resolution at scan time.",
  "intel.dns_soa": "Authoritative SOA zone and serial for the target.",
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
  "settings.pwhois_server": "pWhois server hostname for async client-IP ASN and org enrichment.",
  "settings.rate_limit": "Maximum API requests per minute per client IP.",
  "settings.scan_concurrency": "Maximum scan jobs processed in parallel by the API worker pool.",
  "settings.scan_timeouts":
    "Per-gatherer timeouts in seconds. Increase for slow CT or traceroute targets.",
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
  "settings.alert_rules":
    "Filter which change_details events trigger webhook notifications.",
  "settings.alert_match_types":
    "Comma-separated change types. Leave empty to match all severities above the minimum.",
  "settings.alert_webhooks":
    "Limit delivery to specific integrations. Empty means all enabled webhooks.",

  "reports.status": "queued, running, completed, or failed PDF generation state.",
  "reports.snapshot": "Snapshot used as the source data for this report.",
  "reports.download": "Download the generated PDF when status is completed.",

  "snapshot.changes":
    "Structured field-level diffs with severity, type, and summary versus the previous snapshot.",
  "snapshot.raw_data": "Full JSON payload returned by all gatherers for this scan.",
}

export function getHelpCopy(id: string): string | undefined {
  return HELP_COPY[id]
}