export type Webhook = {
  id: string
  name: string
  type: string
  url: string
  config?: Record<string, unknown>
  enabled: boolean
}

export const GRAPH_DRIFT_TYPES = [
  "graph_bgp_origin_added",
  "graph_bgp_origin_removed",
  "graph_as_path_added",
  "graph_as_path_removed",
  "graph_rpki_changed",
  "graph_traceroute_hop_added",
  "graph_traceroute_hop_removed",
  "graph_traceroute_reordered",
]

export const MAIL_SECURITY_CHANGE_TYPES = [
  "caa_added",
  "caa_removed",
  "dnssec_status",
  "mta_sts_added",
  "mta_sts_mode",
  "tls_rpt_added",
  "tls_rpt_changed",
  "bimi_added",
  "dmarc_policy",
  "mail_posture",
  "security_txt_added",
  "security_contact_added",
  "csp_changed",
  "hsts_changed",
  "hsts_preload_status",
  "tls_ocsp_stapling",
  "new_ct_certificate",
  "redirect_chain_changed",
  "enrichment_hibp_breach",
]

export const CHANGE_TYPE_HINTS =
  "cert_expiry, bgp_origin, new_ct_subdomain, caa_added, dnssec_status, mta_sts_mode, security_contact_added, redirect_chain_changed, new_ct_certificate, enrichment_hibp_breach, graph_traceroute_hop_removed"

export type AlertRule = {
  id: string
  name: string
  enabled: boolean
  min_severity: string
  match_types: string[]
  webhook_ids: string[]
}

export type SystemSettings = {
  dns_servers: string
  dkim_selectors?: string
  pwhois_server: string
  rate_limit: number
  api_key?: string
  shodan_api_key?: string
  hibp_api_key?: string
  riskiq_api_user?: string
  riskiq_api_key?: string
  virustotal_api_key?: string
  censys_api_id?: string
  censys_api_secret?: string
  scan_concurrency?: number
  scan_job_timeout_sec?: number
  default_gatherer_timeout_sec?: number
  ct_http_timeout_sec?: number
  traceroute_timeout_sec?: number
  traceroute_redact_scanner_prefix?: boolean
  traceroute_redact_local_extra_hops?: number
  screenshot_timeout_sec?: number
  scanner_http_timeout_sec?: number
  enrichment_http_timeout_sec?: number
  wayback_http_timeout_sec?: number
  retention_max_snapshots?: number
  audit_retention_days?: number
  schedule_enabled?: boolean
  schedule_interval_minutes?: number
  schedule_stale_hours?: number
  schedule_tags?: string[]
  alert_rules?: AlertRule[]
  auth_enabled?: boolean
  enrollment_code_ttl_hours?: number
  enrollment_code_length?: number
  recovery_code_length?: number
  session_ttl_hours?: number
  max_code_attempts?: number
  code_attempt_window_minutes?: number
  webauthn_rp_id?: string
  webauthn_rp_origin?: string
}

export const defaultSystemSettings = (): SystemSettings => ({
  dns_servers: "8.8.8.8,1.1.1.1",
  pwhois_server: "whois.pwhois.org",
  rate_limit: 30,
  scan_concurrency: 2,
  scan_job_timeout_sec: 120,
  default_gatherer_timeout_sec: 20,
  ct_http_timeout_sec: 60,
  traceroute_timeout_sec: 40,
  traceroute_redact_scanner_prefix: true,
  traceroute_redact_local_extra_hops: 0,
  screenshot_timeout_sec: 25,
  scanner_http_timeout_sec: 8,
  enrichment_http_timeout_sec: 12,
  wayback_http_timeout_sec: 30,
  retention_max_snapshots: 0,
  audit_retention_days: 90,
  schedule_enabled: false,
  schedule_interval_minutes: 60,
  schedule_stale_hours: 24,
  schedule_tags: [],
  alert_rules: [],
  auth_enabled: true,
  enrollment_code_ttl_hours: 24,
  enrollment_code_length: 8,
  recovery_code_length: 12,
  session_ttl_hours: 168,
  max_code_attempts: 5,
  code_attempt_window_minutes: 15,
  webauthn_rp_id: "",
  webauthn_rp_origin: "",
})
