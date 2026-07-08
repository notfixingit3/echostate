package db

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/notfixingit3/echostate/internal/models"
)

var infraClusterEdgeTypes = []string{
	"HOSTED_ON",
	"HAS_JARM",
	"HAS_JA3S",
	"HAS_FAVICON",
	"SIGNED_BY",
	"HAS_SAN",
	"RESOLVES_TO",
}

type clusterPair struct {
	ID1    string
	Host1  string
	ID2    string
	Host2  string
	Shared int
}

type unionFind struct {
	parent map[string]string
}

func newUnionFind(ids []string) *unionFind {
	parent := make(map[string]string, len(ids))
	for _, id := range ids {
		parent[id] = id
	}
	return &unionFind{parent: parent}
}

func (uf *unionFind) find(id string) string {
	for uf.parent[id] != id {
		uf.parent[id] = uf.parent[uf.parent[id]]
		id = uf.parent[id]
	}
	return id
}

func (uf *unionFind) union(a, b string) {
	rootA := uf.find(a)
	rootB := uf.find(b)
	if rootA != rootB {
		uf.parent[rootB] = rootA
	}
}

func (c *Neo4jClient) getInfraClusters(ctx context.Context, targetID string) ([]models.GraphCluster, error) {
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	query := `
		MATCH (t1:Target)-[r1]->(n)<-[r2]-(t2:Target)
		WHERE t1.id < t2.id
		  AND type(r1) IN $types
		  AND type(r1) = type(r2)
		  AND ($target_id = '' OR t1.id = $target_id OR t2.id = $target_id)
		WITH t1, t2, count(DISTINCT n) AS shared_count
		WHERE shared_count >= 2
		RETURN t1.id AS id1, t1.host AS host1, t2.id AS id2, t2.host AS host2, shared_count
		ORDER BY shared_count DESC
	`

	result, err := session.Run(ctx, query, map[string]any{
		"types":     infraClusterEdgeTypes,
		"target_id": targetID,
	})
	if err != nil {
		return nil, err
	}

	records, err := result.Collect(ctx)
	if err != nil {
		return nil, err
	}

	var pairs []clusterPair
	idSet := make(map[string]struct{})
	hostByID := make(map[string]string)

	for _, record := range records {
		pair := clusterPair{
			ID1:    fmt.Sprint(record.Values[0]),
			Host1:  fmt.Sprint(record.Values[1]),
			ID2:    fmt.Sprint(record.Values[2]),
			Host2:  fmt.Sprint(record.Values[3]),
			Shared: intProp(map[string]any{"shared": record.Values[4]}, "shared"),
		}
		if pair.ID1 == "" || pair.ID2 == "" {
			continue
		}
		pairs = append(pairs, pair)
		idSet[pair.ID1] = struct{}{}
		idSet[pair.ID2] = struct{}{}
		hostByID[pair.ID1] = pair.Host1
		hostByID[pair.ID2] = pair.Host2
	}

	if len(pairs) == 0 {
		return nil, nil
	}

	ids := make([]string, 0, len(idSet))
	for id := range idSet {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	uf := newUnionFind(ids)
	for _, pair := range pairs {
		uf.union(pair.ID1, pair.ID2)
	}

	groups := make(map[string][]string)
	for _, id := range ids {
		root := uf.find(id)
		groups[root] = append(groups[root], id)
	}

	var clusters []models.GraphCluster
	for root, members := range groups {
		if len(members) < 2 {
			continue
		}
		sort.Strings(members)

		labels := make([]string, len(members))
		for i, id := range members {
			labels[i] = hostByID[id]
		}
		sort.Strings(labels)

		signals, err := c.clusterSharedSignals(ctx, members)
		if err != nil {
			return nil, err
		}

		clusters = append(clusters, models.GraphCluster{
			ID:            "cluster:" + root,
			TargetCount:   len(members),
			TargetIDs:     members,
			TargetLabels:  labels,
			SharedSignals: signals,
		})
	}

	sort.Slice(clusters, func(i, j int) bool {
		if clusters[i].TargetCount != clusters[j].TargetCount {
			return clusters[i].TargetCount > clusters[j].TargetCount
		}
		return clusters[i].ID < clusters[j].ID
	})

	return clusters, nil
}

func (c *Neo4jClient) clusterSharedSignals(ctx context.Context, targetIDs []string) ([]string, error) {
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	query := `
		MATCH (t:Target)-[r]->(n)
		WHERE t.id IN $target_ids AND type(r) IN $types
		WITH n, labels(n)[0] AS node_type, type(r) AS rel_type, count(DISTINCT t) AS target_count
		WHERE target_count = size($target_ids)
		RETURN node_type, rel_type, n
		ORDER BY rel_type, node_type
		LIMIT 12
	`

	result, err := session.Run(ctx, query, map[string]any{
		"target_ids": targetIDs,
		"types":      infraClusterEdgeTypes,
	})
	if err != nil {
		return nil, err
	}

	records, err := result.Collect(ctx)
	if err != nil {
		return nil, err
	}

	var signals []string
	for _, record := range records {
		nodeType := fmt.Sprint(record.Values[0])
		relType := fmt.Sprint(record.Values[1])
		nodeVal := record.Values[2]
		node, ok := nodeVal.(neo4j.Node)
		if !ok {
			continue
		}
		label := graphNodeLabel(nodeType, node.Props)
		signals = append(signals, fmt.Sprintf("%s (%s)", label, relType))
	}
	return signals, nil
}

func (c *Neo4jClient) getSharedSANs(ctx context.Context, targetID string) ([]models.GraphSharedSAN, error) {
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	var query string
	params := map[string]any{}
	if targetID != "" {
		query = `
			MATCH (s:CertSAN)<-[:HAS_SAN]-(t:Target)
			WITH s, collect(DISTINCT t.id) AS target_ids, collect(DISTINCT t.host) AS target_labels
			WHERE size(target_ids) > 1 AND $target_id IN target_ids
			RETURN s.name AS name, target_ids, target_labels, size(target_ids) AS target_count
			ORDER BY target_count DESC, name
		`
		params["target_id"] = targetID
	} else {
		query = `
			MATCH (s:CertSAN)<-[:HAS_SAN]-(t:Target)
			WITH s, collect(DISTINCT t.id) AS target_ids, collect(DISTINCT t.host) AS target_labels
			WHERE size(target_ids) > 1
			RETURN s.name AS name, target_ids, target_labels, size(target_ids) AS target_count
			ORDER BY target_count DESC, name
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

	var sans []models.GraphSharedSAN
	for _, record := range records {
		nameVal, _ := record.Get("name")
		targetIDsVal, _ := record.Get("target_ids")
		targetLabelsVal, _ := record.Get("target_labels")
		targetCountVal, _ := record.Get("target_count")

		name := stringProp(map[string]any{"name": nameVal}, "name")
		sans = append(sans, models.GraphSharedSAN{
			Name:         name,
			TargetCount:  intProp(map[string]any{"count": targetCountVal}, "count"),
			TargetIDs:    stringSliceProp(targetIDsVal),
			TargetLabels: stringSliceProp(targetLabelsVal),
		})
	}

	return sans, nil
}

func markSharedSANNodes(nodeMap map[string]models.GraphNode, shared []models.GraphSharedSAN) {
	if len(shared) == 0 {
		return
	}
	sharedNames := make(map[string]int, len(shared))
	for _, san := range shared {
		sharedNames[strings.ToLower(san.Name)] = san.TargetCount
	}

	for id, node := range nodeMap {
		if node.Type != "CertSAN" {
			continue
		}
		name := strings.ToLower(stringProp(node.Props, "name"))
		if count, ok := sharedNames[name]; ok {
			if node.Props == nil {
				node.Props = map[string]any{}
			}
			node.Props["shared"] = true
			node.Props["shared_target_count"] = count
			nodeMap[id] = node
		}
	}
}
