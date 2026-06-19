package pdf

import (
	"fmt"
	"strings"
	"time"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/image"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/extension"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"

	"github.com/notfixingit3/echostate/internal/models"
)

const (
	maxCopyrights        = 10
	maxErrorsShown       = 20
	maxChangesShown      = 25
	maxChangeDetails     = 20
	maxCtSubdomainsShown = 40
	maxTracerouteHops    = 30
	maxTechStackItems    = 20
	maxJSAssets          = 15
	maxBuckets           = 20
	maxCrawlPaths        = 25
	pdfDateFormat        = "2006-01-02 15:04:05 UTC"
	reportProjectURL     = "https://github.com/notfixingit3/echostate"
	coverLogoHeight      = 40.0
	coverLogoPercent     = 32.0
	screenshotRowHeight  = 55.0
)

type reportSection struct {
	title  string
	render func(core.Maroto)
}

// RenderReport renders snapshot reconnaissance data as a PDF using maroto v2.
func RenderReport(data ReportData) ([]byte, error) {
	if data.Result == nil {
		return nil, fmt.Errorf("scan result is nil")
	}
	result := data.Result

	sections := buildReportSections(data)
	sectionPages, err := measureSectionPages(data, sections)
	if err != nil {
		return nil, err
	}

	m := buildReportMaroto(result.Host)
	addCoverPage(m, result)
	addTableOfContents(m, sections, sectionPages)
	for _, section := range sections {
		addSectionHeader(m, section.title)
		section.render(m)
	}

	doc, err := m.Generate()
	if err != nil {
		return nil, fmt.Errorf("generate PDF: %w", err)
	}

	pdfBytes, err := injectBookmarks(doc.GetBytes(), data, sections, sectionPages)
	if err != nil {
		return nil, err
	}
	return pdfBytes, nil
}

func buildReportMaroto(host string) core.Maroto {
	cfg := config.NewBuilder().
		WithTitle("EchoState Reconnaissance Report", true).
		WithCreator("EchoState", true).
		WithSubject(fmt.Sprintf("Reconnaissance report for %s", host), true).
		WithPageNumber(props.PageNumber{
			Pattern: "Page {current} of {total}",
			Place:   props.Bottom,
			Size:    8,
		}).
		Build()
	return maroto.New(cfg)
}

func buildReportSections(data ReportData) []reportSection {
	result := data.Result
	findings := collectSecurityFindings(result, data.ChangeDetails)

	sections := []reportSection{
		{"Report Metadata", func(m core.Maroto) { addReportMetadata(m, result, data.ClientIP) }},
		{"Security Findings", func(m core.Maroto) { addSecurityFindings(m, findings) }},
		{"Intel Highlights", func(m core.Maroto) { addIntelHighlights(m, result, data.PWhois) }},
	}

	if len(data.Changes) > 0 || len(data.ChangeDetails) > 0 {
		sections = append(sections, reportSection{
			"Snapshot Changes",
			func(m core.Maroto) { addChanges(m, data.Changes, data.ChangeDetails) },
		})
	}
	if len(data.ScreenshotJPEG) > 0 || len(result.Screenshot) > 0 {
		sections = append(sections, reportSection{
			"Web Screenshot",
			func(m core.Maroto) { addScreenshot(m, data.ScreenshotJPEG, result.Screenshot) },
		})
	}
	if len(data.PWhoisData) > 0 || (data.PWhois != nil && data.PWhois.HasData()) {
		sections = append(sections, reportSection{
			"Submitter / pWhois",
			func(m core.Maroto) { addPWhoisSection(m, data.PWhois, data.PWhoisData) },
		})
	}

	sections = append(sections,
		reportSection{"WHOIS", func(m core.Maroto) { addWHOIS(m, result.WHOIS) }},
		reportSection{"ASN / BGP", func(m core.Maroto) { addASN(m, result.ASN) }},
		reportSection{"DNS", func(m core.Maroto) { addDNS(m, result.DNS) }},
		reportSection{"TLS / Certificate", func(m core.Maroto) { addTLS(m, result.TLS) }},
		reportSection{"Web", func(m core.Maroto) { addWeb(m, result.Web) }},
		reportSection{"Favicon", func(m core.Maroto) { addFavicon(m, result.Favicon) }},
		reportSection{"Crawl", func(m core.Maroto) { addCrawl(m, result.Crawl) }},
		reportSection{"Storage", func(m core.Maroto) { addStorage(m, result.Storage) }},
		reportSection{"Certificate Transparency", func(m core.Maroto) { addCT(m, result.CT) }},
		reportSection{"Traceroute", func(m core.Maroto) { addTraceroute(m, result.Traceroute) }},
	)

	if len(result.Enrichment) > 0 {
		sections = append(sections, reportSection{
			"Third-Party Enrichment",
			func(m core.Maroto) { addEnrichment(m, result.Enrichment) },
		})
	}
	if len(result.Errors) > 0 {
		sections = append(sections, reportSection{
			"Errors",
			func(m core.Maroto) { addErrors(m, result.Errors) },
		})
	}

	return sections
}

func addTableOfContents(m core.Maroto, sections []reportSection, sectionPages []int) {
	m.AddRows(row.New(6).Add(col.New(12)))
	m.AddRows(
		row.New(10).WithStyle(&props.Cell{BackgroundColor: colorPtr(colorHeaderBG)}).Add(
			text.NewCol(12, "Table of Contents", props.Text{
				Size:  16,
				Style: fontstyle.Bold,
				Color: colorPtr(colorTealDark),
				Top:   2,
				Left:  2,
			}),
		),
	)
	m.AddRows(row.New(4).Add(col.New(12)))

	headers := []string{"Section", "Page"}
	colWidths := []int{10, 2}
	var rows [][]tableCell
	for i, section := range sections {
		pageLabel := "—"
		if i < len(sectionPages) && sectionPages[i] > 0 {
			pageLabel = fmt.Sprintf("%d", sectionPages[i])
		}
		rows = append(rows, []tableCell{
			{text: fmt.Sprintf("%d. %s", i+1, section.title)},
			{text: pageLabel, bold: true, color: colorPtr(colorTealDark)},
		})
	}
	addTable(m, colWidths, headers, rows)

	m.AddRows(row.New(6).Add(col.New(12)))
	m.AddRows(
		text.NewRow(5, "Open the PDF bookmarks panel (sidebar) to jump directly to any section.", props.Text{
			Size:  8,
			Style: fontstyle.Italic,
			Color: colorPtr(colorMuted),
		}),
	)
}

func addCoverPage(m core.Maroto, result *models.ScanResult) {
	m.AddRows(row.New(12).Add(col.New(12)))

	if len(logoPNG) > 0 {
		m.AddRows(
			image.NewFromBytesRow(coverLogoHeight, logoPNG, extension.Png, props.Rect{
				Center:  true,
				Percent: coverLogoPercent,
			}),
		)
		m.AddRows(row.New(4).Add(col.New(12)))
	}

	m.AddRows(
		text.NewRow(14, "EchoState Reconnaissance Report", props.Text{
			Size:  22,
			Style: fontstyle.Bold,
			Align: align.Center,
		}),
	)

	m.AddRows(row.New(8).Add(col.New(12)))

	scannedAt := result.ScannedAt.Format(pdfDateFormat)
	if result.ScannedAt.IsZero() {
		scannedAt = "N/A"
	}

	target := valueOrNA(result.Host)
	targetLink := hyperlinkURL(targetHTTPSURL(result.Host))
	targetProps := props.Text{Size: 12, Align: align.Center}
	if targetLink != nil {
		targetProps.Hyperlink = targetLink
		targetProps.Color = colorPtr(colorLink)
	}
	m.AddRows(text.NewRow(8, "Target: "+target, targetProps))
	m.AddRows(
		text.NewRow(8, fmt.Sprintf("Scanned At: %s", scannedAt), props.Text{
			Size:  12,
			Align: align.Center,
		}),
	)

	m.AddRows(row.New(16).Add(col.New(12)))

	projectURL := reportProjectURL
	m.AddRows(
		text.NewRow(6, "Passive reconnaissance intelligence report", props.Text{
			Size:  9,
			Align: align.Center,
			Style: fontstyle.Italic,
		}),
		text.NewRow(6, "Generated by EchoState", props.Text{
			Size:  8,
			Align: align.Center,
			Style: fontstyle.Italic,
			Color: colorPtr(colorMuted),
		}),
		text.NewRow(6, reportProjectURL, props.Text{
			Size:      8,
			Align:     align.Center,
			Style:     fontstyle.Italic,
			Color:     colorPtr(colorLink),
			Hyperlink: &projectURL,
		}),
	)
}

