package db

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/notfixingit3/echostate/internal/models"
)

const graphEdgeLimit = 3000

var graphViewEdgeTypes = map[string][]string{
	"infra": {
		"HAS_SNAPSHOT",
		"HOSTED_ON",
		"RESOLVES_TO",
		"HAS_JARM",
		"HAS_FAVICON",
		"SIGNED_BY",
	},
	"bgp": {
		"HOSTED_ON",
		"IN_PREFIX",
		"ANNOUNCES",
		"VISIBLE_ORIGIN",
		"HAS_AS_PATH",
		"AS_PATH_NEXT",
		"PATH_TO_PREFIX",
	},
	"traceroute": {
		"TRACEROUTE_HOP",
		"NEXT_HOP",
		"REACHES",
		"RESOLVES_TO",
		"SHARED_AT",
	},
	"ct": {
		"DISCOVERED_VIA_CT",
		"SCANNED_AS",
	},
	"dns": {
		"USES_NS",
		"USES_MX",
		"ALIASES_TO",
	},
	"cert": {
		"HAS_SAN",
		"SCANNED_AS",
		"SIGNED_BY",
	},
	"peering": {
		"HOSTED_ON",
		"PRESENT_AT_IX",
	},
}

// GetGraph returns nodes and edges for the requested graph view. When targetID
// is non-empty, only the subgraph connected to that target is returned.
func (c *Neo4jClient) GetGraph(ctx context.Context, targetID, view string) (*models.GraphResponse, error) {
	view = normalizeGraphView(view)

	session := c.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	relTypes := graphViewEdgeTypes[view]
	query, params := buildGraphQuery(view, targetID, relTypes)

	result, err := session.Run(ctx, query, params)
	if err != nil {
		return nil, err
	}

	records, err := result.Collect(ctx)
	if err != nil {
		return nil, err
	}

	nodeMap := make(map[string]models.GraphNode)
	edges := make([]models.GraphEdge, 0)

	for i, record := range records {
		aVal, _ := record.Get("a")
		bVal, _ := record.Get("b")
		relVal, _ := record.Get("rel")
		relPropsVal, _ := record.Get("rel_props")

		aNode, aOK := aVal.(neo4j.Node)
		bNode, bOK := bVal.(neo4j.Node)
		if !aOK || !bOK {
			continue
		}

		source, ok := addGraphNode(nodeMap, aNode)
		if !ok {
			continue
		}
		target, ok := addGraphNode(nodeMap, bNode)
		if !ok {
			continue
		}

		rel := fmt.Sprint(relVal)
		relProps := map[string]any{}
		if props, ok := relPropsVal.(map[string]any); ok {
			relProps = props
		}
		edges = append(edges, models.GraphEdge{
			ID:     fmt.Sprintf("edge-%d", i),
			Source: source,
			Target: target,
			Label:  rel,
			Color:  graphEdgeColor(rel, relProps),
			Props:  relProps,
		})
	}

	nodes := make([]models.GraphNode, 0, len(nodeMap))
	stats := make(map[string]int)
	for _, node := range nodeMap {
		nodes = append(nodes, node)
		stats[node.Type]++
	}

	response := &models.GraphResponse{
		View:  view,
		Nodes: nodes,
		Edges: edges,
		Stats: stats,
	}

	if view == "traceroute" {
		paths, err := c.getTraceroutePaths(ctx, targetID)
		if err != nil {
			return nil, err
		}
		response.Paths = paths
		response.Geo = buildGeoPointsFromPaths(paths)

		sharedHops, err := c.getSharedHops(ctx, targetID)
		if err != nil {
			return nil, err
		}
		response.SharedHops = sharedHops
		response.VantageDivergence = buildVantageDivergence(paths)
		markSharedGraphNodes(nodeMap, sharedHops)
		nodes = rebuildNodeList(nodeMap)
		stats = make(map[string]int)
		for _, node := range nodes {
			stats[node.Type]++
		}
		response.Nodes = nodes
		response.Stats = stats
	}

	if view == "peering" {
		peeringIX, err := c.getPeeringIX(ctx, targetID)
		if err != nil {
			return nil, err
		}
		response.PeeringIX = peeringIX
	}

	if view == "bgp" {
		asPaths, err := c.getBGPASPaths(ctx, targetID)
		if err != nil {
			return nil, err
		}
		response.ASPaths = asPaths
	}

	if view == "cert" {
		sharedSANs, err := c.getSharedSANs(ctx, targetID)
		if err != nil {
			return nil, err
		}
		response.SharedSANs = sharedSANs
		markSharedSANNodes(nodeMap, sharedSANs)
		nodes = rebuildNodeList(nodeMap)
		stats = make(map[string]int)
		for _, node := range nodes {
			stats[node.Type]++
		}
		response.Nodes = nodes
		response.Stats = stats
	}

	if view == "infra" {
		clusters, err := c.getInfraClusters(ctx, targetID)
		if err != nil {
			return nil, err
		}
		response.Clusters = clusters
	}

	return response, nil
}

