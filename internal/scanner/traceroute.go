package scanner

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/notfixingit3/echostate/internal/config"
)

const (
	tracerouteTimeout       = 35 * time.Second
	vantageLocal              = "local"
	vantageExternal           = "external"
)

// HackerTarget discontinued the public traceroute API (returns HTTP 404).
var errExternalTracerouteUnavailable = errors.New("HackerTarget traceroute API unavailable")

var (
	runTracerouteFunc      = runTraceroute
	fetchExternalTraceFunc = fetchExternalTraceroute
	hopLinePattern         = regexp.MustCompile(`^\s*(\d+)\s+(\S+)(?:\s+([\d.]+)\s*ms)?`)
	rttPattern             = regexp.MustCompile(`([\d.]+)\s*ms`)
)

type tracerouteHop struct {
	Hop     int     `json:"hop"`
	IP      string  `json:"ip,omitempty"`
	RTTMs   float64 `json:"rtt_ms,omitempty"`
	Timeout bool    `json:"timeout,omitempty"`
}

func gatherTraceroute(ctx context.Context, host string) (string, map[string]any, error) {
	host = NormalizeHost(host)
	if host == "" {
		return "traceroute", nil, fmt.Errorf("empty host")
	}

	if net.ParseIP(host) != nil {
		return "traceroute", map[string]any{
			"skipped": "traceroute is not run for raw IP targets",
		}, nil
	}

	ctx, cancel := context.WithTimeout(ctx, tracerouteTimeout)
	defer cancel()

	var (
		localOutput string
		localErr    error
		extOutput   string
		extErr      error
		wg          sync.WaitGroup
	)

	wg.Add(2)
	go func() {
		defer wg.Done()
		localOutput, localErr = runTracerouteFunc(ctx, host)
	}()
	go func() {
		defer wg.Done()
		extOutput, extErr = fetchExternalTraceFunc(ctx, host)
	}()
	wg.Wait()

	localHops := parseTracerouteOutput(localOutput)
	localHops, localPrefixRedacted := redactScannerTraceroutePrefix(localHops)
	enrichHopGeo(ctx, localHops)

	externalHops := parseTracerouteOutput(extOutput)
	enrichHopGeo(ctx, externalHops)

	vantages := []map[string]any{
		buildVantageResult(vantageLocal, "Local scanner", "traceroute", localHops, localErr),
		buildVantageResult(vantageExternal, "External vantage", "remote-api", externalHops, extErr),
	}
	if localPrefixRedacted > 0 && len(vantages) > 0 {
		vantages[0]["prefix_redacted"] = localPrefixRedacted
	}

	result := map[string]any{
		"destination": host,
		"hops":        localHops,
		"hop_count":   len(localHops),
		"vantages":    vantages,
	}
	if localPrefixRedacted > 0 {
		result["local_prefix_redacted"] = localPrefixRedacted
	}

	var warnings []string
	if localErr != nil {
		warnings = append(warnings, formatTracerouteVantageWarning("local", localErr))
	}
	if extErr != nil && !errors.Is(extErr, errExternalTracerouteUnavailable) {
		warnings = append(warnings, formatTracerouteVantageWarning("external", extErr))
	}
	if len(warnings) > 0 {
		result["warning"] = strings.Join(warnings, "; ")
	}

	if len(localHops) == 0 && len(externalHops) == 0 {
		if len(warnings) > 0 {
			return "traceroute", result, fmt.Errorf("%s", result["warning"])
		}
		return "traceroute", result, fmt.Errorf("no traceroute hops collected")
	}

	return "traceroute", result, nil
}

func redactScannerTraceroutePrefix(hops []map[string]any) ([]map[string]any, int) {
	settings := config.GetSettings()
	if !config.TracerouteRedactScannerPrefixEnabled(settings) {
		return hops, 0
	}

	skip := 0
	for skip < len(hops) {
		if isScannerPrefixHop(hops[skip]) {
			skip++
			continue
		}
		break
	}

	// Hide the first public hop after local/private hops when it is not the destination.
	if skip < len(hops)-1 && !hopTimedOut(hops[skip]) {
		ip := stringProp(hops[skip], "ip")
		if parsed := net.ParseIP(ip); parsed != nil && !isNonPublicScannerHop(parsed) {
			skip++
		}
	}

	extra := settings.TracerouteRedactLocalExtraHops
	if extra > 0 {
		skip += extra
	}
	if skip <= 0 {
		return hops, 0
	}
	if skip >= len(hops) {
		return nil, skip
	}

	trimmed := append([]map[string]any(nil), hops[skip:]...)
	renumberTracerouteHops(trimmed)
	return trimmed, skip
}

func isScannerPrefixHop(hop map[string]any) bool {
	if hopTimedOut(hop) {
		return true
	}
	ip := stringProp(hop, "ip")
	if ip == "" {
		return true
	}
	parsed := net.ParseIP(ip)
	return parsed != nil && isNonPublicScannerHop(parsed)
}

func hopTimedOut(hop map[string]any) bool {
	return hop["timeout"] == true
}