func addReportMetadata(m core.Maroto, result *models.ScanResult, clientIP string) {
	addKeyValueRow(m, "Report Generated At", time.Now().UTC().Format(pdfDateFormat))
	addKeyValueRow(m, "Target", valueOrNA(result.Host))

	scannedAt := result.ScannedAt.Format(pdfDateFormat)
	if result.ScannedAt.IsZero() {
		scannedAt = "N/A"
	}
	addKeyValueRow(m, "Scanned At", scannedAt)

	if strings.TrimSpace(clientIP) != "" && clientIP != "unknown" {
		addKeyValueRow(m, "Client IP", clientIP)
	}
}

func addSecurityFindings(m core.Maroto, findings []securityFinding) {
	if len(findings) == 0 {
		m.AddRows(text.NewRow(6, "No security findings available.", props.Text{Size: 10}))
		return
	}

	headers := []string{"Severity", "Finding", "Detail"}
	colWidths := []int{2, 5, 5}
	var rows [][]tableCell
	for _, finding := range findings {
		rows = append(rows, []tableCell{
			{text: strings.ToUpper(finding.severity), color: colorForSeverity(finding.severity), bold: true},
			{text: finding.summary},
			{text: finding.detail, color: colorPtr(colorMuted)},
		})
	}
	addTable(m, colWidths, headers, rows)
}

func addIntelHighlights(m core.Maroto, result *models.ScanResult, pwhois *PWhoisInfo) {
	addKeyValueRow(m, "Resolved IP", stringVal(result.ASN, "ip"))
	addKeyValueRow(m, "ASN", stringVal(result.ASN, "asn"))
	addKeyValueRow(m, "AS Name", stringVal(result.ASN, "as_name"))
	addKeyValueRow(m, "Prefix", firstNonNA(stringVal(result.ASN, "prefix"), pwhoisPrefix(pwhois)))
	addKeyValueRow(m, "Country", firstNonNA(stringVal(result.ASN, "country"), pwhoisCountry(pwhois)))
	addKeyValueRow(m, "Registrar", stringVal(result.WHOIS, "registrar"))
	addKeyValueRow(m, "Web Title", stringVal(result.Web, "title"))

	certDays := stringVal(result.TLS, "days_remaining")
	addKeyValueRowStyled(m, "Cert Expires", stringVal(result.TLS, "not_after"), colorForCertDays(certDays), nil)
	addKeyValueRowStyled(m, "Days Remaining", certDays, colorForCertDays(certDays), nil)

	addKeyValueRow(m, "JARM", stringVal(result.TLS, "jarm"))
	addKeyValueRow(m, "JA3S", stringVal(result.TLS, "ja3s"))
	addKeyValueRow(m, "Favicon MMH3", firstNonNA(stringVal(result.Favicon, "mmh3"), stringVal(result.Favicon, "shodan")))
	addKeyValueRow(m, "Cert Issuer", stringVal(result.TLS, "issuer"))
	addKeyValueRow(m, "TLS Version", stringVal(result.TLS, "tls_version"))
	addKeyValueRowStyled(m, "OCSP Stapled", stringVal(result.TLS, "ocsp_stapled"), colorForBoolish(stringVal(result.TLS, "ocsp_stapled"), "true", "yes"), nil)

	dnssecStatus := stringVal(nestedMap(result.DNS, "DNSSEC"), "status")
	addKeyValueRowStyled(m, "DNSSEC", dnssecStatus, colorForBoolish(dnssecStatus, "secure", "signed", "valid"), nil)

	addKeyValueRow(m, "CT Certificates", ctCertificateSummary(result.CT))
	addKeyValueRow(m, "HSTS Preload", hstsPreloadSummary(result.Web))
	addKeyValueRow(m, "Cookie Names", cookieNameSummary(result.Web))
	addKeyValueRow(m, "DNS A", stringVal(result.DNS, "A"))
	addKeyValueRow(m, "DNS AAAA", stringVal(result.DNS, "AAAA"))
	addKeyValueRow(m, "PTR", stringVal(result.DNS, "PTR"))
	addKeyValueRow(m, "CDN Provider", stringVal(nestedMap(result.DNS, "INFRA_LABELS"), "cdn_provider"))
	addKeyValueRow(m, "Mail Provider", stringVal(nestedMap(result.DNS, "INFRA_LABELS"), "mail_provider"))

	posture := nestedMap(result.DNS, "MAIL_POSTURE")
	mailSummary := mailPostureSummary(result.DNS)
	addKeyValueRowStyled(m, "Mail Posture", mailSummary, colorForMailGrade(stringVal(posture, "grade")), nil)

	addKeyValueRow(m, "BIMI", bimiSummary(result.DNS))

	hijackRisk := stringVal(nestedMap(result.ASN, "routing"), "hijack_risk")
	addKeyValueRowStyled(m, "Hijack Risk", hijackRisk, colorForHijackRisk(hijackRisk), nil)

	stability := stringVal(nestedMap(nestedMap(result.ASN, "routing"), "path_profile"), "stability")
	addKeyValueRowStyled(m, "BGP Path Stability", stability, colorForBoolish(stability, "stable", "consistent"), nil)

	addKeyValueRow(m, "CT Subdomains", ctCount(result.CT))
	addKeyValueRow(m, "Sitemap Paths", crawlPathSummary(result.Crawl))
	addKeyValueRow(m, "Security.txt", securityTxtSummary(result.Crawl))
	addKeyValueRow(m, "MTA-STS", stringVal(nestedMap(result.DNS, "MTA_STS"), "mode"))
	addKeyValueRow(m, "CAA Records", caaCount(result.DNS))
	addKeyValueRow(m, "Redirects", redirectSummaryFromWeb(result.Web))
	addKeyValueRow(m, "Storage Buckets", storageBucketSummary(result.Storage))
	addKeyValueRow(m, "Provider Hints", providerHintSummary(result.Storage))
	addKeyValueRow(m, "Meta Description", stringVal(nestedMap(result.Web, "meta"), "description"))
}

func crawlPathSummary(data map[string]any) string {
	if len(data) == 0 {
		return "N/A"
	}
	if count := stringVal(data, "sitemap_url_count"); count != "N/A" {
		return count + " URLs"
	}
	if urls := stringSlice(data, "sitemap_urls"); len(urls) > 0 {
		return fmt.Sprintf("%d URLs", len(urls))
	}
	robots := nestedMap(data, "robots")
	if len(robots) > 0 {
		paths := len(stringSlice(robots, "disallow")) + len(stringSlice(robots, "allow"))
		if paths > 0 {
			return fmt.Sprintf("%d robots rules", paths)
		}
		if len(stringSlice(robots, "sitemaps")) > 0 {
			return "robots.txt only"
		}
	}
	return "N/A"
}

func storageBucketSummary(data map[string]any) string {
	buckets := mapSlice(data, "buckets")
	if len(buckets) == 0 {
		return "N/A"
	}
	return fmt.Sprintf("%d", len(buckets))
}