func normalizeGraphView(view string) string {
	switch strings.ToLower(strings.TrimSpace(view)) {
	case "bgp":
		return "bgp"
	case "traceroute", "trace", "path":
		return "traceroute"
	case "ct", "subdomains", "certificate-transparency":
		return "ct"
	case "dns":
		return "dns"
	case "cert", "san", "tls":
		return "cert"
	case "peering", "ix", "peeringdb":
		return "peering"
	default:
		return "infra"
	}
}

func buildGraphQuery(view, targetID string, relTypes []string) (string, map[string]any) {
	params := map[string]any{
		"rel_types": relTypes,
		"limit":     graphEdgeLimit,
	}

	if targetID != "" {
		params["target_id"] = targetID
		query := `
			MATCH (t:Target {id: $target_id})
			OPTIONAL MATCH (t)-[*0..3]-(n)
			WITH collect(DISTINCT n) AS nodes
			UNWIND nodes AS a
			MATCH (a)-[r]->(b)
			WHERE b IN nodes AND type(r) IN $rel_types
			RETURN a, type(r) AS rel, b, properties(r) AS rel_props
			LIMIT $limit
		`
		return query, params
	}

	query := `
		MATCH (a)-[r]->(b)
		WHERE type(r) IN $rel_types
		RETURN a, type(r) AS rel, b, properties(r) AS rel_props
		LIMIT $limit
	`
	return query, params
}

