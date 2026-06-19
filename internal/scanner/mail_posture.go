package scanner

import (
	"fmt"
	"strings"
)

func computeMailPosture(data map[string]any) map[string]any {
	score := 100
	var findings []string

	spf, hasSPF := data["SPF"].(map[string]any)
	switch {
	case !hasSPF:
		score -= 25
		findings = append(findings, "No SPF record published")
	default:
		policy := strings.ToLower(strings.TrimSpace(stringValMap(spf, "policy")))
		switch {
		case strings.HasSuffix(policy, "+all"):
			score -= 45
			findings = append(findings, "SPF policy allows any sender (+all)")
		case strings.HasSuffix(policy, "?all"):
			score -= 20
			findings = append(findings, "SPF policy is neutral (?all)")
		case strings.HasSuffix(policy, "~all"):
			score -= 10
			findings = append(findings, "SPF policy is soft-fail (~all)")
		case strings.HasSuffix(policy, "-all"):
			findings = append(findings, "SPF policy is strict (-all)")
		default:
			score -= 10
			findings = append(findings, "SPF record present but terminal policy unclear")
		}
	}

	dmarc, hasDMARC := data["DMARC_PARSED"].(map[string]any)
	switch {
	case !hasDMARC:
		score -= 30
		findings = append(findings, "No DMARC policy at _dmarc")
	default:
		policy := strings.ToLower(strings.TrimSpace(stringValMap(dmarc, "policy")))
		switch policy {
		case "reject":
			findings = append(findings, "DMARC policy is reject")
		case "quarantine":
			score -= 5
			findings = append(findings, "DMARC policy is quarantine")
		case "none":
			score -= 20
			findings = append(findings, "DMARC policy is none (monitoring only)")
		default:
			score -= 15
			findings = append(findings, "DMARC record present but policy is unclear")
		}
	}

	dkimRecords, hasDKIM := data["DKIM"].([]any)
	if !hasDKIM || len(dkimRecords) == 0 {
		score -= 20
		findings = append(findings, "No DKIM selectors found for common names")
	} else {
		findings = append(findings, fmt.Sprintf("DKIM keys found (%d selector(s))", len(dkimRecords)))
	}

	mtaSts, hasMTASTS := data["MTA_STS"].(map[string]any)
	switch {
	case !hasMTASTS:
		score -= 10
		findings = append(findings, "No MTA-STS policy published")
	default:
		mode := strings.ToLower(stringValMap(mtaSts, "mode"))
		switch mode {
		case "enforce":
			findings = append(findings, "MTA-STS mode is enforce")
		case "testing":
			score -= 5
			findings = append(findings, "MTA-STS mode is testing")
		default:
			score -= 8
			findings = append(findings, "MTA-STS policy present but mode is not enforce")
		}
	}

	if _, hasTLSRPT := data["TLS_RPT"].(map[string]any); !hasTLSRPT {
		score -= 5
		findings = append(findings, "No TLS-RPT record at _smtp._tls")
	} else {
		findings = append(findings, "TLS-RPT reporting configured")
	}

	if _, hasBIMI := data["BIMI"].(map[string]any); !hasBIMI {
		score -= 3
		findings = append(findings, "No BIMI record at default._bimi")
	} else {
		findings = append(findings, "BIMI record published")
	}

	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return map[string]any{
		"score":    score,
		"grade":    mailPostureGrade(score),
		"findings": findings,
	}
}

func mailPostureGrade(score int) string {
	switch {
	case score >= 90:
		return "A"
	case score >= 80:
		return "B"
	case score >= 65:
		return "C"
	case score >= 50:
		return "D"
	default:
		return "F"
	}
}

func stringValMap(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	v, ok := m[key]
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return strings.TrimSpace(fmt.Sprint(v))
	}
	return strings.TrimSpace(s)
}