func securityTxtSummary(data map[string]any) string {
	securityTxt := nestedMap(data, "security_txt")
	if len(securityTxt) == 0 {
		return "N/A"
	}
	if contacts := stringSlice(securityTxt, "contacts"); len(contacts) > 0 {
		return contacts[0]
	}
	if expires := stringVal(securityTxt, "expires"); expires != "N/A" {
		return "expires " + expires
	}
	return "present"
}

func caaCount(data map[string]any) string {
	records := mapSlice(data, "CAA")
	if len(records) == 0 {
		return "N/A"
	}
	return fmt.Sprintf("%d", len(records))
}

func redirectSummaryFromWeb(data map[string]any) string {
	chain := mapSlice(data, "redirect_chain")
	if len(chain) <= 1 {
		return "N/A"
	}
	return fmt.Sprintf("%d hops", len(chain)-1)
}

func bimiSummary(data map[string]any) string {
	bimi := nestedMap(data, "BIMI")
	if len(bimi) == 0 {
		return "N/A"
	}
	if record := stringVal(bimi, "record"); record != "N/A" {
		return "published"
	}
	return "present"
}

func ctCertificateSummary(data map[string]any) string {
	if count := stringVal(data, "certificate_count"); count != "N/A" {
		return count
	}
	certs := mapSlice(data, "certificates")
	if len(certs) == 0 {
		return "N/A"
	}
	return fmt.Sprintf("%d", len(certs))
}

func hstsPreloadSummary(data map[string]any) string {
	preload := nestedMap(data, "hsts_preload")
	if len(preload) == 0 {
		return "N/A"
	}
	if preload["preloaded"] == true {
		return "preloaded"
	}
	return stringVal(preload, "preload_status")
}

func cookieNameSummary(data map[string]any) string {
	names := stringSlice(data, "cookie_names")
	if len(names) == 0 {
		return "N/A"
	}
	return fmt.Sprintf("%d", len(names))
}

func providerHintSummary(data map[string]any) string {
	hints := mapSlice(data, "provider_hints")
	if len(hints) == 0 {
		return "N/A"
	}
	return fmt.Sprintf("%d", len(hints))
}

func addPWhoisSection(m core.Maroto, info *PWhoisInfo, data map[string]any) {
	if len(data) == 0 && (info == nil || !info.HasData()) {
		return
	}

	if info != nil && info.HasData() {
		addKeyValueRow(m, "Origin AS", valueOrNA(info.OriginAS))
		addKeyValueRow(m, "Organization", valueOrNA(info.OrgName))
		addKeyValueRow(m, "Country", valueOrNA(info.CountryCode))
		addKeyValueRow(m, "City", valueOrNA(info.City))
		addKeyValueRow(m, "Prefix", valueOrNA(info.Prefix))
		if info.LookedUpAt != nil && !info.LookedUpAt.IsZero() {
			addKeyValueRow(m, "Looked Up At", info.LookedUpAt.UTC().Format(pdfDateFormat))
		}
	}

	if len(data) > 0 {
		addSubheader(m, "pWhois Record")
		for _, key := range sortedMapKeys(data) {
			addKeyValueRow(m, formatFieldLabel(key), formatScalar(data[key]))
		}
	}
}

func addChanges(m core.Maroto, changes []string, details []models.ChangeDetail) {
	if len(changes) == 0 && len(details) == 0 {
		return
	}

	if len(details) > 0 {
		addSubheader(m, "Structured changes")
		headers := []string{"Severity", "Change"}
		colWidths := []int{2, 10}
		var rows [][]tableCell
		for i, detail := range details {
			if i >= maxChangeDetails {
				m.AddRows(text.NewRow(5, fmt.Sprintf("... and %d more change(s)", len(details)-maxChangeDetails), props.Text{
					Size:  9,
					Style: fontstyle.Italic,
				}))
				break
			}
			line := fmt.Sprintf("[%s] %s", detail.Type, detail.Summary)
			if detail.Field != "" {
				line = fmt.Sprintf("%s (%s)", line, detail.Field)
			}
			rows = append(rows, []tableCell{
				{text: strings.ToUpper(detail.Severity), color: colorForSeverity(detail.Severity), bold: true},
				{text: line},
			})
		}
		if len(rows) > 0 {
			addTable(m, colWidths, headers, rows)
		}
	}

	if len(changes) > 0 {
		addSubheader(m, "Summary")
		addBulletItems(m, changes, maxChangesShown)
	}
}

func addScreenshot(m core.Maroto, jpeg []byte, meta map[string]any) {
	if len(jpeg) == 0 && len(meta) == 0 {
		return
	}

	if len(jpeg) > 0 {
		m.AddRows(
			image.NewFromBytesRow(screenshotRowHeight, jpeg, extension.Jpeg, props.Rect{
				Center:  true,
				Percent: 80,
			}),
		)
	}

	if len(meta) > 0 {
		screenshotURL := stringVal(meta, "url")
		addKeyValueRowStyled(m, "URL", screenshotURL, nil, hyperlinkURL(screenshotURL))
		addKeyValueRow(m, "Captured At", stringVal(meta, "captured_at"))
		addKeyValueRow(m, "Dimensions", screenshotDimensions(meta))
		addKeyValueRow(m, "Format", stringVal(meta, "format"))
		if errMsg := stringVal(meta, "error"); errMsg != "N/A" {
			addKeyValueRow(m, "Capture Error", errMsg)
		}
	} else if len(jpeg) == 0 {
		m.AddRows(text.NewRow(6, "No screenshot available.", props.Text{Size: 10}))
	}
}

func addSectionHeader(m core.Maroto, title string) {
	m.AddRows(row.New(4).Add(col.New(12)))
	m.AddRows(
		row.New(9).WithStyle(&props.Cell{BackgroundColor: colorPtr(colorTeal)}).Add(
			text.NewCol(12, title, props.Text{
				Size:  13,
				Style: fontstyle.Bold,
				Color: &props.Color{Red: 255, Green: 255, Blue: 255},
				Top:   2,
				Left:  2,
			}),
		),
	)
	m.AddRows(row.New(2).Add(col.New(12)))
}

func addWHOIS(m core.Maroto, data map[string]any) {
	if len(data) == 0 {
		m.AddRows(text.NewRow(6, "No WHOIS data available.", props.Text{Size: 10}))
		return
	}

	addKeyValueRow(m, "Domain", stringVal(data, "domain"))
	addKeyValueRow(m, "Registrar", stringVal(data, "registrar"))
	addKeyValueRow(m, "Registrant Email", stringVal(data, "registrant_email"))
	addKeyValueRow(m, "Abuse Email", stringVal(data, "abuse_email"))
	addKeyValueRow(m, "Expiration Date", stringVal(data, "expiration_date"))
	addKeyValueRow(m, "Name Servers", stringVal(data, "name_servers"))
	addKeyValueRow(m, "Status", stringVal(data, "status"))
	addKeyValueRow(m, "RDAP Source", stringVal(data, "rdap_source"))

	if rdap := nestedMap(data, "rdap"); len(rdap) > 0 {
		addSubheader(m, "RDAP")
		for _, key := range []string{"domain", "registrar", "registrant_org", "registrant_email", "abuse_email", "expiration_date", "registration_date"} {
			if val := stringVal(rdap, key); val != "N/A" {
				addKeyValueRow(m, formatFieldLabel(key), val)
			}
		}
	}

}