func (c *Neo4jClient) getTraceroutePaths(ctx context.Context, targetID string) ([]models.GraphPath, error) {
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	var query string
	params := map[string]any{}
	if targetID != "" {
		query = `
			MATCH (t:Target {id: $target_id})-[r:TRACEROUTE_HOP]->(h:Hop)
			RETURN t.id AS target_id, t.host AS target_label, h.hop AS hop, h.ip AS ip,
			       h.rtt_ms AS rtt_ms, h.timeout AS timeout, h.country AS country, h.city AS city,
			       h.latitude AS latitude, h.longitude AS longitude, r.order AS ord,
			       coalesce(h.vantage, 'local') AS vantage, h.vantage_label AS vantage_label
			ORDER BY target_id, vantage, ord
		`
		params["target_id"] = targetID
	} else {
		query = `
			MATCH (t:Target)-[r:TRACEROUTE_HOP]->(h:Hop)
			RETURN t.id AS target_id, t.host AS target_label, h.hop AS hop, h.ip AS ip,
			       h.rtt_ms AS rtt_ms, h.timeout AS timeout, h.country AS country, h.city AS city,
			       h.latitude AS latitude, h.longitude AS longitude, r.order AS ord,
			       coalesce(h.vantage, 'local') AS vantage, h.vantage_label AS vantage_label
			ORDER BY target_id, vantage, ord
		`
	}

	result, err := session.Run(ctx, query, params)
	if err != nil {
		return nil, err
	}

	records, err := result.Collect(ctx)
	if err != nil {
		return nil, err
	}

	pathMap := make(map[string]*models.GraphPath)
	for _, record := range records {
		targetIDVal, _ := record.Get("target_id")
		targetLabelVal, _ := record.Get("target_label")
		vantageVal, _ := record.Get("vantage")
		vantageLabelVal, _ := record.Get("vantage_label")
		hopVal, _ := record.Get("hop")
		ipVal, _ := record.Get("ip")
		rttVal, _ := record.Get("rtt_ms")
		timeoutVal, _ := record.Get("timeout")
		countryVal, _ := record.Get("country")
		cityVal, _ := record.Get("city")
		latVal, _ := record.Get("latitude")
		lonVal, _ := record.Get("longitude")

		vantage := stringProp(map[string]any{"vantage": vantageVal}, "vantage")
		if vantage == "" {
			vantage = "local"
		}
		targetKey := fmt.Sprintf("%s:%s", targetIDVal, vantage)
		path, ok := pathMap[targetKey]
		if !ok {
			path = &models.GraphPath{
				TargetID:     fmt.Sprint(targetIDVal),
				TargetLabel:  fmt.Sprint(targetLabelVal),
				Vantage:      vantage,
				VantageLabel: stringProp(map[string]any{"vantage_label": vantageLabelVal}, "vantage_label"),
				Hops:         []models.GraphPathHop{},
			}
			if path.VantageLabel == "" {
				path.VantageLabel = vantage
			}
			pathMap[targetKey] = path
		}

		hopNum := intProp(map[string]any{"hop": hopVal}, "hop")
		ip := stringProp(map[string]any{"ip": ipVal}, "ip")
		timeout := boolProp(map[string]any{"timeout": timeoutVal}, "timeout")
		rtt := floatProp(map[string]any{"rtt_ms": rttVal}, "rtt_ms")

		label := fmt.Sprintf("%d. %s", hopNum, hopLabel(ip, timeout))
		pathHop := models.GraphPathHop{
			Hop:     hopNum,
			IP:      ip,
			Label:   label,
			Timeout: timeout,
		}
		if rtt > 0 {
			pathHop.RTTMs = &rtt
		}
		pathHop.Country = stringProp(map[string]any{"country": countryVal}, "country")
		pathHop.City = stringProp(map[string]any{"city": cityVal}, "city")
		if lat := floatProp(map[string]any{"latitude": latVal}, "latitude"); lat != 0 {
			pathHop.Lat = &lat
		}
		if lon := floatProp(map[string]any{"longitude": lonVal}, "longitude"); lon != 0 {
			pathHop.Lon = &lon
		}
		path.Hops = append(path.Hops, pathHop)
	}

	paths := make([]models.GraphPath, 0, len(pathMap))
	for _, path := range pathMap {
		sort.Slice(path.Hops, func(i, j int) bool {
			return path.Hops[i].Hop < path.Hops[j].Hop
		})
		paths = append(paths, *path)
	}

	sort.Slice(paths, func(i, j int) bool {
		if paths[i].TargetLabel == paths[j].TargetLabel {
			return paths[i].Vantage < paths[j].Vantage
		}
		return paths[i].TargetLabel < paths[j].TargetLabel
	})

	return paths, nil
}

func buildVantageDivergence(paths []models.GraphPath) []models.GraphVantageDivergence {
	byTarget := make(map[string][]models.GraphPath)
	for _, path := range paths {
		if len(path.Hops) == 0 {
			continue
		}
		byTarget[path.TargetID] = append(byTarget[path.TargetID], path)
	}

	var divergences []models.GraphVantageDivergence
	for _, targetPaths := range byTarget {
		if len(targetPaths) < 2 {
			continue
		}
		for i := 0; i < len(targetPaths); i++ {
			for j := i + 1; j < len(targetPaths); j++ {
				a := targetPaths[i]
				b := targetPaths[j]
				ipsA := pathHopIPs(a.Hops)
				ipsB := pathHopIPs(b.Hops)
				onlyA, onlyB, divergesAt := compareTraceroutePaths(ipsA, ipsB)
				if len(onlyA) == 0 && len(onlyB) == 0 && divergesAt == 0 {
					continue
				}
				divergences = append(divergences, models.GraphVantageDivergence{
					TargetID:    a.TargetID,
					TargetLabel: a.TargetLabel,
					VantageA:    a.VantageLabel,
					VantageB:    b.VantageLabel,
					DivergesAt:  divergesAt,
					OnlyInA:     onlyA,
					OnlyInB:     onlyB,
				})
			}
		}
	}

	sort.Slice(divergences, func(i, j int) bool {
		return divergences[i].TargetLabel < divergences[j].TargetLabel
	})
	return divergences
}

func pathHopIPs(hops []models.GraphPathHop) []string {
	var ips []string
	for _, hop := range hops {
		if hop.Timeout || hop.IP == "" {
			ips = append(ips, "*")
			continue
		}
		ips = append(ips, hop.IP)
	}
	return ips
}

