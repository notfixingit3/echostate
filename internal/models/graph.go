package models

// GraphNode is a node in the infrastructure relationship graph.
type GraphNode struct {
	ID    string         `json:"id"`
	Label string         `json:"label"`
	Type  string         `json:"type"`
	Props map[string]any `json:"props,omitempty"`
}

// GraphEdge connects two graph nodes.
type GraphEdge struct {
	ID     string         `json:"id"`
	Source string         `json:"source"`
	Target string         `json:"target"`
	Label  string         `json:"label"`
	Color  string         `json:"color,omitempty"`
	Props  map[string]any `json:"props,omitempty"`
}

// GraphSharedSAN is a certificate SAN shared by multiple targets.
type GraphSharedSAN struct {
	Name         string   `json:"name"`
	TargetCount  int      `json:"target_count"`
	TargetIDs    []string `json:"target_ids"`
	TargetLabels []string `json:"target_labels"`
}

// GraphCluster groups targets sharing multiple infrastructure signals.
type GraphCluster struct {
	ID            string   `json:"id"`
	TargetCount   int      `json:"target_count"`
	TargetIDs     []string `json:"target_ids"`
	TargetLabels  []string `json:"target_labels"`
	SharedSignals []string `json:"shared_signals,omitempty"`
}

// GraphSharedHop is a traceroute hop IP shared by multiple targets.
type GraphSharedHop struct {
	IP           string   `json:"ip"`
	TargetCount  int      `json:"target_count"`
	TargetIDs    []string `json:"target_ids"`
	TargetLabels []string `json:"target_labels"`
}

// BGPRouteDiff summarizes BGP routing changes between snapshots.
type BGPRouteDiff struct {
	HijackRiskFrom        string   `json:"hijack_risk_from,omitempty"`
	HijackRiskTo          string   `json:"hijack_risk_to,omitempty"`
	RPKIFrom              string   `json:"rpki_from,omitempty"`
	RPKITo                string   `json:"rpki_to,omitempty"`
	VisibleOriginsAdded   []string `json:"visible_origins_added,omitempty"`
	VisibleOriginsRemoved []string `json:"visible_origins_removed,omitempty"`
	ASPathsAdded          []string `json:"as_paths_added,omitempty"`
	ASPathsRemoved        []string `json:"as_paths_removed,omitempty"`
	Changed               bool     `json:"changed"`
}

// TracerouteRouteDiff summarizes traceroute changes between snapshots.
type TracerouteRouteDiff struct {
	HopCountFrom int      `json:"hop_count_from,omitempty"`
	HopCountTo   int      `json:"hop_count_to,omitempty"`
	HopsAdded    []string `json:"hops_added,omitempty"`
	HopsRemoved  []string `json:"hops_removed,omitempty"`
	Changed      bool     `json:"changed"`
}

// GraphRouteDiff compares the latest snapshot to the previous one.
type GraphRouteDiff struct {
	HasPrevious        bool                 `json:"has_previous"`
	CurrentSnapshotID  string               `json:"current_snapshot_id,omitempty"`
	PreviousSnapshotID string               `json:"previous_snapshot_id,omitempty"`
	BGP                *BGPRouteDiff        `json:"bgp,omitempty"`
	Traceroute         *TracerouteRouteDiff `json:"traceroute,omitempty"`
}

// GraphPathHop is one hop in an ordered traceroute path.
type GraphPathHop struct {
	Hop      int      `json:"hop"`
	IP       string   `json:"ip,omitempty"`
	Label    string   `json:"label"`
	RTTMs    *float64 `json:"rtt_ms,omitempty"`
	Timeout  bool     `json:"timeout,omitempty"`
	Country  string   `json:"country,omitempty"`
	City     string   `json:"city,omitempty"`
	Lat      *float64 `json:"lat,omitempty"`
	Lon      *float64 `json:"lon,omitempty"`
}

// GraphGeoPoint is a geolocated point for map rendering.
type GraphGeoPoint struct {
	IP          string   `json:"ip"`
	Lat         float64  `json:"lat"`
	Lon         float64  `json:"lon"`
	Country     string   `json:"country,omitempty"`
	City        string   `json:"city,omitempty"`
	Hop         int      `json:"hop,omitempty"`
	TargetID    string   `json:"target_id,omitempty"`
	TargetLabel string   `json:"target_label,omitempty"`
	Label       string   `json:"label,omitempty"`
}

// GraphASPath is a BGP AS path observed for a target prefix.
type GraphASPath struct {
	TargetID    string   `json:"target_id"`
	TargetLabel string   `json:"target_label"`
	Prefix      string   `json:"prefix,omitempty"`
	Path        []string `json:"path"`
	PathLabel   string   `json:"path_label"`
}

// GraphPath is an ordered traceroute path for a target.
type GraphPath struct {
	TargetID     string         `json:"target_id"`
	TargetLabel  string         `json:"target_label"`
	Vantage      string         `json:"vantage,omitempty"`
	VantageLabel string         `json:"vantage_label,omitempty"`
	Hops         []GraphPathHop `json:"hops"`
}

// GraphVantageDivergence highlights asymmetric routing between traceroute vantages.
type GraphVantageDivergence struct {
	TargetID     string   `json:"target_id"`
	TargetLabel  string   `json:"target_label"`
	VantageA     string   `json:"vantage_a"`
	VantageB     string   `json:"vantage_b"`
	DivergesAt   int      `json:"diverges_at,omitempty"`
	OnlyInA      []string `json:"only_in_a,omitempty"`
	OnlyInB      []string `json:"only_in_b,omitempty"`
}

// GraphPeeringIX is an internet exchange where a target's ASN is present.
type GraphPeeringIX struct {
	TargetID    string `json:"target_id"`
	TargetLabel string `json:"target_label"`
	ASN         string `json:"asn"`
	IXID        string `json:"ix_id"`
	IXName      string `json:"ix_name"`
	Country     string `json:"country,omitempty"`
	City        string `json:"city,omitempty"`
	SpeedMbps   int    `json:"speed_mbps,omitempty"`
}

// GraphResponse is returned by GET /api/graph.
type GraphResponse struct {
	View  string         `json:"view"`
	Nodes []GraphNode    `json:"nodes"`
	Edges []GraphEdge    `json:"edges"`
	Stats map[string]int `json:"stats"`
	Paths    []GraphPath     `json:"paths,omitempty"`
	Geo      []GraphGeoPoint `json:"geo,omitempty"`
	ASPaths    []GraphASPath    `json:"as_paths,omitempty"`
	SharedHops  []GraphSharedHop  `json:"shared_hops,omitempty"`
	SharedSANs  []GraphSharedSAN  `json:"shared_sans,omitempty"`
	Clusters            []GraphCluster           `json:"clusters,omitempty"`
	RouteDiff           *GraphRouteDiff          `json:"route_diff,omitempty"`
	VantageDivergence   []GraphVantageDivergence `json:"vantage_divergence,omitempty"`
	PeeringIX           []GraphPeeringIX         `json:"peering_ix,omitempty"`
}