func addASN(m core.Maroto, data map[string]any) {
	if len(data) == 0 {
		m.AddRows(text.NewRow(6, "No ASN/BGP data available.", props.Text{Size: 10}))
		return
	}

	addKeyValueRow(m, "ASN", stringVal(data, "asn"))
	addKeyValueRow(m, "AS Name", stringVal(data, "as_name"))
	addKeyValueRow(m, "IP", stringVal(data, "ip"))
	addKeyValueRow(m, "Prefix", stringVal(data, "prefix"))
	addKeyValueRow(m, "Country", stringVal(data, "country"))
	addKeyValueRow(m, "Registry", stringVal(data, "registry"))
	addKeyValueRow(m, "Allocated", stringVal(data, "allocated"))

	if routing := nestedMap(data, "routing"); len(routing) > 0 {
		addSubheader(m, "BGP Routing")
		hijackRisk := stringVal(routing, "hijack_risk")
		addKeyValueRowStyled(m, "Hijack Risk", hijackRisk, colorForHijackRisk(hijackRisk), nil)
		addKeyValueRow(m, "Visible Origins", stringVal(routing, "visible_origins"))
		addKeyValueRow(m, "First Seen", stringVal(routing, "first_seen"))
		addKeyValueRow(m, "Last Seen", stringVal(routing, "last_seen"))
		if profile := nestedMap(routing, "path_profile"); len(profile) > 0 {
			stability := stringVal(profile, "stability")
			addKeyValueRowStyled(m, "Path Stability", stability, colorForBoolish(stability, "stable", "consistent"), nil)
			addKeyValueRow(m, "Path Count", stringVal(profile, "path_count"))
			addKeyValueRow(m, "Unique Paths", stringVal(profile, "unique_paths"))
		}
		if notes := stringSlice(routing, "notes"); len(notes) > 0 {
			addSubheader(m, "Routing Notes")
			addBulletItems(m, notes, maxErrorsShown)
		}
	}

	if peering := nestedMap(data, "peeringdb"); len(peering) > 0 {
		addSubheader(m, "PeeringDB")
		addKeyValueRow(m, "Network Name", stringVal(peering, "name"))
		addKeyValueRow(m, "Website", stringVal(peering, "website"))
		addKeyValueRow(m, "IX Count", stringVal(peering, "ix_count"))
		addKeyValueRow(m, "Facility Count", stringVal(peering, "fac_count"))
	}
}

func addDNS(m core.Maroto, data map[string]any) {
	if len(data) == 0 {
		m.AddRows(text.NewRow(6, "No DNS data available.", props.Text{Size: 10}))
		return
	}

	addKeyValueRow(m, "A", stringVal(data, "A", "a"))
	addKeyValueRow(m, "AAAA", stringVal(data, "AAAA", "aaaa"))
	addKeyValueRow(m, "PTR", stringVal(data, "PTR", "ptr"))
	addKeyValueRow(m, "MX", stringVal(data, "MX", "mx"))
	addKeyValueRow(m, "NS", stringVal(data, "NS", "ns"))
	addKeyValueRow(m, "TXT", stringVal(data, "TXT", "txt"))
	addKeyValueRow(m, "CNAME", stringVal(data, "CNAME", "cname"))

	if soa := nestedMap(data, "SOA"); len(soa) > 0 {
		addSubheader(m, "SOA")
		addKeyValueRow(m, "Zone", stringVal(soa, "zone"))
		addKeyValueRow(m, "Mname", stringVal(soa, "mname"))
		addKeyValueRow(m, "Rname", stringVal(soa, "rname"))
		addKeyValueRow(m, "Serial", stringVal(soa, "serial"))
		addKeyValueRow(m, "Refresh", stringVal(soa, "refresh"))
		addKeyValueRow(m, "Retry", stringVal(soa, "retry"))
		addKeyValueRow(m, "Expire", stringVal(soa, "expire"))
		addKeyValueRow(m, "Minimum", stringVal(soa, "minimum"))
	}

	addKeyValueRow(m, "DMARC (TXT)", stringVal(data, "DMARC", "dmarc"))
	if dmarc := nestedMap(data, "DMARC_PARSED"); len(dmarc) > 0 {
		addSubheader(m, "DMARC Parsed")
		addKeyValueRow(m, "Policy", stringVal(dmarc, "policy"))
		addKeyValueRow(m, "Subdomain Policy", stringVal(dmarc, "subdomain_policy"))
		addKeyValueRow(m, "Percentage", stringVal(dmarc, "percentage"))
		addKeyValueRow(m, "Aggregate Report URI", stringVal(dmarc, "aggregate_report_uri"))
		addKeyValueRow(m, "Record", stringVal(dmarc, "record"))
	}

	if spf := nestedMap(data, "SPF"); len(spf) > 0 {
		addSubheader(m, "SPF")
		addKeyValueRow(m, "Policy", stringVal(spf, "policy"))
		addKeyValueRow(m, "Record", stringVal(spf, "record"))
	} else {
		addKeyValueRow(m, "SPF", stringVal(data, "SPF", "spf"))
	}

	if caaRecords := mapSlice(data, "CAA"); len(caaRecords) > 0 {
		addSubheader(m, "CAA")
		for i, record := range caaRecords {
			if i >= 10 {
				m.AddRows(text.NewRow(5, fmt.Sprintf("... and %d more CAA record(s)", len(caaRecords)-10), props.Text{
					Size:  9,
					Style: fontstyle.Italic,
				}))
				break
			}
			addKeyValueRow(m, fmt.Sprintf("Record %d", i+1), stringVal(record, "record"))
		}
	}

	if bimi := nestedMap(data, "BIMI"); len(bimi) > 0 {
		addSubheader(m, "BIMI")
		addKeyValueRow(m, "Record", stringVal(bimi, "record"))
		addKeyValueRow(m, "Logo URL", stringVal(bimi, "l"))
		addKeyValueRow(m, "Authority", stringVal(bimi, "a"))
	}

	if infra := nestedMap(data, "INFRA_LABELS"); len(infra) > 0 {
		addSubheader(m, "Infrastructure Labels")
		addKeyValueRow(m, "CDN Provider", stringVal(infra, "cdn_provider"))
		addKeyValueRow(m, "Mail Provider", stringVal(infra, "mail_provider"))
		addKeyValueRow(m, "CNAME Target", stringVal(infra, "cname_target"))
	}

	if mtaSts := nestedMap(data, "MTA_STS"); len(mtaSts) > 0 {
		addSubheader(m, "MTA-STS")
		addKeyValueRow(m, "Mode", stringVal(mtaSts, "mode"))
		addKeyValueRow(m, "Version", stringVal(mtaSts, "version"))
		addKeyValueRow(m, "Max Age", stringVal(mtaSts, "max_age"))
		addKeyValueRow(m, "MX Hosts", stringVal(mtaSts, "mx"))
	}

	if tlsRpt := nestedMap(data, "TLS_RPT"); len(tlsRpt) > 0 {
		addSubheader(m, "TLS-RPT")
		addKeyValueRow(m, "Record", stringVal(tlsRpt, "record"))
		addKeyValueRow(m, "RUA", stringVal(tlsRpt, "rua"))
	}

	if dkimRecords := mapSlice(data, "DKIM"); len(dkimRecords) > 0 {
		addSubheader(m, "DKIM")
		for i, record := range dkimRecords {
			if i >= 5 {
				m.AddRows(text.NewRow(5, fmt.Sprintf("... and %d more DKIM record(s)", len(dkimRecords)-5), props.Text{
					Size:  9,
					Style: fontstyle.Italic,
				}))
				break
			}
			label := fmt.Sprintf("Selector %d", i+1)
			addKeyValueRow(m, label, formatMapLines(record, ""))
		}
	}

	if dnssec := nestedMap(data, "DNSSEC"); len(dnssec) > 0 {
		addSubheader(m, "DNSSEC")
		dnssecStatus := stringVal(dnssec, "status")
		addKeyValueRowStyled(m, "Status", dnssecStatus, colorForBoolish(dnssecStatus, "secure", "signed", "valid"), nil)
		addKeyValueRow(m, "Signed", stringVal(dnssec, "signed"))
		addKeyValueRow(m, "Zone", stringVal(dnssec, "zone"))
		addKeyValueRow(m, "DNSKEY Count", stringVal(dnssec, "dnskey_count"))
		addKeyValueRow(m, "DS Count", stringVal(dnssec, "ds_count"))
		addKeyValueRow(m, "Authentic Data", stringVal(dnssec, "authentic_data"))
	}

	if posture := nestedMap(data, "MAIL_POSTURE"); len(posture) > 0 {
		addSubheader(m, "Mail Posture")
		grade := stringVal(posture, "grade")
		score := stringVal(posture, "score")
		addKeyValueRowStyled(m, "Grade", grade, colorForMailGrade(grade), nil)
		addKeyValueRowStyled(m, "Score", score, colorForMailGrade(grade), nil)
		if findings := stringSlice(posture, "findings"); len(findings) > 0 {
			addBulletItems(m, findings, maxErrorsShown)
		}
	}
}