func compareTraceroutePaths(a, b []string) (onlyA, onlyB []string, divergesAt int) {
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}
	for i := 0; i < minLen; i++ {
		if a[i] != b[i] {
			divergesAt = i + 1
			break
		}
	}
	if divergesAt == 0 && len(a) != len(b) {
		divergesAt = minLen + 1
	}

	setA := make(map[string]struct{})
	setB := make(map[string]struct{})
	for _, ip := range a {
		if ip != "*" {
			setA[ip] = struct{}{}
		}
	}
	for _, ip := range b {
		if ip != "*" {
			setB[ip] = struct{}{}
		}
	}
	for ip := range setA {
		if _, ok := setB[ip]; !ok {
			onlyA = append(onlyA, ip)
		}
	}
	for ip := range setB {
		if _, ok := setA[ip]; !ok {
			onlyB = append(onlyB, ip)
		}
	}
	sort.Strings(onlyA)
	sort.Strings(onlyB)
	return onlyA, onlyB, divergesAt
}

func (c *Neo4jClient) getPeeringIX(ctx context.Context, targetID string) ([]models.GraphPeeringIX, error) {
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	var query string
	params := map[string]any{}
	if targetID != "" {
		query = `
			MATCH (t:Target {id: $target_id})-[:HOSTED_ON]->(a:ASN)-[r:PRESENT_AT_IX]->(x:IX)
			RETURN t.id AS target_id, t.host AS target_label, a.number AS asn,
			       toString(x.id) AS ix_id, x.name AS ix_name, x.country AS country, x.city AS city, r.speed AS speed
			ORDER BY ix_name
		`
		params["target_id"] = targetID
	} else {
		query = `
			MATCH (t:Target)-[:HOSTED_ON]->(a:ASN)-[r:PRESENT_AT_IX]->(x:IX)
			RETURN t.id AS target_id, t.host AS target_label, a.number AS asn,
			       toString(x.id) AS ix_id, x.name AS ix_name, x.country AS country, x.city AS city, r.speed AS speed
			ORDER BY target_label, ix_name
		`
	}

	result, err := session.Run(ctx, query, params)
	if err != nil {
		return nil, err
	}

	records, err := result.Collect(ctx)
	if err != nil {
		return nil, err
	}

	var entries []models.GraphPeeringIX
	for _, record := range records {
		targetIDVal, _ := record.Get("target_id")
		targetLabelVal, _ := record.Get("target_label")
		asnVal, _ := record.Get("asn")
		ixIDVal, _ := record.Get("ix_id")
		ixNameVal, _ := record.Get("ix_name")
		countryVal, _ := record.Get("country")
		cityVal, _ := record.Get("city")
		speedVal, _ := record.Get("speed")

		entries = append(entries, models.GraphPeeringIX{
			TargetID:    fmt.Sprint(targetIDVal),
			TargetLabel: fmt.Sprint(targetLabelVal),
			ASN:         fmt.Sprint(asnVal),
			IXID:        stringProp(map[string]any{"ix_id": ixIDVal}, "ix_id"),
			IXName:      stringProp(map[string]any{"ix_name": ixNameVal}, "ix_name"),
			Country:     stringProp(map[string]any{"country": countryVal}, "country"),
			City:        stringProp(map[string]any{"city": cityVal}, "city"),
			SpeedMbps:   intProp(map[string]any{"speed": speedVal}, "speed"),
		})
	}

	return entries, nil
}

func graphEdgeColor(rel string, props map[string]any) string {
	switch rel {
	case "VISIBLE_ORIGIN", "ANNOUNCES":
		return riskColor(stringProp(props, "risk"))
	default:
		return ""
	}
}

func riskColor(risk string) string {
	switch strings.ToLower(strings.TrimSpace(risk)) {
	case "low":
		return "#22c55e"
	case "medium":
		return "#eab308"
	case "high":
		return "#ef4444"
	default:
		return "#94a3b8"
	}
}

func rebuildNodeList(nodeMap map[string]models.GraphNode) []models.GraphNode {
	nodes := make([]models.GraphNode, 0, len(nodeMap))
	for _, node := range nodeMap {
		nodes = append(nodes, node)
	}
	return nodes
}

