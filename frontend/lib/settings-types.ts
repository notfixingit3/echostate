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

export const CHANGE_TYPE_HINTS =
  "cert_expiry, bgp_origin, bgp_hijack_risk, new_ct_subdomain, dmarc_policy, web_title, graph_bgp_origin_added, graph_traceroute_hop_removed"

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
  censys_api_id?: string
  censys_api_secret?: string
  scan_concurrency?: number
  scan_job_timeout_sec?: number
  default_gatherer_timeout_sec?: number
  ct_http_timeout_sec?: number
  traceroute_timeout_sec?: number
  screenshot_timeout_sec?: number
  retention_max_snapshots?: number
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
  screenshot_timeout_sec: 25,
  retention_max_snapshots: 0,
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
