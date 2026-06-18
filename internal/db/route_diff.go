package db

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/notfixingit3/echostate/internal/models"
)

func (d *DB) GetRouteDiffForTarget(ctx context.Context, targetID uuid.UUID) (*models.GraphRouteDiff, error) {
	rows, err := d.Pool.Query(ctx, `
		SELECT id, raw_data
		FROM snapshots
		WHERE target_id = $1
		ORDER BY scanned_at DESC
		LIMIT 2
	`, targetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snapshots []struct {
		id      uuid.UUID
		rawData map[string]any
	}
	for rows.Next() {
		var snap struct {
			id      uuid.UUID
			rawData map[string]any
		}
		if err := rows.Scan(&snap.id, &snap.rawData); err != nil {
			return nil, err
		}
		snapshots = append(snapshots, snap)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(snapshots) == 0 {
		return nil, pgx.ErrNoRows
	}

	diff := &models.GraphRouteDiff{
		HasPrevious:       len(snapshots) > 1,
		CurrentSnapshotID: snapshots[0].id.String(),
	}
	if len(snapshots) > 1 {
		diff.PreviousSnapshotID = snapshots[1].id.String()
		diff.BGP = compareBGPRoutes(snapshots[1].rawData, snapshots[0].rawData)
		diff.Traceroute = compareTracerouteRoutes(snapshots[1].rawData, snapshots[0].rawData)
	}

	return diff, nil
}

func compareBGPRoutes(previous, current map[string]any) *models.BGPRouteDiff {
	prevRouting := routingMap(previous)
	currRouting := routingMap(current)

	diff := &models.BGPRouteDiff{
		HijackRiskFrom: stringProp(prevRouting, "hijack_risk"),
		HijackRiskTo:   stringProp(currRouting, "hijack_risk"),
		RPKIFrom:       overallRPKIStatus(prevRouting),
		RPKITo:         overallRPKIStatus(currRouting),
	}

	prevOrigins := visibleOrigins(prevRouting)
	currOrigins := visibleOrigins(currRouting)
	diff.VisibleOriginsAdded, diff.VisibleOriginsRemoved = stringDiff(prevOrigins, currOrigins)

	prevPaths := asPaths(prevRouting)
	currPaths := asPaths(currRouting)
	diff.ASPathsAdded, diff.ASPathsRemoved = stringDiff(prevPaths, currPaths)

	diff.Changed = diff.HijackRiskFrom != diff.HijackRiskTo ||
		diff.RPKIFrom != diff.RPKITo ||
		len(diff.VisibleOriginsAdded) > 0 ||
		len(diff.VisibleOriginsRemoved) > 0 ||
		len(diff.ASPathsAdded) > 0 ||
		len(diff.ASPathsRemoved) > 0

	return diff
}

func compareTracerouteRoutes(previous, current map[string]any) *models.TracerouteRouteDiff {
	prevHops := tracerouteIPs(previous)
	currHops := tracerouteIPs(current)

	diff := &models.TracerouteRouteDiff{
		HopCountFrom: len(prevHops),
		HopCountTo:   len(currHops),
	}
	diff.HopsAdded, diff.HopsRemoved = stringDiff(prevHops, currHops)
	diff.Changed = diff.HopCountFrom != diff.HopCountTo ||
		len(diff.HopsAdded) > 0 ||
		len(diff.HopsRemoved) > 0

	return diff
}

func routingMap(raw map[string]any) map[string]any {
	asn, ok := raw["asn"].(map[string]any)
	if !ok {
		return nil
	}
	routing, ok := asn["routing"].(map[string]any)
	if !ok {
		return nil
	}
	return routing
}

func visibleOrigins(routing map[string]any) []string {
	originsAny, ok := routing["visible_origins"].([]any)
	if !ok {
		return nil
	}
	var origins []string
	for _, item := range originsAny {
		asn := normalizeASNValue(item)
		if asn != "" {
			origins = append(origins, asn)
		}
	}
	sort.Strings(origins)
	return origins
}

func asPaths(routing map[string]any) []string {
	pathsAny, ok := routing["as_paths"].([]any)
	if !ok {
		return nil
	}
	var paths []string
	for _, item := range pathsAny {
		path := strings.TrimSpace(fmt.Sprint(item))
		if path != "" {
			paths = append(paths, path)
		}
	}
	sort.Strings(paths)
	return paths
}

func tracerouteIPs(raw map[string]any) []string {
	tr, ok := raw["traceroute"].(map[string]any)
	if !ok {
		return nil
	}
	hopsAny, ok := tr["hops"].([]any)
	if !ok {
		return nil
	}
	var ips []string
	for _, hopAny := range hopsAny {
		hop, ok := hopAny.(map[string]any)
		if !ok || hop["timeout"] == true {
			continue
		}
		ip := stringProp(hop, "ip")
		if ip != "" {
			ips = append(ips, ip)
		}
	}
	return ips
}

func stringDiff(previous, current []string) (added, removed []string) {
	prevSet := make(map[string]struct{}, len(previous))
	for _, item := range previous {
		prevSet[item] = struct{}{}
	}
	currSet := make(map[string]struct{}, len(current))
	for _, item := range current {
		currSet[item] = struct{}{}
	}

	for item := range currSet {
		if _, ok := prevSet[item]; !ok {
			added = append(added, item)
		}
	}
	for item := range prevSet {
		if _, ok := currSet[item]; !ok {
			removed = append(removed, item)
		}
	}
	sort.Strings(added)
	sort.Strings(removed)
	return added, removed
}

