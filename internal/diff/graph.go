package diff

import (
	"fmt"
	"sort"
	"strings"
)

// diffGraph detects BGP/traceroute path drift used by the graph views (G13).
func diffGraph(previous, current map[string]any) []Entry {
	var entries []Entry
	entries = append(entries, diffBGPRouteDrift(previous, current)...)
	entries = append(entries, diffTracerouteDrift(previous, current)...)
	return entries
}

func diffBGPRouteDrift(previous, current map[string]any) []Entry {
	prevRouting := routingSection(previous)
	curRouting := routingSection(current)
	if curRouting == nil {
		return nil
	}

	var entries []Entry

	prevRPKI := overallRPKI(prevRouting)
	curRPKI := overallRPKI(curRouting)
	if curRPKI != "" && curRPKI != prevRPKI {
		entries = append(entries, Entry{
			Type: "graph_rpki_changed", Severity: SeverityWarning, Field: "asn.routing.rpki",
			Summary: fmt.Sprintf("RPKI validation changed: %s → %s", emptyDash(prevRPKI), curRPKI),
		})
	}

	added, removed := stringSetDiff(visibleOrigins(prevRouting), visibleOrigins(curRouting))
	for _, origin := range added {
		entries = append(entries, Entry{
			Type: "graph_bgp_origin_added", Severity: SeverityWarning, Field: "asn.routing.visible_origins",
			Summary: fmt.Sprintf("New visible BGP origin: AS%s", origin),
			Detail:  origin,
		})
	}
	for _, origin := range removed {
		entries = append(entries, Entry{
			Type: "graph_bgp_origin_removed", Severity: SeverityWarning, Field: "asn.routing.visible_origins",
			Summary: fmt.Sprintf("BGP origin no longer visible: AS%s", origin),
			Detail:  origin,
		})
	}

	pathAdded, pathRemoved := stringSetDiff(asPaths(prevRouting), asPaths(curRouting))
	for _, path := range pathAdded {
		severity := SeverityInfo
		if len(pathAdded) > 0 && len(pathRemoved) > 0 {
			severity = SeverityWarning
		}
		entries = append(entries, Entry{
			Type: "graph_as_path_added", Severity: severity, Field: "asn.routing.as_paths",
			Summary: fmt.Sprintf("New BGP AS path: %s", truncatePath(path)),
			Detail:  path,
		})
	}
	for _, path := range pathRemoved {
		entries = append(entries, Entry{
			Type: "graph_as_path_removed", Severity: SeverityWarning, Field: "asn.routing.as_paths",
			Summary: fmt.Sprintf("BGP AS path removed: %s", truncatePath(path)),
			Detail:  path,
		})
	}

	return entries
}

func diffTracerouteDrift(previous, current map[string]any) []Entry {
	prevHops := tracerouteHopIPs(previous)
	curHops := tracerouteHopIPs(current)
	if len(curHops) == 0 && len(prevHops) == 0 {
		return nil
	}

	added, removed := stringSetDiff(prevHops, curHops)
	var entries []Entry
	for _, hop := range added {
		entries = append(entries, Entry{
			Type: "graph_traceroute_hop_added", Severity: SeverityInfo, Field: "traceroute.hops",
			Summary: fmt.Sprintf("Traceroute hop added: %s", hop),
			Detail:  hop,
		})
	}
	for _, hop := range removed {
		entries = append(entries, Entry{
			Type: "graph_traceroute_hop_removed", Severity: SeverityWarning, Field: "traceroute.hops",
			Summary: fmt.Sprintf("Traceroute hop removed: %s", hop),
			Detail:  hop,
		})
	}
	if len(added) == 0 && len(removed) == 0 && len(prevHops) > 0 && len(curHops) > 0 && !sameOrdered(prevHops, curHops) {
		entries = append(entries, Entry{
			Type: "graph_traceroute_reordered", Severity: SeverityWarning, Field: "traceroute.hops",
			Summary: "Traceroute hop order changed",
		})
	}
	return entries
}

func routingSection(raw map[string]any) map[string]any {
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
	if routing == nil {
		return nil
	}
	raw, ok := routing["visible_origins"].([]any)
	if !ok {
		return nil
	}
	var out []string
	for _, item := range raw {
		asn := strings.TrimSpace(fmt.Sprint(item))
		asn = strings.TrimPrefix(asn, "AS")
		if asn != "" {
			out = append(out, asn)
		}
	}
	sort.Strings(out)
	return out
}

func asPaths(routing map[string]any) []string {
	if routing == nil {
		return nil
	}
	raw, ok := routing["as_paths"].([]any)
	if !ok {
		return nil
	}
	var out []string
	for _, item := range raw {
		path := strings.TrimSpace(fmt.Sprint(item))
		if path != "" {
			out = append(out, path)
		}
	}
	sort.Strings(out)
	return out
}

func overallRPKI(routing map[string]any) string {
	if routing == nil {
		return ""
	}
	status, ok := routing["rpki"].(map[string]any)
	if !ok {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(status["overall"]))
}

func tracerouteHopIPs(raw map[string]any) []string {
	tr, ok := raw["traceroute"].(map[string]any)
	if !ok {
		return nil
	}
	hops, ok := tr["hops"].([]any)
	if !ok {
		return nil
	}
	var ips []string
	for _, hopAny := range hops {
		hop, ok := hopAny.(map[string]any)
		if !ok || hop["timeout"] == true {
			continue
		}
		ip := stringVal(hop, "ip")
		if ip != "" {
			ips = append(ips, ip)
		}
	}
	return ips
}

func stringSetDiff(previous, current []string) (added, removed []string) {
	prevSet := toStringSet(previous)
	currSet := toStringSet(current)
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

func sameOrdered(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func truncatePath(path string) string {
	if len(path) <= 80 {
		return path
	}
	return path[:77] + "..."
}