func addTLS(m core.Maroto, data map[string]any) {
	if len(data) == 0 {
		m.AddRows(text.NewRow(6, "No TLS data available.", props.Text{Size: 10}))
		return
	}

	addKeyValueRow(m, "Subject", stringVal(data, "subject"))
	addKeyValueRow(m, "Issuer", stringVal(data, "issuer"))
	addKeyValueRow(m, "DNS Names", stringVal(data, "dns_names"))
	addKeyValueRow(m, "IP SANs", stringVal(data, "ip_sans"))
	addKeyValueRow(m, "Not Before", stringVal(data, "not_before"))
	addKeyValueRow(m, "Not After", stringVal(data, "not_after"))
	daysRemaining := stringVal(data, "days_remaining")
	addKeyValueRowStyled(m, "Days Remaining", daysRemaining, colorForCertDays(daysRemaining), nil)
	expired := stringVal(data, "expired")
	var expiredColor *props.Color
	if strings.EqualFold(expired, "true") {
		expiredColor = colorPtr(colorRed)
	} else if expired != "N/A" {
		expiredColor = colorPtr(colorGreen)
	}
	addKeyValueRowStyled(m, "Expired", expired, expiredColor, nil)
	addKeyValueRow(m, "Chain Length", stringVal(data, "chain_length"))
	addKeyValueRow(m, "Serial", stringVal(data, "serial"))
	addKeyValueRow(m, "TLS Version", stringVal(data, "tls_version"))
	addKeyValueRow(m, "Negotiated Cipher", stringVal(data, "negotiated_cipher"))
	addKeyValueRow(m, "OCSP Stapled", stringVal(data, "ocsp_stapled"))
	addKeyValueRow(m, "OCSP Response Bytes", stringVal(data, "ocsp_response_bytes"))
	addKeyValueRow(m, "Cert Version", stringVal(data, "version"))
	addKeyValueRow(m, "Signature Algorithm", stringVal(data, "signature_algorithm"))
	addKeyValueRow(m, "Cipher Suite", stringVal(data, "cipher_suite"))
	addKeyValueRow(m, "JARM", stringVal(data, "jarm"))
	addKeyValueRow(m, "JA3S", stringVal(data, "ja3s"))

	if chain := mapSlice(data, "chain"); len(chain) > 0 {
		addSubheader(m, "Certificate Chain")
		for i, cert := range chain {
			addKeyValueRow(m, fmt.Sprintf("Cert %d Subject", i+1), stringVal(cert, "subject"))
			addKeyValueRow(m, fmt.Sprintf("Cert %d Issuer", i+1), stringVal(cert, "issuer"))
		}
	}
}

func addWeb(m core.Maroto, data map[string]any) {
	if len(data) == 0 {
		m.AddRows(text.NewRow(6, "No web data available.", props.Text{Size: 10}))
		return
	}

	addKeyValueRow(m, "Title", stringVal(data, "title"))
	pageURL := stringVal(data, "url")
	addKeyValueRowStyled(m, "URL", pageURL, nil, hyperlinkURL(pageURL))
	headerURL := stringVal(data, "header_url")
	addKeyValueRowStyled(m, "Header URL", headerURL, nil, hyperlinkURL(headerURL))

	if chain := mapSlice(data, "redirect_chain"); len(chain) > 0 {
		addSubheader(m, "Redirect Chain")
		addRedirectChainTable(m, chain)
	}

	if emails := stringSlice(data, "contact_emails"); len(emails) > 0 {
		addSubheader(m, "Contact Emails")
		addBulletItems(m, emails, maxErrorsShown)
	}

	if headers := nestedMap(data, "headers"); len(headers) > 0 {
		addSubheader(m, "Response Headers")
		addKeyValueRow(m, "Headers", formatMapLines(headers, ""))
	}

	if security := nestedMap(data, "security_headers"); len(security) > 0 {
		addSubheader(m, "Security Headers")
		addKeyValueRow(m, "Security Headers", formatMapLines(security, ""))
	}

	if tech := stringSlice(data, "tech_stack"); len(tech) > 0 {
		addSubheader(m, "Tech Stack")
		addBulletItems(m, tech, maxTechStackItems)
	}

	if assets := mapSlice(data, "js_assets"); len(assets) > 0 {
		addSubheader(m, "JavaScript Assets")
		addJSAssetsTable(m, assets)
	}

	if plugins := mapSlice(data, "wordpress_plugins"); len(plugins) > 0 {
		addSubheader(m, "WordPress Plugins")
		addBulletItems(m, formatWordPressItems(plugins), maxTechStackItems)
	}

	if themes := mapSlice(data, "wordpress_themes"); len(themes) > 0 {
		addSubheader(m, "WordPress Themes")
		addBulletItems(m, formatWordPressItems(themes), maxTechStackItems)
	}

	if preload := nestedMap(data, "hsts_preload"); len(preload) > 0 {
		addSubheader(m, "HSTS Preload")
		addKeyValueRow(m, "Status", stringVal(preload, "status"))
		addKeyValueRow(m, "Preload Status", stringVal(preload, "preload_status"))
		addKeyValueRow(m, "Preloaded", stringVal(preload, "preloaded"))
	}

	if names := stringSlice(data, "cookie_names"); len(names) > 0 {
		addSubheader(m, "Cookie Names")
		addBulletItems(m, names, maxErrorsShown)
	}

	if meta := nestedMap(data, "meta"); len(meta) > 0 {
		addSubheader(m, "Meta / Open Graph")
		for _, key := range sortedMapKeys(meta) {
			addKeyValueRow(m, formatFieldLabel(key), formatScalar(meta[key]))
		}
	}

	if hints := mapSlice(data, "provider_hints"); len(hints) > 0 {
		addSubheader(m, "Provider Hints")
		var lines []string
		for i, hint := range hints {
			if i >= maxTechStackItems {
				lines = append(lines, fmt.Sprintf("... and %d more hint(s)", len(hints)-maxTechStackItems))
				break
			}
			provider := stringVal(hint, "provider")
			pattern := stringVal(hint, "pattern")
			if provider != "N/A" && pattern != "N/A" {
				lines = append(lines, fmt.Sprintf("%s (%s)", provider, pattern))
			} else {
				lines = append(lines, formatMapLines(hint, ""))
			}
		}
		addBulletItems(m, lines, 0)
	}

	copyrights := extractCopyrights(data)
	if len(copyrights) == 0 {
		addKeyValueRow(m, "Copyrights", "None detected")
	} else {
		addSubheader(m, "Copyrights")
		addBulletItems(m, copyrights, maxCopyrights)
	}
}

func addFavicon(m core.Maroto, data map[string]any) {
	if len(data) == 0 {
		m.AddRows(text.NewRow(6, "No favicon data available.", props.Text{Size: 10}))
		return
	}

	addKeyValueRow(m, "URL", stringVal(data, "url"))
	addKeyValueRow(m, "MMH3", stringVal(data, "mmh3"))
	addKeyValueRow(m, "Shodan Hash", stringVal(data, "shodan"))
	addKeyValueRow(m, "SHA256", stringVal(data, "sha256"))
	addKeyValueRow(m, "Size", stringVal(data, "size"))
}