func markSharedGraphNodes(nodeMap map[string]models.GraphNode, shared []models.GraphSharedHop) {
	if len(shared) == 0 {
		return
	}
	sharedIPs := make(map[string]int, len(shared))
	for _, hop := range shared {
		sharedIPs[hop.IP] = hop.TargetCount
	}

	for id, node := range nodeMap {
		switch node.Type {
		case "Hop":
			ip := stringProp(node.Props, "ip")
			if count, ok := sharedIPs[ip]; ok {
				if node.Props == nil {
					node.Props = map[string]any{}
				}
				node.Props["shared"] = true
				node.Props["shared_target_count"] = count
				nodeMap[id] = node
			}
		case "SharedHop":
			if node.Props == nil {
				node.Props = map[string]any{}
			}
			node.Props["shared"] = true
			if count, ok := sharedIPs[stringProp(node.Props, "ip")]; ok {
				node.Props["shared_target_count"] = count
			}
			nodeMap[id] = node
		}
	}
}

func (c *Neo4jClient) getSharedHops(ctx context.Context, targetID string) ([]models.GraphSharedHop, error) {
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	var query string
	params := map[string]any{}
	if targetID != "" {
		query = `
			MATCH (s:SharedHop)<-[:SHARED_AT]-(:Hop)<-[:TRACEROUTE_HOP]-(t:Target)
			WITH s, collect(DISTINCT t.id) AS target_ids, collect(DISTINCT t.host) AS target_labels
			WHERE size(target_ids) > 1 AND $target_id IN target_ids
			RETURN s.ip AS ip, target_ids, target_labels, size(target_ids) AS target_count
			ORDER BY target_count DESC, ip
		`
		params["target_id"] = targetID
	} else {
		query = `
			MATCH (s:SharedHop)<-[:SHARED_AT]-(:Hop)<-[:TRACEROUTE_HOP]-(t:Target)
			WITH s, collect(DISTINCT t.id) AS target_ids, collect(DISTINCT t.host) AS target_labels
			WHERE size(target_ids) > 1
			RETURN s.ip AS ip, target_ids, target_labels, size(target_ids) AS target_count
			ORDER BY target_count DESC, ip
		`
	}

	result, err := session.Run(ctx, query, params)
	if err != nil {
		return nil, err
	}

	records, err := result.Collect(ctx)
	if err != nil {
		return nil, err
	}

	var hops []models.GraphSharedHop
	for _, record := range records {
		ipVal, _ := record.Get("ip")
		targetIDsVal, _ := record.Get("target_ids")
		targetLabelsVal, _ := record.Get("target_labels")
		targetCountVal, _ := record.Get("target_count")

		ip := stringProp(map[string]any{"ip": ipVal}, "ip")
		targetIDs := stringSliceProp(targetIDsVal)
		targetLabels := stringSliceProp(targetLabelsVal)
		targetCount := intProp(map[string]any{"count": targetCountVal}, "count")
		hops = append(hops, models.GraphSharedHop{
			IP:           ip,
			TargetCount:  targetCount,
			TargetIDs:    targetIDs,
			TargetLabels: targetLabels,
		})
	}

	return hops, nil
}

func stringSliceProp(value any) []string {
	items, ok := value.([]any)
	if !ok {
		return nil
	}
	var out []string
	for _, item := range items {
		text := strings.TrimSpace(fmt.Sprint(item))
		if text != "" {
			out = append(out, text)
		}
	}
	return out
}

func buildGeoPointsFromPaths(paths []models.GraphPath) []models.GraphGeoPoint {
	var points []models.GraphGeoPoint
	seen := make(map[string]struct{})

	for _, path := range paths {
		for _, hop := range path.Hops {
			if hop.Timeout || hop.IP == "" || hop.Lat == nil || hop.Lon == nil {
				continue
			}
			key := fmt.Sprintf("%s:%d:%s", path.TargetID, hop.Hop, hop.IP)
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}

			label := hop.IP
			if hop.City != "" {
				label = hop.City + ", " + hop.Country
			} else if hop.Country != "" {
				label = hop.Country
			}

			points = append(points, models.GraphGeoPoint{
				IP:          hop.IP,
				Lat:         *hop.Lat,
				Lon:         *hop.Lon,
				Country:     hop.Country,
				City:        hop.City,
				Hop:         hop.Hop,
				TargetID:    path.TargetID,
				TargetLabel: path.TargetLabel,
				Label:       label,
			})
		}
	}

	return points
}

