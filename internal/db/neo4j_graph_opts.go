package db

import (
	"context"
	"strings"

	"github.com/notfixingit3/echostate/internal/models"
)

type GraphQueryOpts struct {
	TargetID          string
	View              string
	SnapshotID        string
	CompareSnapshotID string
	CompareMode       string
}

func GraphQueryOptsFromRequest(targetID, view, snapshotID, compareSnapshotID, compareMode string) GraphQueryOpts {
	return GraphQueryOpts{
		TargetID:          strings.TrimSpace(targetID),
		View:              strings.TrimSpace(view),
		SnapshotID:        strings.TrimSpace(snapshotID),
		CompareSnapshotID: strings.TrimSpace(compareSnapshotID),
		CompareMode:       strings.TrimSpace(compareMode),
	}
}

func (c *Neo4jClient) GetGraphWithOpts(ctx context.Context, opts GraphQueryOpts) (*models.GraphResponse, error) {
	primary, err := c.fetchGraph(ctx, opts, opts.SnapshotID)
	if err != nil {
		return nil, err
	}

	primary.SnapshotID = opts.SnapshotID
	primary.CompareSnapshotID = opts.CompareSnapshotID
	primary.CompareMode = opts.CompareMode
	if opts.SnapshotID != "" {
		for _, edge := range primary.Edges {
			if scanned := int64(intProp(edge.Props, "scanned_at")); scanned > primary.ScannedAt {
				primary.ScannedAt = scanned
			}
		}
	}

	if opts.CompareSnapshotID != "" && opts.SnapshotID != "" {
		baseline, err := c.fetchGraph(ctx, opts, opts.CompareSnapshotID)
		if err == nil && baseline != nil {
			primary.TopologyDiff = diffTopology(primary, baseline)
		}
	}

	return primary, nil
}

func diffTopology(current, baseline *models.GraphResponse) *models.GraphTopologyDiff {
	if current == nil || baseline == nil {
		return nil
	}

	curNodes := make(map[string]struct{}, len(current.Nodes))
	baseNodes := make(map[string]struct{}, len(baseline.Nodes))
	curEdges := make(map[string]struct{}, len(current.Edges))
	baseEdges := make(map[string]struct{}, len(baseline.Edges))

	for _, node := range current.Nodes {
		curNodes[node.ID] = struct{}{}
	}
	for _, node := range baseline.Nodes {
		baseNodes[node.ID] = struct{}{}
	}
	for _, edge := range current.Edges {
		curEdges[edge.ID] = struct{}{}
	}
	for _, edge := range baseline.Edges {
		baseEdges[edge.ID] = struct{}{}
	}

	diff := &models.GraphTopologyDiff{}
	for id := range curNodes {
		if _, ok := baseNodes[id]; !ok {
			diff.AddedNodeIDs = append(diff.AddedNodeIDs, id)
		}
	}
	for id := range baseNodes {
		if _, ok := curNodes[id]; !ok {
			diff.RemovedNodeIDs = append(diff.RemovedNodeIDs, id)
		}
	}
	for id := range curEdges {
		if _, ok := baseEdges[id]; !ok {
			diff.AddedEdgeIDs = append(diff.AddedEdgeIDs, id)
		}
	}
	for id := range baseEdges {
		if _, ok := curEdges[id]; !ok {
			diff.RemovedEdgeIDs = append(diff.RemovedEdgeIDs, id)
		}
	}

	diff.Changed = len(diff.AddedNodeIDs) > 0 ||
		len(diff.RemovedNodeIDs) > 0 ||
		len(diff.AddedEdgeIDs) > 0 ||
		len(diff.RemovedEdgeIDs) > 0

	return diff
}