func addCrawl(m core.Maroto, data map[string]any) {
	if len(data) == 0 {
		m.AddRows(text.NewRow(6, "No crawl data available.", props.Text{Size: 10}))
		return
	}

	for _, item := range []struct {
		title string
		key   string
	}{
		{title: "humans.txt", key: "humans_txt"},
		{title: "ads.txt", key: "ads_txt"},
		{title: "app-ads.txt", key: "app_ads_txt"},
	} {
		if file := nestedMap(data, item.key); len(file) > 0 {
			addSubheader(m, item.title)
			addKeyValueRow(m, "Source", stringVal(file, "source"))
			if lines := stringSlice(file, "lines"); len(lines) > 0 {
				addBulletItems(m, lines, maxCrawlPaths)
			}
		}
	}

	if securityTxt := nestedMap(data, "security_txt"); len(securityTxt) > 0 {
		addSubheader(m, "security.txt")
		addKeyValueRow(m, "Source", stringVal(securityTxt, "source"))
		if contacts := stringSlice(securityTxt, "contacts"); len(contacts) > 0 {
			addKeyValueRow(m, "Contacts", strings.Join(contacts, ", "))
		}
		addKeyValueRow(m, "Expires", stringVal(securityTxt, "expires"))
		addKeyValueRow(m, "Canonical", stringVal(securityTxt, "canonical"))
		if policies := stringSlice(securityTxt, "policies"); len(policies) > 0 {
			addKeyValueRow(m, "Policies", strings.Join(policies, ", "))
		}
	}

	if robots := nestedMap(data, "robots"); len(robots) > 0 {
		addSubheader(m, "robots.txt")
		if disallow := stringSlice(robots, "disallow"); len(disallow) > 0 {
			addKeyValueRow(m, "Disallow", strings.Join(disallow, ", "))
		}
		if allow := stringSlice(robots, "allow"); len(allow) > 0 {
			addKeyValueRow(m, "Allow", strings.Join(allow, ", "))
		}
		if sitemaps := stringSlice(robots, "sitemaps"); len(sitemaps) > 0 {
			addKeyValueRow(m, "Sitemaps", strings.Join(sitemaps, ", "))
		}
	}

	if sitemapURLs := stringSlice(data, "sitemap_urls"); len(sitemapURLs) > 0 {
		addSubheader(m, "Sitemap URLs")
		addBulletItems(m, sitemapURLs, maxCrawlPaths)
	}

	addKeyValueRow(m, "Sitemap URL Count", stringVal(data, "sitemap_url_count"))

	if crawlErrors := stringSlice(data, "errors"); len(crawlErrors) > 0 {
		addSubheader(m, "Crawl Errors")
		addBulletItems(m, crawlErrors, maxErrorsShown)
	}

	if stringSlice(data, "sitemap_urls") == nil && stringVal(data, "sitemap_url_count") == "N/A" {
		if robots := nestedMap(data, "robots"); len(robots) > 0 && len(stringSlice(robots, "sitemaps")) > 0 {
			m.AddRows(text.NewRow(5, "Sitemap URLs declared in robots.txt but page URLs could not be fetched.", props.Text{
				Size:  9,
				Style: fontstyle.Italic,
			}))
		}
	}
}

func addEnrichment(m core.Maroto, data map[string]any) {
	if len(data) == 0 {
		return
	}

	status := stringVal(data, "status")
	if status == "pending" {
		m.AddRows(text.NewRow(6, "Enrichment was still running when this report was generated. Re-generate the report after enrichment completes for full Shodan/Censys/HIBP data.", props.Text{
			Size:  9,
			Style: fontstyle.Italic,
		}))
		return
	}

	if completed := stringVal(data, "completed_at"); completed != "N/A" {
		addKeyValueRow(m, "Completed At", completed)
	}
	if status != "N/A" {
		addKeyValueRow(m, "Status", status)
	}

	for _, key := range sortedMapKeys(data) {
		if key == "status" || key == "completed_at" || key == "started_at" {
			continue
		}
		if strings.HasSuffix(key, "_error") {
			addKeyValueRow(m, formatFieldLabel(strings.TrimSuffix(key, "_error"))+" Error", formatScalar(data[key]))
			continue
		}

		block, ok := data[key].(map[string]any)
		if !ok {
			addKeyValueRow(m, formatFieldLabel(key), formatScalar(data[key]))
			continue
		}

		addSubheader(m, formatFieldLabel(key))
		addEnrichmentProvider(m, key, block)
	}
}

func addEnrichmentProvider(m core.Maroto, provider string, block map[string]any) {
	switch provider {
	case "shodan":
		if host, ok := block["host"].(map[string]any); ok {
			addKeyValueRow(m, "Resolved IP", stringVal(host, "ip"))
			addKeyValueRow(m, "Organization", stringVal(host, "org"))
			addKeyValueRow(m, "ISP", stringVal(host, "isp"))
			addKeyValueRow(m, "ASN", stringVal(host, "asn"))
			if ports := formatPortList(host["ports"]); ports != "" {
				addKeyValueRow(m, "Open Ports", ports)
			}
			if count := stringVal(host, "vuln_count"); count != "N/A" {
				addKeyValueRow(m, "Known Vulns", count)
			}
			if vulns := stringSlice(host, "vulns"); len(vulns) > 0 {
				addSubheader(m, "Vulnerability IDs")
				addBulletItems(m, vulns, maxErrorsShown)
			}
			if services := mapSlice(host, "services"); len(services) > 0 {
				addSubheader(m, "Banner Services")
				addBulletItems(m, formatEnrichmentServices(services, "shodan"), 0)
			}
		}
		if search, ok := block["favicon_search"].(map[string]any); ok {
			addKeyValueRow(m, "Favicon Hash", stringVal(search, "query"))
			addKeyValueRow(m, "Matching Hosts", stringVal(search, "total"))
			addBulletItems(m, formatShodanFaviconMatches(search), maxErrorsShown)
		}
	case "censys":
		if host, ok := block["host"].(map[string]any); ok {
			addKeyValueRow(m, "Resolved IP", stringVal(host, "ip"))
			addKeyValueRow(m, "ASN", stringVal(host, "asn"))
			addKeyValueRow(m, "AS Name", stringVal(host, "as_name"))
			if services := mapSlice(host, "services"); len(services) > 0 {
				addSubheader(m, "Services")
				addBulletItems(m, formatEnrichmentServices(services, "censys"), 0)
			}
		}
		if search, ok := block["jarm_search"].(map[string]any); ok {
			addKeyValueRow(m, "JARM Fingerprint", stringVal(search, "query"))
			addKeyValueRow(m, "Matching Hosts", censysSearchTotal(search))
			addBulletItems(m, formatCensysJARMHits(search), maxErrorsShown)
		}
	case "hibp":
		addKeyValueRow(m, "Emails Checked", stringVal(block, "checked"))
		if results := mapSlice(block, "results"); len(results) > 0 {
			var lines []string
			for _, item := range results {
				email := stringVal(item, "email")
				if stringVal(item, "pwned") == "true" {
					lines = append(lines, email+" (breached)")
				} else {
					lines = append(lines, email+" (clean)")
				}
			}
			addBulletItems(m, lines, maxErrorsShown)
		}
	case "riskiq":
		addKeyValueRow(m, "Query", stringVal(block, "query"))
		addKeyValueRow(m, "Total Records", stringVal(block, "total"))
	case "wayback":
		addKeyValueRow(m, "Domain", stringVal(block, "query"))
		addKeyValueRow(m, "Archived URLs", stringVal(block, "total"))
		if urls := mapSlice(block, "urls"); len(urls) > 0 {
			var lines []string
			for i, item := range urls {
				if i >= maxCrawlPaths {
					lines = append(lines, fmt.Sprintf("... and %d more URL(s)", len(urls)-maxCrawlPaths))
					break
				}
				lines = append(lines, fmt.Sprintf("%s %s", stringVal(item, "timestamp"), stringVal(item, "url")))
			}
			addBulletItems(m, lines, 0)
		}
	case "virustotal":
		addKeyValueRow(m, "Domain", stringVal(block, "query"))
		addKeyValueRow(m, "Passive DNS Records", stringVal(block, "total"))
		if records := mapSlice(block, "records"); len(records) > 0 {
			var lines []string
			for i, item := range records {
				if i >= maxErrorsShown {
					lines = append(lines, fmt.Sprintf("... and %d more record(s)", len(records)-maxErrorsShown))
					break
				}
				lines = append(lines, fmt.Sprintf("%s → %s", stringVal(item, "date"), stringVal(item, "ip")))
			}
			addBulletItems(m, lines, 0)
		}
	default:
		for _, field := range sortedMapKeys(block) {
			addKeyValueRow(m, formatFieldLabel(field), formatScalar(block[field]))
		}
	}
}

