// Shared TypeScript types mirroring Go backend models.
// All JSON keys use snake_case to match the Go API responses.

export type ReportStatus = "pending" | "running" | "completed" | "failed"

export interface TargetSummary {
  id: string
  host: string
  created_at: string
  snapshot_count: number
  latest_snapshot_at?: string | null
  latest_asn?: string | null
  latest_as_name?: string | null
  latest_web_title?: string | null
}

export interface TargetDetail extends TargetSummary {
  latest_snapshot?: Snapshot | null
}

export interface SnapshotSummary {
  id: string
  target_id: string
  host: string
  scanned_at: string
  last_seen: string
  data_hash: string
  changes?: string[]
  client_ip: string
  pwhois_looked_up_at?: string | null
  pwhois_origin_as?: string | null
  pwhois_org_name?: string | null
  pwhois_country_code?: string | null
  pwhois_city?: string | null
  pwhois_prefix?: string | null
  asn?: string | null
  as_name?: string | null
  web_title?: string | null
  registrar?: string | null
  resolved_ip?: string | null
}

export interface Snapshot {
  id: string
  target_id: string
  scanned_at: string
  last_seen: string
  data_hash: string
  raw_data?: {
    host?: string
    scanned_at?: string
    whois?: Record<string, unknown>
    asn?: Record<string, unknown>
    web?: Record<string, unknown>
    errors?: string[]
  }
  changes?: string[]
  pwhois_data?: Record<string, unknown> | null
  pwhois_looked_up_at?: string | null
  pwhois_origin_as?: string | null
  pwhois_org_name?: string | null
  pwhois_country_code?: string | null
  pwhois_city?: string | null
  pwhois_prefix?: string | null
  client_ip: string
}

export interface ReportSummary {
  id: string
  snapshot_id: string
  host: string
  status: ReportStatus
  created_at: string
  completed_at?: string | null
}

export interface Report {
  id: string
  snapshot_id: string
  host: string
  status: ReportStatus
  created_at: string
  completed_at?: string | null
  download_url?: string
  error?: string
}

export interface PaginatedResponse<T> {
  data: T[]
  page: number
  limit: number
  total: number
}

export interface ScanResponse {
  snapshot_id: string
  target_id: string
  host: string
  scanned_at: string
  raw_data?: Snapshot["raw_data"]
  changes?: string[]
}

export interface CreateReportRequest {
  snapshot_id: string
}

export interface ApiErrorResponse {
  error?: string
  retry_after?: number
  message?: string
}
