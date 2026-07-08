package scanner

import (
	"strings"
)

func buildBGPPathProfile(paths []string) map[string]any {
	if len(paths) == 0 {
		return nil
	}

	shortest := paths[0]
	longest := paths[0]
	for _, path := range paths[1:] {
		if len(strings.Fields(path)) < len(strings.Fields(shortest)) {
			shortest = path
		}
		if len(strings.Fields(path)) > len(strings.Fields(longest)) {
			longest = path
		}
	}

	stability := "stable"
	switch {
	case len(paths) >= 5:
		stability = "volatile"
	case len(paths) >= 2:
		stability = "diverse"
	}

	return map[string]any{
		"path_count":   len(paths),
		"primary_path": shortest,
		"longest_path": longest,
		"stability":    stability,
		"unique_paths": len(paths),
	}
}
