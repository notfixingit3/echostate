// Shared TypeScript types mirroring Go backend models.
// All JSON keys use snake_case to match the Go API responses.

export type ReportStatus = "pending" | "running" | "completed" | "failed"

export interface ChangeDetail {
  type: string
  severity: "info" | "warning" | "critical" | string
  summary: string
  field?: string
  detail?: string
}

export interface TargetSummary {
  id: string
  host: string
  tags?: string[]
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

export interface TargetCollectionSummary {
  id: string
  name: string
  description: string
  created_at: string
  updated_at: string
  target_count: number
}

export interface TargetCollectionDetail extends TargetCollectionSummary {
  targets: TargetSummary[]
}

export interface CollectionRescanJob {
  target_id: string
  host: string
  job_id: string
}

export interface CollectionRescanResponse {
  collection_id: string
  jobs: CollectionRescanJob[]
}

export type NoteStatus = "active" | "archived" | "trashed"

export interface NoteReference {
  type: "url" | "target" | "snapshot" | "collection" | "graph_node" | string
  label?: string
  url?: string
  id?: string
}

export interface GraphPinnedNode {
  x: number
  y: number
}

export interface SavedGraphView {
  id: string
  name: string
  description: string
  view_mode: GraphViewMode | string
  target_id?: string | null
  vantage_filter: string
  snapshot_id?: string | null
  compare_snapshot_id?: string | null
  compare_mode: GraphCompareMode | string
  pinned_nodes: Record<string, GraphPinnedNode>
  selected_node_id?: string | null
  created_at: string
  updated_at: string
  target_host?: string | null
}

export type GraphViewMode =
  | "infra"
  | "bgp"
  | "traceroute"
  | "peering"
  | "ct"
  | "dns"
  | "cert"

export type GraphCompareMode = "previous" | "latest"

export interface InvestigationNote {
  id: string
  title: string
  body: string
  status: NoteStatus
  target_id?: string | null
  snapshot_id?: string | null
  collection_id?: string | null
  graph_node_id?: string | null
  graph_node_label?: string | null
  graph_node_type?: string | null
  references: NoteReference[]
  trashed_at?: string | null
  archived_at?: string | null
  created_at: string
  updated_at: string
  target_host?: string | null
  collection_name?: string | null
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
    tls?: Record<string, unknown>
    dns?: Record<string, unknown>
    errors?: string[]
  }
  changes?: string[]
  change_details?: ChangeDetail[]
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

export interface GraphNode {
  id: string
  label: string
  type: string
  props?: Record<string, unknown>
}

export interface GraphEdge {
  id: string
  source: string
  target: string
  label: string
  color?: string
  props?: Record<string, unknown>
}

export interface GraphPathHop {
  hop: number
  ip?: string
  label: string
  rtt_ms?: number
  timeout?: boolean
  country?: string
  city?: string
  lat?: number
  lon?: number
}

export interface GraphGeoPoint {
  ip: string
  lat: number
  lon: number
  country?: string
  city?: string
  hop?: number
  target_id?: string
  target_label?: string
  label?: string
}

export interface GraphASPath {
  target_id: string
  target_label: string
  prefix?: string
  path: string[]
  path_label: string
}

export interface GraphPath {
  target_id: string
  target_label: string
  vantage?: string
  vantage_label?: string
  hops: GraphPathHop[]
}

export interface GraphVantageDivergence {
  target_id: string
  target_label: string
  vantage_a: string
  vantage_b: string
  diverges_at?: number
  only_in_a?: string[]
  only_in_b?: string[]
}

export interface GraphPeeringIX {
  target_id: string
  target_label: string
  asn: string
  ix_id: string
  ix_name: string
  country?: string
  city?: string
  speed_mbps?: number
}

export interface ScreenshotEntry {
  snapshot_id: string
  scanned_at: string
  url?: string
  width?: number
  height?: number
  format?: string
  thumbnail?: string
  error?: string
}

export interface GraphIntelEvent {
  id: string
  type: string
  severity?: string
  summary: string
  field?: string
  detail?: string
  snapshot_id?: string
  detected_at?: number
  source?: string
}

export interface GraphTopologyDiff {
  added_node_ids?: string[]
  removed_node_ids?: string[]
  added_edge_ids?: string[]
  removed_edge_ids?: string[]
  changed?: boolean
}

export interface GraphResponse {
  view: string
  nodes: GraphNode[]
  edges: GraphEdge[]
  stats: Record<string, number>
  snapshot_id?: string
  scanned_at?: number
  compare_snapshot_id?: string
  compare_mode?: "previous" | "latest"
  topology_diff?: GraphTopologyDiff
  events?: GraphIntelEvent[]
  paths?: GraphPath[]
  geo?: GraphGeoPoint[]
  as_paths?: GraphASPath[]
  shared_hops?: GraphSharedHop[]
  shared_sans?: GraphSharedSAN[]
  clusters?: GraphCluster[]
  route_diff?: GraphRouteDiff
  vantage_divergence?: GraphVantageDivergence[]
  peering_ix?: GraphPeeringIX[]
}

export interface GraphSharedSAN {
  name: string
  target_count: number
  target_ids: string[]
  target_labels: string[]
}

export interface GraphCluster {
  id: string
  target_count: number
  target_ids: string[]
  target_labels: string[]
  shared_signals?: string[]
}

export interface GraphSharedHop {
  ip: string
  target_count: number
  target_ids: string[]
  target_labels: string[]
}

export interface BGPRouteDiff {
  hijack_risk_from?: string
  hijack_risk_to?: string
  rpki_from?: string
  rpki_to?: string
  visible_origins_added?: string[]
  visible_origins_removed?: string[]
  as_paths_added?: string[]
  as_paths_removed?: string[]
  changed: boolean
}

export interface TracerouteRouteDiff {
  hop_count_from?: number
  hop_count_to?: number
  hops_added?: string[]
  hops_removed?: string[]
  changed: boolean
}

export interface GraphRouteDiff {
  has_previous: boolean
  current_snapshot_id?: string
  previous_snapshot_id?: string
  bgp?: BGPRouteDiff
  traceroute?: TracerouteRouteDiff
}


