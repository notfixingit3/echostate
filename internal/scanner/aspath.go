package scanner

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

const maxStoredASPaths = 8

var fetchASPathsFunc = fetchASPaths

func fetchASPaths(ctx context.Context, resource string) ([]string, error) {
	resource = strings.TrimSpace(resource)
	if resource == "" {
		return nil, fmt.Errorf("empty resource")
	}

	payload, err := ripeStat(ctx, "bgp-state", resource)
	if err != nil {
		return nil, err
	}

	data, ok := payload["data"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("missing bgp-state data")
	}

	states, ok := data["bgp_state"].([]any)
	if !ok || len(states) == 0 {
		return nil, fmt.Errorf("no bgp-state entries")
	}

	seen := make(map[string]struct{})
	var paths []string

	for _, stateAny := range states {
		state, ok := stateAny.(map[string]any)
		if !ok {
			continue
		}
		pathAny, ok := state["path"].([]any)
		if !ok || len(pathAny) == 0 {
			continue
		}

		var asns []string
		for _, asnAny := range pathAny {
			switch v := asnAny.(type) {
			case float64:
				asns = append(asns, strconv.Itoa(int(v)))
			case int:
				asns = append(asns, strconv.Itoa(v))
			case int64:
				asns = append(asns, strconv.Itoa(int(v)))
			case string:
				asns = append(asns, normalizeASN(v))
			}
		}
		if len(asns) == 0 {
			continue
		}

		joined := strings.Join(asns, " ")
		if _, exists := seen[joined]; exists {
			continue
		}
		seen[joined] = struct{}{}
		paths = append(paths, joined)
	}

	if len(paths) == 0 {
		return nil, fmt.Errorf("no AS paths found")
	}

	sort.Slice(paths, func(i, j int) bool {
		return len(strings.Fields(paths[i])) < len(strings.Fields(paths[j]))
	})
	if len(paths) > maxStoredASPaths {
		paths = paths[:maxStoredASPaths]
	}
	return paths, nil
}

func parseASPath(path string) []string {
	fields := strings.Fields(strings.TrimSpace(path))
	var asns []string
	for _, field := range fields {
		asn := normalizeASN(field)
		if asn != "" {
			asns = append(asns, asn)
		}
	}
	return asns
}