func isNonPublicScannerHop(ip net.IP) bool {
	return ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast()
}

func renumberTracerouteHops(hops []map[string]any) {
	for index, hop := range hops {
		hop["hop"] = index + 1
	}
}

func buildVantageResult(id, label, source string, hops []map[string]any, err error) map[string]any {
	result := map[string]any{
		"id":        id,
		"label":     label,
		"source":    source,
		"hops":      hops,
		"hop_count": len(hops),
	}
	if err != nil {
		if errors.Is(err, errExternalTracerouteUnavailable) {
			result["skipped"] = err.Error()
		} else {
			result["warning"] = err.Error()
		}
	}
	return result
}

func formatTracerouteVantageWarning(label string, err error) string {
	msg := strings.TrimSpace(err.Error())
	for _, prefix := range []string{"external traceroute: ", "traceroute: "} {
		msg = strings.TrimPrefix(msg, prefix)
	}
	return label + ": " + msg
}

func fetchExternalTraceroute(ctx context.Context, host string) (string, error) {
	host = NormalizeHost(host)
	if host == "" {
		return "", fmt.Errorf("empty host")
	}

	_ = ctx
	return "", errExternalTracerouteUnavailable
}

func runTraceroute(ctx context.Context, host string) (string, error) {
	cmd := exec.CommandContext(
		ctx,
		"traceroute",
		"-n",
		"-w", "2",
		"-m", "20",
		"-q", "1",
		host,
	)
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil && output == "" {
		return "", fmt.Errorf("traceroute: %w", err)
	}
	return output, nil
}

func parseTracerouteOutput(output string) []map[string]any {
	if strings.TrimSpace(output) == "" {
		return nil
	}

	var hops []map[string]any
	seen := make(map[int]struct{})

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "traceroute") {
			continue
		}

		hop, ok := parseTracerouteLine(line)
		if !ok {
			continue
		}
		if _, exists := seen[hop.Hop]; exists {
			continue
		}
		seen[hop.Hop] = struct{}{}
		hops = append(hops, hopToMap(hop))
	}

	return hops
}

func parseTracerouteLine(line string) (tracerouteHop, bool) {
	matches := hopLinePattern.FindStringSubmatch(line)
	if len(matches) < 3 {
		return tracerouteHop{}, false
	}

	hopNum, err := strconv.Atoi(matches[1])
	if err != nil || hopNum <= 0 {
		return tracerouteHop{}, false
	}

	token := strings.TrimSpace(matches[2])
	if token == "*" || strings.HasPrefix(token, "*") {
		return tracerouteHop{Hop: hopNum, Timeout: true}, true
	}

	ip := extractIPFromHopToken(token, line)

	hop := tracerouteHop{Hop: hopNum, IP: ip}
	if len(matches) >= 4 && matches[3] != "" {
		if rtt, err := strconv.ParseFloat(matches[3], 64); err == nil {
			hop.RTTMs = rtt
		}
	} else if rttMatch := rttPattern.FindStringSubmatch(line); len(rttMatch) >= 2 {
		if rtt, err := strconv.ParseFloat(rttMatch[1], 64); err == nil {
			hop.RTTMs = rtt
		}
	}
	return hop, true
}

func extractIPFromHopToken(token, line string) string {
	if net.ParseIP(token) != nil {
		return token
	}

	if idx := strings.Index(line, "("); idx >= 0 {
		end := strings.Index(line[idx:], ")")
		if end > 1 {
			candidate := strings.TrimSpace(line[idx+1 : idx+end])
			if net.ParseIP(candidate) != nil {
				return candidate
			}
		}
	}

	if net.ParseIP(token) != nil {
		return token
	}
	return token
}

func hopToMap(hop tracerouteHop) map[string]any {
	m := map[string]any{
		"hop": hop.Hop,
	}
	if hop.Timeout {
		m["timeout"] = true
	}
	if hop.IP != "" {
		m["ip"] = hop.IP
	}
	if hop.RTTMs > 0 {
		m["rtt_ms"] = hop.RTTMs
	}
	return m
}

func tracerouteVantages(trMap map[string]any) []map[string]any {
	if vantagesAny, ok := trMap["vantages"].([]any); ok && len(vantagesAny) > 0 {
		var vantages []map[string]any
		for _, item := range vantagesAny {
			if vantage, ok := item.(map[string]any); ok {
				vantages = append(vantages, vantage)
			}
		}
		if len(vantages) > 0 {
			return vantages
		}
	}

	hopsAny, ok := trMap["hops"].([]any)
	if !ok || len(hopsAny) == 0 {
		return nil
	}

	var hops []map[string]any
	for _, hopAny := range hopsAny {
		if hop, ok := hopAny.(map[string]any); ok {
			hops = append(hops, hop)
		}
	}
	if len(hops) == 0 {
		return nil
	}

	return []map[string]any{
		{
			"id":        vantageLocal,
			"label":     "Local scanner",
			"source":    "traceroute",
			"hops":      hops,
			"hop_count": len(hops),
		},
	}
}