func (c *Neo4jClient) getBGPASPaths(ctx context.Context, targetID string) ([]models.GraphASPath, error) {
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	var query string
	params := map[string]any{}
	if targetID != "" {
		query = `
			MATCH (t:Target {id: $target_id})-[r:HAS_AS_PATH]->(start:ASN)
			OPTIONAL MATCH (t)-[:IN_PREFIX]->(p:Prefix)
			WITH t, r.path_index AS path_index, collect(DISTINCT p.cidr)[0] AS prefix, start
			OPTIONAL MATCH path = (start)-[:AS_PATH_NEXT*0..20]->(end:ASN)
			WHERE ALL(rel IN relationships(path) WHERE rel.target_id = $target_id AND rel.path_index = path_index)
			WITH t, path_index, prefix, [node IN nodes(path) | node.number] AS asn_chain
			RETURN t.id AS target_id, t.host AS target_label, prefix, path_index, asn_chain
			ORDER BY path_index
		`
		params["target_id"] = targetID
	} else {
		query = `
			MATCH (t:Target)-[r:HAS_AS_PATH]->(start:ASN)
			OPTIONAL MATCH (t)-[:IN_PREFIX]->(p:Prefix)
			WITH t, r.path_index AS path_index, collect(DISTINCT p.cidr)[0] AS prefix, start
			OPTIONAL MATCH path = (start)-[:AS_PATH_NEXT*0..20]->(end:ASN)
			WHERE ALL(rel IN relationships(path) WHERE rel.target_id = t.id AND rel.path_index = path_index)
			WITH t, path_index, prefix, [node IN nodes(path) | node.number] AS asn_chain
			RETURN t.id AS target_id, t.host AS target_label, prefix, path_index, asn_chain
			ORDER BY target_label, path_index
		`
	}

	result, err := session.Run(ctx, query, params)
	if err != nil {
		return nil, err
	}

	records, err := result.Collect(ctx)
	if err != nil {
		return nil, err
	}

	var paths []models.GraphASPath
	for _, record := range records {
		targetIDVal, _ := record.Get("target_id")
		targetLabelVal, _ := record.Get("target_label")
		prefixVal, _ := record.Get("prefix")
		chainAny, _ := record.Get("asn_chain")
		chainSlice, ok := chainAny.([]any)
		if !ok || len(chainSlice) == 0 {
			continue
		}

		var asns []string
		for _, asnAny := range chainSlice {
			asn := stringProp(map[string]any{"asn": asnAny}, "asn")
			if asn != "" {
				asns = append(asns, asn)
			}
		}
		if len(asns) == 0 {
			continue
		}

		labels := make([]string, len(asns))
		for i, asn := range asns {
			labels[i] = "AS" + asn
		}

		paths = append(paths, models.GraphASPath{
			TargetID:    fmt.Sprint(targetIDVal),
			TargetLabel: fmt.Sprint(targetLabelVal),
			Prefix:      stringProp(map[string]any{"prefix": prefixVal}, "prefix"),
			Path:        asns,
			PathLabel:   strings.Join(labels, " → "),
		})
	}

	return paths, nil
}

func hopLabel(ip string, timeout bool) string {
	if timeout || ip == "" {
		return "*"
	}
	return ip
}

func addGraphNode(nodeMap map[string]models.GraphNode, node neo4j.Node) (string, bool) {
	graphNode, ok := neo4jNodeToGraph(node)
	if !ok {
		return "", false
	}
	nodeMap[graphNode.ID] = graphNode
	return graphNode.ID, true
}

func neo4jNodeToGraph(node neo4j.Node) (models.GraphNode, bool) {
	nodeType := primaryNeo4jLabel(node.Labels)
	if nodeType == "" {
		return models.GraphNode{}, false
	}

	id := graphNodeID(nodeType, node.Props)
	if id == "" {
		return models.GraphNode{}, false
	}

	label := graphNodeLabel(nodeType, node.Props)
	props := make(map[string]any, len(node.Props))
	for key, value := range node.Props {
		props[key] = value
	}
	if nodeType == "Prefix" {
		if risk := stringProp(props, "hijack_risk"); risk != "" {
			props["risk_color"] = riskColor(risk)
		}
	}

	return models.GraphNode{
		ID:    id,
		Label: label,
		Type:  nodeType,
		Props: props,
	}, true
}

