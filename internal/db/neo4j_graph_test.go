package db

import (
	"testing"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/notfixingit3/echostate/internal/models"
)

func TestNeo4jNodeToGraph(t *testing.T) {
	node, ok := neo4jNodeToGraph(neo4j.Node{
		Labels: []string{"Target"},
		Props: map[string]any{
			"id":   "abc-123",
			"host": "example.com",
		},
	})
	requireOK(t, ok)
	if node.ID != "Target:abc-123" || node.Label != "example.com" || node.Type != "Target" {
		t.Fatalf("unexpected target node: %#v", node)
	}

	jarm, ok := neo4jNodeToGraph(neo4j.Node{
		Labels: []string{"JARM"},
		Props:  map[string]any{"hash": "27d40d40d29d40d1dc28d1d000a1d0a1d2d2d0002d1d2d2d2d2d2d2d2d2d2d2d2"},
	})
	requireOK(t, ok)
	if jarm.Type != "JARM" || jarm.ID != "jarm:27d40d40d29d40d1dc28d1d000a1d0a1d2d2d0002d1d2d2d2d2d2d2d2d2d2d2d2" {
		t.Fatalf("unexpected jarm node: %#v", jarm)
	}
}

func TestGraphNodeLabel(t *testing.T) {
	got := graphNodeLabel("ASN", map[string]any{"number": "15169"})
	if got != "AS15169" {
		t.Fatalf("graphNodeLabel() = %q, want AS15169", got)
	}
}

func requireOK(t *testing.T, ok bool) {
	t.Helper()
	if !ok {
		t.Fatal("expected ok")
	}
}

func TestAddGraphNodeDedupes(t *testing.T) {
	nodeMap := make(map[string]models.GraphNode)
	node := neo4j.Node{
		Labels: []string{"IP"},
		Props:  map[string]any{"address": "93.184.216.34"},
	}

	id1, ok := addGraphNode(nodeMap, node)
	requireOK(t, ok)
	id2, ok := addGraphNode(nodeMap, node)
	requireOK(t, ok)

	if id1 != id2 || len(nodeMap) != 1 {
		t.Fatalf("expected single deduped node, got map=%#v", nodeMap)
	}
}

func TestNeo4jNodeToGraphEmpty(t *testing.T) {
	_, ok := neo4jNodeToGraph(neo4j.Node{})
	if ok {
		t.Fatal("expected false for empty node")
	}
}

func TestNeo4jNodeToGraphPrefixAndHop(t *testing.T) {
	prefix, ok := neo4jNodeToGraph(neo4j.Node{
		Labels: []string{"Prefix"},
		Props:  map[string]any{"cidr": "93.184.216.0/24"},
	})
	requireOK(t, ok)
	if prefix.ID != "prefix:93.184.216.0/24" || prefix.Label != "93.184.216.0/24" {
		t.Fatalf("unexpected prefix node: %#v", prefix)
	}

	hop, ok := neo4jNodeToGraph(neo4j.Node{
		Labels: []string{"Hop"},
		Props: map[string]any{
			"target_id": "target-1",
			"hop":       3,
			"ip":        "10.0.0.1",
		},
	})
	requireOK(t, ok)
	if hop.ID != "hop:target-1:local:3" || hop.Label != "3. 10.0.0.1" {
		t.Fatalf("unexpected hop node: %#v", hop)
	}
}

func TestNormalizeGraphView(t *testing.T) {
	if normalizeGraphView("BGP") != "bgp" {
		t.Fatalf("expected bgp view")
	}
	if normalizeGraphView("path") != "traceroute" {
		t.Fatalf("expected traceroute view")
	}
	if normalizeGraphView("ct") != "ct" {
		t.Fatalf("expected ct view")
	}
	if normalizeGraphView("dns") != "dns" {
		t.Fatalf("expected dns view")
	}
	if normalizeGraphView("san") != "cert" {
		t.Fatalf("expected cert view")
	}
	if normalizeGraphView("") != "infra" {
		t.Fatalf("expected infra view")
	}
}

func TestNeo4jNodeToGraphCertSAN(t *testing.T) {
	san, ok := neo4jNodeToGraph(neo4j.Node{
		Labels: []string{"CertSAN"},
		Props:  map[string]any{"name": "api.example.com"},
	})
	requireOK(t, ok)
	if san.ID != "certsan:api.example.com" {
		t.Fatalf("unexpected cert san node: %#v", san)
	}
}

func TestNeo4jNodeToGraphSubdomainAndDNSHost(t *testing.T) {
	sub, ok := neo4jNodeToGraph(neo4j.Node{
		Labels: []string{"Subdomain"},
		Props:  map[string]any{"host": "api.example.com"},
	})
	requireOK(t, ok)
	if sub.ID != "subdomain:api.example.com" {
		t.Fatalf("unexpected subdomain node: %#v", sub)
	}

	dns, ok := neo4jNodeToGraph(neo4j.Node{
		Labels: []string{"DNSHost"},
		Props:  map[string]any{"host": "ns1.example.com"},
	})
	requireOK(t, ok)
	if dns.ID != "dnshost:ns1.example.com" {
		t.Fatalf("unexpected dns host node: %#v", dns)
	}
}