func addStorage(m core.Maroto, data map[string]any) {
	if len(data) == 0 {
		m.AddRows(text.NewRow(6, "No storage hints available.", props.Text{Size: 10}))
		return
	}

	buckets := mapSlice(data, "buckets")
	hints := mapSlice(data, "provider_hints")
	if len(buckets) == 0 && len(hints) == 0 {
		m.AddRows(text.NewRow(6, "No cloud storage buckets detected.", props.Text{Size: 10}))
		return
	}

	if len(hints) > 0 {
		addSubheader(m, "Provider Hints")
		var hintLines []string
		for i, hint := range hints {
			if i >= maxTechStackItems {
				hintLines = append(hintLines, fmt.Sprintf("... and %d more hint(s)", len(hints)-maxTechStackItems))
				break
			}
			provider := stringVal(hint, "provider")
			pattern := stringVal(hint, "pattern")
			if provider != "N/A" && pattern != "N/A" {
				hintLines = append(hintLines, fmt.Sprintf("%s (%s)", provider, pattern))
			} else {
				hintLines = append(hintLines, formatMapLines(hint, ""))
			}
		}
		addBulletItems(m, hintLines, 0)
	}

	if len(buckets) == 0 {
		return
	}

	addSubheader(m, "Detected Buckets")
	var lines []string
	for i, bucket := range buckets {
		if i >= maxBuckets {
			lines = append(lines, fmt.Sprintf("... and %d more bucket(s)", len(buckets)-maxBuckets))
			break
		}
		provider := stringVal(bucket, "provider")
		name := stringVal(bucket, "name")
		url := stringVal(bucket, "url")
		switch {
		case provider != "N/A" && name != "N/A":
			lines = append(lines, fmt.Sprintf("%s: %s", provider, name))
		case url != "N/A":
			lines = append(lines, url)
		default:
			lines = append(lines, formatMapLines(bucket, ""))
		}
	}
	addBulletItems(m, lines, 0)
}

func addCT(m core.Maroto, data map[string]any) {
	if len(data) == 0 {
		m.AddRows(text.NewRow(6, "No certificate transparency data available.", props.Text{Size: 10}))
		return
	}

	if skipped, ok := data["skipped"].(string); ok && strings.TrimSpace(skipped) != "" {
		addKeyValueRow(m, "Skipped", skipped)
		return
	}

	addKeyValueRow(m, "Domain", stringVal(data, "domain"))
	addKeyValueRow(m, "Source", stringVal(data, "source"))
	addKeyValueRow(m, "Count", ctCount(data))

	if subdomains := stringSlice(data, "subdomains"); len(subdomains) > 0 {
		addSubheader(m, "Subdomains")
		addBulletItems(m, subdomains, maxCtSubdomainsShown)
	}

	if certs := mapSlice(data, "certificates"); len(certs) > 0 {
		addSubheader(m, "Certificate Metadata")
		for i, cert := range certs {
			if i >= 10 {
				m.AddRows(text.NewRow(5, fmt.Sprintf("... and %d more certificate(s)", len(certs)-10), props.Text{
					Size:  9,
					Style: fontstyle.Italic,
				}))
				break
			}
			line := fmt.Sprintf("%s · %s → %s", stringVal(cert, "common_name"), stringVal(cert, "issuer"), stringVal(cert, "not_after"))
			addKeyValueRow(m, fmt.Sprintf("Cert %d", i+1), line)
		}
	}
}

func addRedirectChainTable(m core.Maroto, chain []map[string]any) {
	headers := []string{"#", "URL", "Status"}
	colWidths := []int{1, 8, 3}
	var rows [][]tableCell
	for i, hop := range chain {
		if i >= maxCrawlPaths {
			break
		}
		url := stringVal(hop, "url")
		status := stringVal(hop, "status")
		if status == "N/A" {
			status = "—"
		}
		rows = append(rows, []tableCell{
			{text: fmt.Sprintf("%d", i+1)},
			{text: truncate(url, 90), link: hyperlinkURL(url)},
			{text: status},
		})
	}
	addTable(m, colWidths, headers, rows)
}

func addJSAssetsTable(m core.Maroto, assets []map[string]any) {
	headers := []string{"URL", "Hint"}
	colWidths := []int{8, 4}
	var rows [][]tableCell
	for i, asset := range assets {
		if i >= maxJSAssets {
			m.AddRows(text.NewRow(5, fmt.Sprintf("... and %d more asset(s)", len(assets)-maxJSAssets), props.Text{
				Size:  9,
				Style: fontstyle.Italic,
			}))
			break
		}
		url := stringVal(asset, "url")
		hint := stringVal(asset, "hint")
		if hint == "N/A" {
			hint = "—"
		}
		rows = append(rows, []tableCell{
			{text: truncate(url, 72), link: hyperlinkURL(url)},
			{text: hint, color: colorPtr(colorMuted)},
		})
	}
	if len(rows) > 0 {
		addTable(m, colWidths, headers, rows)
	}
}

func addTracerouteHopTable(m core.Maroto, hops []map[string]any) {
	headers := []string{"Hop", "IP", "Host", "RTT", "Location"}
	colWidths := []int{1, 3, 3, 2, 3}
	var rows [][]tableCell
	for i, hop := range hops {
		if i >= maxTracerouteHops {
			m.AddRows(text.NewRow(5, fmt.Sprintf("... and %d more hop(s)", len(hops)-maxTracerouteHops), props.Text{
				Size:  9,
				Style: fontstyle.Italic,
			}))
			break
		}
		host := stringVal(hop, "host", "hostname")
		if host == "N/A" {
			host = "—"
		}
		rows = append(rows, []tableCell{
			{text: stringVal(hop, "hop", "number", "ttl"), bold: true},
			{text: tracerouteHopIP(hop)},
			{text: host},
			{text: formatHopRTT(hop)},
			{text: tracerouteLocation(hop), color: colorPtr(colorMuted)},
		})
	}
	if len(rows) > 0 {
		addTable(m, colWidths, headers, rows)
	}
}

