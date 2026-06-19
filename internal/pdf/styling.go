package pdf

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/johnfercher/maroto/v2/pkg/props"

	"github.com/notfixingit3/echostate/internal/models"
)

var (
	colorTeal      = props.Color{Red: 13, Green: 148, Blue: 136}
	colorTealDark  = props.Color{Red: 15, Green: 118, Blue: 110}
	colorGreen     = props.Color{Red: 22, Green: 163, Blue: 74}
	colorAmber     = props.Color{Red: 217, Green: 119, Blue: 6}
	colorRed       = props.Color{Red: 220, Green: 38, Blue: 38}
	colorMuted     = props.Color{Red: 100, Green: 116, Blue: 139}
	colorHeaderBG  = props.Color{Red: 240, Green: 253, Blue: 250}
	colorTableHead = props.Color{Red: 241, Green: 245, Blue: 249}
	colorTableRow  = props.Color{Red: 248, Green: 250, Blue: 252}
	colorLink      = props.Color{Red: 8, Green: 145, Blue: 178}
)

type securityFinding struct {
	severity string
	summary  string
	detail   string
}

func colorPtr(c props.Color) *props.Color {
	return &c
}

func hyperlinkURL(raw string) *string {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "N/A" {
		return nil
	}
	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		return &raw
	}
	return nil
}

func colorForSeverity(severity string) *props.Color {
	switch strings.ToLower(strings.TrimSpace(severity)) {
	case "critical", "high":
		return colorPtr(colorRed)
	case "warning", "medium":
		return colorPtr(colorAmber)
	case "info", "low":
		return colorPtr(colorMuted)
	default:
		return nil
	}
}

func colorForHijackRisk(value string) *props.Color {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "high", "critical":
		return colorPtr(colorRed)
	case "medium", "moderate":
		return colorPtr(colorAmber)
	case "low", "none":
		return colorPtr(colorGreen)
	default:
		return nil
	}
}

func colorForMailGrade(grade string) *props.Color {
	switch strings.ToUpper(strings.TrimSpace(grade)) {
	case "A", "B":
		return colorPtr(colorGreen)
	case "C":
		return colorPtr(colorAmber)
	case "D", "F":
		return colorPtr(colorRed)
	default:
		return nil
	}
}

func colorForCertDays(days string) *props.Color {
	if days == "N/A" {
		return nil
	}
	n, err := strconv.Atoi(strings.TrimSpace(days))
	if err != nil {
		return nil
	}
	switch {
	case n < 0:
		return colorPtr(colorRed)
	case n <= 30:
		return colorPtr(colorAmber)
	case n <= 90:
		return colorPtr(colorMuted)
	default:
		return colorPtr(colorGreen)
	}
}

func colorForBoolish(value string, goodValues ...string) *props.Color {
	v := strings.ToLower(strings.TrimSpace(value))
	for _, good := range goodValues {
		if v == strings.ToLower(good) {
			return colorPtr(colorGreen)
		}
	}
	if v == "false" || v == "no" || v == "unsigned" || v == "none" {
		return colorPtr(colorAmber)
	}
	return nil
}

func collectSecurityFindings(result *models.ScanResult, details []models.ChangeDetail) []securityFinding {
	var findings []securityFinding

	appendFinding := func(severity, summary, detail string) {
		findings = append(findings, securityFinding{
			severity: severity,
			summary:  summary,
			detail:   detail,
		})
	}

	if risk := stringVal(nestedMap(result.ASN, "routing"), "hijack_risk"); risk != "N/A" {
		if c := colorForHijackRisk(risk); c == colorPtr(colorRed) || c == colorPtr(colorAmber) {
			appendFinding("warning", "BGP hijack risk is "+risk, "Review visible origins in the ASN / BGP section.")
		}
	}

	if days := stringVal(result.TLS, "days_remaining"); days != "N/A" {
		if c := colorForCertDays(days); c == colorPtr(colorRed) || c == colorPtr(colorAmber) {
			appendFinding("critical", "TLS certificate expires in "+days+" days", stringVal(result.TLS, "not_after"))
		}
	}
	if expired := stringVal(result.TLS, "expired"); strings.EqualFold(expired, "true") {
		appendFinding("critical", "TLS certificate is expired", stringVal(result.TLS, "not_after"))
	}

	if posture := nestedMap(result.DNS, "MAIL_POSTURE"); len(posture) > 0 {
		grade := stringVal(posture, "grade")
		if c := colorForMailGrade(grade); c == colorPtr(colorRed) || c == colorPtr(colorAmber) {
			score := stringVal(posture, "score")
			appendFinding("warning", fmt.Sprintf("Email posture grade %s (%s/100)", grade, score), "See DNS mail posture findings.")
		}
	}

	if dnssec := nestedMap(result.DNS, "DNSSEC"); len(dnssec) > 0 {
		status := stringVal(dnssec, "status")
		if strings.EqualFold(status, "unsigned") {
			appendFinding("info", "DNSSEC is unsigned", "Zone does not publish DNSSEC signatures.")
		}
	}

	for _, detail := range details {
		if detail.Severity == "" && detail.Summary == "" {
			continue
		}
		appendFinding(detail.Severity, detail.Summary, detail.Field)
	}

	if len(result.Errors) > 0 {
		appendFinding("warning", fmt.Sprintf("%d gatherer error(s)", len(result.Errors)), result.Errors[0])
	}

	if len(findings) == 0 {
		appendFinding("info", "No elevated security findings detected", "Review intel highlights and section detail for context.")
	}

	return findings
}