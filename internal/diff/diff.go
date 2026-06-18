package diff

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/notfixingit3/echostate/internal/models"
)

// Entry is a structured field-level change between snapshots.
type Entry struct {
	Type     string `json:"type"`
	Severity string `json:"severity"`
	Summary  string `json:"summary"`
	Field    string `json:"field,omitempty"`
	Detail   string `json:"detail,omitempty"`
}

const (
	SeverityInfo     = "info"
	SeverityWarning  = "warning"
	SeverityCritical = "critical"
)

// Compute returns structured changes between previous raw_data and the current scan.
func Compute(previous map[string]any, current *models.ScanResult) []Entry {
	if len(previous) == 0 || current == nil {
		return nil
	}

	currentMap := map[string]any{}
	data, err := json.Marshal(current)
	if err != nil {
		return []Entry{{Type: "error", Severity: SeverityWarning, Summary: fmt.Sprintf("marshal current: %v", err)}}
	}
	if err := json.Unmarshal(data, &currentMap); err != nil {
		return []Entry{{Type: "error", Severity: SeverityWarning, Summary: fmt.Sprintf("unmarshal current: %v", err)}}
	}

	var entries []Entry
	entries = append(entries, diffTLS(previous, currentMap)...)
	entries = append(entries, diffBGP(previous, currentMap)...)
	entries = append(entries, diffGraph(previous, currentMap)...)
	entries = append(entries, diffCT(previous, currentMap)...)
	entries = append(entries, diffDMARC(previous, currentMap)...)
	entries = append(entries, diffWeb(previous, currentMap)...)
	entries = append(entries, diffGeneric(previous, currentMap)...)

	sort.Slice(entries, func(i, j int) bool {
		return severityRank(entries[i].Severity) > severityRank(entries[j].Severity)
	})
	return entries
}

func Summaries(entries []Entry) []string {
	out := make([]string, 0, len(entries))
	for _, entry := range entries {
		out = append(out, FormatSummary(entry))
	}
	return out
}

func FormatSummary(entry Entry) string {
	if entry.Severity == "" || entry.Severity == SeverityInfo {
		return entry.Summary
	}
	return fmt.Sprintf("[%s] %s", entry.Severity, entry.Summary)
}

func severityRank(severity string) int {
	switch severity {
	case SeverityCritical:
		return 3
	case SeverityWarning:
		return 2
	default:
		return 1
	}
}

func diffTLS(previous, current map[string]any) []Entry {
	prevTLS, _ := previous["tls"].(map[string]any)
	curTLS, _ := current["tls"].(map[string]any)
	if curTLS == nil {
		return nil
	}

	var entries []Entry
	prevIssuer := stringVal(prevTLS, "issuer")
	curIssuer := stringVal(curTLS, "issuer")
	if prevIssuer != "" && curIssuer != "" && prevIssuer != curIssuer {
		entries = append(entries, Entry{
			Type: "tls_issuer", Severity: SeverityWarning, Field: "tls.issuer",
			Summary: fmt.Sprintf("TLS issuer changed: %s → %s", prevIssuer, curIssuer),
		})
	}

	prevDays := intVal(prevTLS, "days_remaining")
	curDays := intVal(curTLS, "days_remaining")
	if curDays >= 0 && (prevDays < 0 || curDays < prevDays) {
		severity := SeverityInfo
		if curDays <= 7 {
			severity = SeverityCritical
		} else if curDays <= 30 {
			severity = SeverityWarning
		}
		entries = append(entries, Entry{
			Type: "cert_expiry", Severity: severity, Field: "tls.days_remaining",
			Summary: fmt.Sprintf("TLS certificate expires in %d days", curDays),
			Detail:  fmt.Sprintf("was %d days", prevDays),
		})
	}

	if boolVal(curTLS, "expired") && !boolVal(prevTLS, "expired") {
		entries = append(entries, Entry{
			Type: "cert_expired", Severity: SeverityCritical, Field: "tls.expired",
			Summary: "TLS certificate is expired",
		})
	}
	return entries
}

func diffBGP(previous, current map[string]any) []Entry {
	prevASN, _ := previous["asn"].(map[string]any)
	curASN, _ := current["asn"].(map[string]any)

	var entries []Entry
	if prevASN != nil && curASN == nil {
		if stringVal(prevASN, "asn") != "" || len(prevASN) > 0 {
			entries = append(entries, Entry{
				Type: "removed", Severity: SeverityInfo, Field: "asn",
				Summary: "Removed asn",
			})
		}
		return entries
	}
	if curASN == nil {
		return nil
	}

	prevRouting, _ := prevASN["routing"].(map[string]any)
	curRouting, _ := curASN["routing"].(map[string]any)

	prevOrigin := stringVal(prevASN, "asn")
	curOrigin := stringVal(curASN, "asn")
	if prevOrigin == "" && curOrigin != "" {
		entries = append(entries, Entry{
			Type: "added", Severity: SeverityInfo, Field: "asn",
			Summary: "Added asn",
			Detail:  fmt.Sprintf("AS%s", curOrigin),
		})
	}
	if prevOrigin != "" && curOrigin != "" && prevOrigin != curOrigin {
		entries = append(entries, Entry{
			Type: "bgp_origin", Severity: SeverityCritical, Field: "asn.asn",
			Summary: fmt.Sprintf("BGP origin ASN changed: AS%s → AS%s", prevOrigin, curOrigin),
		})
	}

	prevRisk := stringVal(prevRouting, "hijack_risk")
	curRisk := stringVal(curRouting, "hijack_risk")
	if curRisk != "" && curRisk != prevRisk {
		severity := SeverityWarning
		if strings.EqualFold(curRisk, "high") || strings.EqualFold(curRisk, "critical") {
			severity = SeverityCritical
		}
		entries = append(entries, Entry{
			Type: "bgp_hijack_risk", Severity: severity, Field: "asn.routing.hijack_risk",
			Summary: fmt.Sprintf("BGP hijack risk changed: %s → %s", emptyDash(prevRisk), curRisk),
		})
	}
	return entries
}

