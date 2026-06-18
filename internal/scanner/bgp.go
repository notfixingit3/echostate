package scanner

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const ripestatBaseURL = "https://stat.ripe.net/data"

func enrichRouting(ctx context.Context, ip, originASN, prefix string) (map[string]any, error) {
	ip = strings.TrimSpace(ip)
	if ip == "" {
		return nil, fmt.Errorf("empty ip")
	}

	routing := map[string]any{
		"hijack_risk": "unknown",
		"notes":       []string{},
	}

	status, err := ripeStat(ctx, "routing-status", ip)
	if err == nil {
		if data, ok := status["data"].(map[string]any); ok {
			if visibility, ok := data["visibility"].(map[string]any); ok {
				routing["visibility"] = visibility
			}
			if firstSeen, ok := data["first_seen"]; ok {
				routing["first_seen"] = firstSeen
			}
			if lastSeen, ok := data["last_seen"]; ok {
				routing["last_seen"] = lastSeen
			}
		}
	} else {
		appendNote(routing, "routing-status lookup failed: "+err.Error())
	}

	origins := extractVisibleOrigins(routing)
	routing["visible_origins"] = origins

	risk, notes := assessHijackRisk(originASN, prefix, origins, routing)
	routing["hijack_risk"] = risk
	for _, note := range notes {
		appendNote(routing, note)
	}

	resource := ip
	if prefix != "" {
		resource = prefix
	}
	if paths, err := fetchASPathsFunc(ctx, resource); err == nil && len(paths) > 0 {
		routing["as_paths"] = paths
		if profile := buildBGPPathProfile(paths); profile != nil {
			routing["path_profile"] = profile
		}
	} else if err != nil {
		appendNote(routing, "as-path lookup failed: "+err.Error())
	}

	if prefix != "" {
		if rpki, err := ripeStat(ctx, "rpki-validation", prefix); err == nil {
			routing["rpki"] = rpki["data"]
			if valid := rpkiValidationState(rpki); valid == "invalid" {
				if risk != "high" {
					routing["hijack_risk"] = "high"
				}
				appendNote(routing, "RPKI validation reported invalid route origin")
			} else if valid == "valid" {
				appendNote(routing, "RPKI validation passed for announced prefix")
			}
		}
	}

	return routing, nil
}

func ripeStat(ctx context.Context, endpoint, resource string) (map[string]any, error) {
	url := fmt.Sprintf("%s/%s/data.json?resource=%s", ripestatBaseURL, endpoint, resource)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("ripestat HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if err != nil {
		return nil, err
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}

	return payload, nil
}

func extractVisibleOrigins(routing map[string]any) []string {
	visibility, ok := routing["visibility"].(map[string]any)
	if !ok {
		return nil
	}

	seen := make(map[string]struct{})
	var origins []string

	for _, value := range visibility {
		entries, ok := value.([]any)
		if !ok {
			continue
		}
		for _, entry := range entries {
			row, ok := entry.(map[string]any)
			if !ok {
				continue
			}
			if origin, ok := row["origin_asn"].(string); ok && origin != "" {
				if _, exists := seen[origin]; !exists {
					seen[origin] = struct{}{}
					origins = append(origins, origin)
				}
			}
		}
	}

	return origins
}

func assessHijackRisk(originASN, prefix string, visibleOrigins []string, routing map[string]any) (string, []string) {
	originASN = normalizeASN(originASN)
	var notes []string

	if originASN == "" {
		return "unknown", []string{"No Team Cymru origin ASN available for comparison"}
	}

	if len(visibleOrigins) == 0 {
		return "unknown", []string{"No visible BGP origins returned by RIPEstat"}
	}

	matches := false
	for _, visible := range visibleOrigins {
		if normalizeASN(visible) == originASN {
			matches = true
			break
		}
	}

	if !matches {
		return "high", []string{
			fmt.Sprintf("RIPEstat visible origins %v do not include Cymru origin AS%s", visibleOrigins, originASN),
		}
	}

	if len(visibleOrigins) > 1 {
		notes = append(notes, fmt.Sprintf("Multiple origin ASNs visible in routing data: %v", visibleOrigins))
		return "medium", notes
	}

	if prefix != "" {
		notes = append(notes, fmt.Sprintf("Origin AS%s consistently visible for prefix %s", originASN, prefix))
	}

	return "low", notes
}

func rpkiValidationState(payload map[string]any) string {
	data, ok := payload["data"].(map[string]any)
	if !ok {
		return ""
	}
	validating, ok := data["validating_roas"].([]any)
	if !ok || len(validating) == 0 {
		return ""
	}
	for _, item := range validating {
		row, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if validity, ok := row["validity"].(string); ok {
			return strings.ToLower(validity)
		}
	}
	return ""
}

func normalizeASN(asn string) string {
	asn = strings.TrimSpace(asn)
	asn = strings.TrimPrefix(strings.ToUpper(asn), "AS")
	return asn
}

func appendNote(routing map[string]any, note string) {
	notes, _ := routing["notes"].([]string)
	routing["notes"] = append(notes, note)
}