func addTraceroute(m core.Maroto, data map[string]any) {
	if len(data) == 0 {
		m.AddRows(text.NewRow(6, "No traceroute data available.", props.Text{Size: 10}))
		return
	}

	if skipped, ok := data["skipped"].(string); ok && strings.TrimSpace(skipped) != "" {
		addKeyValueRow(m, "Skipped", skipped)
		return
	}

	addKeyValueRow(m, "Destination", stringVal(data, "destination"))
	addKeyValueRow(m, "Hop Count", stringVal(data, "hop_count"))
	if warning := stringVal(data, "warning"); warning != "N/A" {
		addKeyValueRowStyled(m, "Warning", warning, colorPtr(colorAmber), nil)
	}

	if vantages := mapSlice(data, "vantages"); len(vantages) > 0 {
		for i, vantage := range vantages {
			label := fmt.Sprintf("Vantage %d", i+1)
			if name := stringVal(vantage, "name", "label"); name != "N/A" {
				label = name
			}
			if source := stringVal(vantage, "source"); source != "N/A" {
				label += " (" + source + ")"
			}
			addSubheader(m, label)
			if warning := stringVal(vantage, "warning"); warning != "N/A" {
				addKeyValueRowStyled(m, "Warning", warning, colorPtr(colorAmber), nil)
			}
			if hops := mapSlice(vantage, "hops"); len(hops) > 0 {
				addTracerouteHopTable(m, hops)
			}
		}
	} else if hops := mapSlice(data, "hops"); len(hops) > 0 {
		addSubheader(m, "Route")
		addTracerouteHopTable(m, hops)
	}
}

func addErrors(m core.Maroto, errors []string) {
	addBulletItems(m, errors, maxErrorsShown)
}

func joinDNSRecords(data map[string]any) string {
	if len(data) == 0 {
		return "N/A"
	}
	a := stringVal(data, "A", "a")
	aaaa := stringVal(data, "AAAA", "aaaa")
	if a == "N/A" && aaaa == "N/A" {
		return "N/A"
	}
	if a == "N/A" {
		return aaaa
	}
	if aaaa == "N/A" {
		return a
	}
	return a + "; " + aaaa
}

func screenshotDimensions(meta map[string]any) string {
	width := stringVal(meta, "width")
	height := stringVal(meta, "height")
	if width == "N/A" || height == "N/A" {
		return "N/A"
	}
	return width + " × " + height
}

func extractCopyrights(data map[string]any) []string {
	if data == nil {
		return nil
	}

	v, ok := data["copyrights"]
	if !ok || v == nil {
		return nil
	}

	var out []string
	seen := make(map[string]struct{})

	add := func(s string) {
		s = strings.TrimSpace(s)
		if s == "" {
			return
		}
		if _, exists := seen[s]; exists {
			return
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}

	switch s := v.(type) {
	case []string:
		for _, item := range s {
			add(item)
		}
	case []any:
		for _, item := range s {
			if item == nil {
				continue
			}
			if str, ok := item.(string); ok {
				add(str)
			}
		}
	case string:
		add(s)
	}

	return out
}

func formatWordPressItems(items []map[string]any) []string {
	var lines []string
	for _, item := range items {
		slug := stringVal(item, "slug")
		version := stringVal(item, "version")
		switch {
		case slug != "N/A" && version != "N/A":
			lines = append(lines, fmt.Sprintf("%s (%s)", slug, version))
		case slug != "N/A":
			lines = append(lines, slug)
		default:
			lines = append(lines, formatMapLines(item, ""))
		}
	}
	return lines
}

func mailPostureSummary(dns map[string]any) string {
	posture := nestedMap(dns, "MAIL_POSTURE")
	if len(posture) == 0 {
		return "N/A"
	}
	grade := stringVal(posture, "grade")
	score := stringVal(posture, "score")
	if grade == "N/A" && score == "N/A" {
		return "N/A"
	}
	if grade != "N/A" && score != "N/A" {
		return grade + " (" + score + "/100)"
	}
	if grade != "N/A" {
		return grade
	}
	return score
}

func ctCount(data map[string]any) string {
	if count := stringVal(data, "count"); count != "N/A" {
		return count
	}
	if subdomains := stringSlice(data, "subdomains"); len(subdomains) > 0 {
		return fmt.Sprintf("%d", len(subdomains))
	}
	return "N/A"
}

func pwhoisPrefix(info *PWhoisInfo) string {
	if info == nil {
		return "N/A"
	}
	return valueOrNA(info.Prefix)
}

func pwhoisCountry(info *PWhoisInfo) string {
	if info == nil {
		return "N/A"
	}
	return valueOrNA(info.CountryCode)
}

func firstNonNA(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" && value != "N/A" {
			return value
		}
	}
	return "N/A"
}

func valueOrNA(s string) string {
	if strings.TrimSpace(s) == "" {
		return "N/A"
	}
	return s
}

func truncate(s string, n int) string {
	if n <= 0 {
		return s
	}
	if len(s) <= n {
		return s
	}
	if n <= 3 {
		return s[:n]
	}
	return s[:n-3] + "..."
}

func formatPortList(raw any) string {
	switch ports := raw.(type) {
	case []any:
		var out []string
		for _, item := range ports {
			val := strings.TrimSpace(fmt.Sprint(item))
			if val != "" && val != "<nil>" {
				out = append(out, val)
			}
		}
		if len(out) == 0 {
			return ""
		}
		return strings.Join(out, ", ")
	case []int:
		var out []string
		for _, port := range ports {
			out = append(out, fmt.Sprintf("%d", port))
		}
		return strings.Join(out, ", ")
	default:
		val := strings.TrimSpace(fmt.Sprint(raw))
		if val == "" || val == "<nil>" || val == "N/A" {
			return ""
		}
		return val
	}
}

func formatEnrichmentServices(services []map[string]any, provider string) []string {
	var lines []string
	for i, service := range services {
		if i >= 6 {
			lines = append(lines, fmt.Sprintf("... and %d more service(s)", len(services)-6))
			break
		}
		switch provider {
		case "censys":
			lines = append(lines, fmt.Sprintf(
				"%s:%s %s",
				stringVal(service, "transport"),
				stringVal(service, "port"),
				stringVal(service, "service_name"),
			))
		default:
			product := stringVal(service, "product")
			version := stringVal(service, "version")
			label := stringVal(service, "service_name")
			if label == "N/A" {
				label = product
			}
			if version != "N/A" && label != "N/A" {
				label += " " + version
			}
			lines = append(lines, fmt.Sprintf(
				"%s:%s %s",
				stringVal(service, "transport"),
				stringVal(service, "port"),
				label,
			))
		}
	}
	return lines
}

func formatShodanFaviconMatches(search map[string]any) []string {
	matches := mapSlice(search, "matches")
	if len(matches) == 0 {
		return nil
	}
	var lines []string
	for i, match := range matches {
		if i >= maxErrorsShown {
			lines = append(lines, fmt.Sprintf("... and %d more host(s)", len(matches)-maxErrorsShown))
			break
		}
		ip := stringVal(match, "ip_str", "ip")
		org := stringVal(match, "org")
		ports := formatPortList(match["ports"])
		line := ip
		if org != "N/A" {
			line += " (" + org + ")"
		}
		if ports != "" {
			line += " ports: " + ports
		}
		lines = append(lines, line)
	}
	return lines
}

func censysSearchTotal(search map[string]any) string {
	if total := stringVal(search, "total"); total != "N/A" {
		return total
	}
	if result, ok := search["result"].(map[string]any); ok {
		if total := stringVal(result, "total"); total != "N/A" {
			return total
		}
	}
	return "N/A"
}

func formatCensysJARMHits(search map[string]any) []string {
	result, _ := search["result"].(map[string]any)
	if len(result) == 0 {
		return nil
	}
	hits, _ := result["hits"].([]any)
	if len(hits) == 0 {
		return nil
	}
	var lines []string
	for i, item := range hits {
		if i >= maxErrorsShown {
			lines = append(lines, fmt.Sprintf("... and %d more host(s)", len(hits)-maxErrorsShown))
			break
		}
		row, ok := item.(map[string]any)
		if !ok {
			continue
		}
		ip := stringVal(row, "ip")
		if ip == "N/A" {
			continue
		}
		if name := stringVal(row, "name"); name != "N/A" {
			lines = append(lines, ip+" ("+name+")")
			continue
		}
		lines = append(lines, ip)
	}
	return lines
}