func primaryNeo4jLabel(labels []string) string {
	if len(labels) == 0 {
		return ""
	}
	return labels[0]
}

func graphNodeID(nodeType string, props map[string]any) string {
	switch nodeType {
	case "Target", "Snapshot":
		return nodeType + ":" + stringProp(props, "id")
	case "ASN":
		return "asn:" + stringProp(props, "number")
	case "IP":
		return "ip:" + stringProp(props, "address")
	case "JARM":
		return "jarm:" + stringProp(props, "hash")
	case "Favicon":
		return "favicon:" + stringProp(props, "mmh3")
	case "CertIssuer":
		return "issuer:" + stringProp(props, "name")
	case "Prefix":
		return "prefix:" + stringProp(props, "cidr")
	case "Hop":
		targetID := stringProp(props, "target_id")
		hop := intProp(props, "hop")
		vantage := stringProp(props, "vantage")
		if vantage == "" {
			vantage = "local"
		}
		if targetID == "" || hop <= 0 {
			return ""
		}
		return fmt.Sprintf("hop:%s:%s:%d", targetID, vantage, hop)
	case "IX":
		return "ix:" + stringProp(props, "id")
	case "Subdomain":
		return "subdomain:" + stringProp(props, "host")
	case "DNSHost":
		return "dnshost:" + stringProp(props, "host")
	case "SharedHop":
		return "sharedhop:" + stringProp(props, "ip")
	case "CertSAN":
		return "certsan:" + stringProp(props, "name")
	default:
		return strings.ToLower(nodeType) + ":" + stringProp(props, "id")
	}
}

func graphNodeLabel(nodeType string, props map[string]any) string {
	switch nodeType {
	case "Target":
		if host := stringProp(props, "host"); host != "" {
			return host
		}
	case "ASN":
		if number := stringProp(props, "number"); number != "" {
			return "AS" + number
		}
	case "IP":
		if address := stringProp(props, "address"); address != "" {
			return address
		}
	case "JARM":
		if hash := stringProp(props, "hash"); hash != "" {
			return truncateLabel(hash, 18)
		}
	case "Favicon":
		if mmh3 := stringProp(props, "mmh3"); mmh3 != "" {
			return "MMH3 " + mmh3
		}
	case "CertIssuer":
		if name := stringProp(props, "name"); name != "" {
			return truncateLabel(name, 28)
		}
	case "Snapshot":
		if id := stringProp(props, "id"); id != "" {
			return truncateLabel(id, 8)
		}
	case "Prefix":
		if cidr := stringProp(props, "cidr"); cidr != "" {
			risk := stringProp(props, "hijack_risk")
			if risk != "" {
				return cidr + " (" + risk + ")"
			}
			return cidr
		}
	case "SharedHop":
		if ip := stringProp(props, "ip"); ip != "" {
			return ip
		}
	case "Hop":
		hop := intProp(props, "hop")
		ip := stringProp(props, "ip")
		timeout := boolProp(props, "timeout")
		label := fmt.Sprintf("%d. %s", hop, hopLabel(ip, timeout))
		if vantageLabel := stringProp(props, "vantage_label"); vantageLabel != "" {
			return vantageLabel + " · " + label
		}
		return label
	case "IX":
		name := stringProp(props, "name")
		city := stringProp(props, "city")
		country := stringProp(props, "country")
		if name == "" {
			name = "IX"
		}
		if city != "" && country != "" {
			return truncateLabel(name+" ("+city+", "+country+")", 36)
		}
		return truncateLabel(name, 36)
	case "Subdomain", "DNSHost":
		if host := stringProp(props, "host"); host != "" {
			return truncateLabel(host, 32)
		}
	case "CertSAN":
		if name := stringProp(props, "name"); name != "" {
			return truncateLabel(name, 36)
		}
	}
	return nodeType
}

func stringProp(props map[string]any, key string) string {
	if props == nil {
		return ""
	}
	value, ok := props[key]
	if !ok || value == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(value))
}

func truncateLabel(value string, max int) string {
	if len(value) <= max {
		return value
	}
	return value[:max] + "…"
}