func diffCT(previous, current map[string]any) []Entry {
	prevCT, _ := previous["ct"].(map[string]any)
	curCT, _ := current["ct"].(map[string]any)
	if curCT == nil {
		return nil
	}

	prevSet := toStringSet(stringSlice(prevCT, "subdomains"))
	var added []string
	for _, name := range stringSlice(curCT, "subdomains") {
		if _, ok := prevSet[name]; !ok {
			added = append(added, name)
		}
	}
	sort.Strings(added)

	var entries []Entry
	for _, name := range added {
		entries = append(entries, Entry{
			Type: "new_ct_subdomain", Severity: SeverityInfo, Field: "ct.subdomains",
			Summary: fmt.Sprintf("New CT subdomain: %s", name),
			Detail:  name,
		})
	}
	return entries
}

func diffDMARC(previous, current map[string]any) []Entry {
	prevDNS, _ := previous["dns"].(map[string]any)
	curDNS, _ := current["dns"].(map[string]any)
	if curDNS == nil {
		return nil
	}
	prevParsed, _ := prevDNS["DMARC_PARSED"].(map[string]any)
	curParsed, _ := curDNS["DMARC_PARSED"].(map[string]any)
	if curParsed == nil {
		return nil
	}

	prevPolicy := stringVal(prevParsed, "policy")
	curPolicy := stringVal(curParsed, "policy")
	if curPolicy != "" && curPolicy != prevPolicy {
		return []Entry{{
			Type: "dmarc_policy", Severity: SeverityWarning, Field: "dns.dmarc.policy",
			Summary: fmt.Sprintf("DMARC policy changed: %s → %s", emptyDash(prevPolicy), curPolicy),
		}}
	}
	return nil
}

func diffWeb(previous, current map[string]any) []Entry {
	prevWeb, _ := previous["web"].(map[string]any)
	curWeb, _ := current["web"].(map[string]any)
	if curWeb == nil {
		return nil
	}
	if prevWeb == nil {
		return []Entry{{
			Type: "added", Severity: SeverityInfo, Field: "web",
			Summary: "Added web",
		}}
	}
	prevTitle := stringVal(prevWeb, "title")
	curTitle := stringVal(curWeb, "title")
	if curTitle != "" && curTitle != prevTitle {
		return []Entry{{
			Type: "web_title", Severity: SeverityInfo, Field: "web.title",
			Summary: fmt.Sprintf("Web title changed: %s → %s", emptyDash(prevTitle), curTitle),
		}}
	}
	return nil
}

func diffGeneric(previous, current map[string]any) []Entry {
	tracked := map[string]bool{"tls": true, "asn": true, "ct": true, "dns": true, "web": true, "screenshot": true, "scanned_at": true, "host": true, "errors": true}
	var entries []Entry
	for key, curVal := range current {
		if tracked[key] {
			continue
		}
		prevVal, ok := previous[key]
		if !ok {
			entries = append(entries, Entry{Type: "added", Severity: SeverityInfo, Field: key, Summary: fmt.Sprintf("Added %s", key)})
			continue
		}
		if !jsonEqual(prevVal, curVal) {
			entries = append(entries, Entry{Type: "changed", Severity: SeverityInfo, Field: key, Summary: fmt.Sprintf("Changed %s", key)})
		}
	}
	for key := range previous {
		if tracked[key] {
			continue
		}
		if _, ok := current[key]; !ok {
			entries = append(entries, Entry{Type: "removed", Severity: SeverityInfo, Field: key, Summary: fmt.Sprintf("Removed %s", key)})
		}
	}
	return entries
}

func stringVal(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	v, ok := m[key]
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return fmt.Sprint(v)
	}
	return strings.TrimSpace(s)
}

func intVal(m map[string]any, key string) int {
	if m == nil {
		return -1
	}
	switch v := m[key].(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case json.Number:
		i, _ := v.Int64()
		return int(i)
	default:
		return -1
	}
}

func boolVal(m map[string]any, key string) bool {
	if m == nil {
		return false
	}
	v, ok := m[key].(bool)
	return ok && v
}

func stringSlice(m map[string]any, key string) []string {
	if m == nil {
		return nil
	}
	raw, ok := m[key].([]any)
	if !ok {
		if typed, ok := m[key].([]string); ok {
			return typed
		}
		return nil
	}
	var out []string
	for _, item := range raw {
		if s, ok := item.(string); ok && strings.TrimSpace(s) != "" {
			out = append(out, strings.TrimSpace(s))
		}
	}
	return out
}

func toStringSet(items []string) map[string]struct{} {
	set := make(map[string]struct{}, len(items))
	for _, item := range items {
		set[item] = struct{}{}
	}
	return set
}

func jsonEqual(a, b any) bool {
	aj, _ := json.Marshal(a)
	bj, _ := json.Marshal(b)
	return string(aj) == string(bj)
}

func emptyDash(value string) string {
	if value == "" {
		return "—"
	}
	return value
}

// MatchesRule reports whether an entry passes an alert rule filter.
func MatchesRule(entry Entry, matchTypes []string, minSeverity string) bool {
	if len(matchTypes) > 0 {
		found := false
		for _, t := range matchTypes {
			if entry.Type == t || entry.Field == t {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return severityRank(entry.Severity) >= severityRank(minSeverity)
}

// ParseRFC3339 is exported for tests.
func ParseRFC3339(value string) (time.Time, bool) {
	t, err := time.Parse(time.RFC3339, value)
	return t, err == nil
}