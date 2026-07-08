package pdf

import (
	"fmt"
	"strings"
	"time"

	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
)

const (
	maxFieldLen       = 512
	valueCharsPerLine = 72
	kvLineHeight      = 4.5
	rawLineHeight     = 3.8
	labelColWidth     = 3
	valueColWidth     = 9
	tableRowHeight    = 6.0
	tableHeaderHeight = 7.0
	tableLineHeight   = 3.8
	tableTotalCols    = 12
)

type tableCell struct {
	text  string
	link  *string
	color *props.Color
	bold  bool
}

func addKeyValueRow(m core.Maroto, label, value string) {
	addKeyValueRowStyled(m, label, value, nil, nil)
}

func addKeyValueRowStyled(m core.Maroto, label, value string, valueColor *props.Color, link *string) {
	value = strings.TrimSpace(value)
	if value == "" {
		value = "N/A"
	}

	lines := wrapText(value, valueCharsPerLine)
	rowHeight := float64(len(lines)) * kvLineHeight
	if rowHeight < kvLineHeight {
		rowHeight = kvLineHeight
	}

	valueProps := props.Text{Size: 10, Top: 0.5}
	if valueColor != nil {
		valueProps.Color = valueColor
	}
	if link != nil {
		valueProps.Hyperlink = link
		if valueProps.Color == nil {
			valueProps.Color = colorPtr(colorLink)
		}
	}

	m.AddRows(
		row.New(rowHeight).Add(
			text.NewCol(labelColWidth, label+":", props.Text{Size: 10, Style: fontstyle.Bold, Top: 0.5}),
			text.NewCol(valueColWidth, strings.Join(lines, "\n"), valueProps),
		),
	)
}

func addTable(m core.Maroto, colWidths []int, headers []string, rows [][]tableCell) {
	if len(headers) == 0 {
		return
	}

	headerCells := make([]tableCell, len(headers))
	for i, header := range headers {
		headerCells[i] = tableCell{text: header, bold: true}
	}
	addTableRow(m, tableHeaderHeight, colWidths, headerCells, true)

	for rowIndex, dataRow := range rows {
		bg := &props.Cell{}
		if rowIndex%2 == 0 {
			bg.BackgroundColor = colorPtr(colorTableRow)
		}
		addTableRowStyled(m, tableRowHeightForCells(colWidths, dataRow), colWidths, dataRow, bg)
	}
}

func addTableRow(m core.Maroto, height float64, colWidths []int, cells []tableCell, header bool) {
	bg := &props.Cell{}
	if header {
		bg.BackgroundColor = colorPtr(colorTableHead)
	}
	addTableRowStyled(m, height, colWidths, cells, bg)
}

func tableCharsForCol(width int) int {
	if width <= 0 {
		return valueCharsPerLine
	}
	chars := int(float64(valueCharsPerLine) * float64(width) / float64(tableTotalCols))
	if chars < 12 {
		return 12
	}
	return chars
}

func tableRowHeightForCells(colWidths []int, cells []tableCell) float64 {
	maxLines := 1
	for i, cell := range cells {
		if i >= len(colWidths) {
			break
		}
		lines := wrapText(cell.text, tableCharsForCol(colWidths[i]))
		if len(lines) > maxLines {
			maxLines = len(lines)
		}
	}
	height := float64(maxLines)*tableLineHeight + 1.5
	if height < tableRowHeight {
		return tableRowHeight
	}
	return height
}

func formatTableCellText(text string, colWidth int) string {
	return strings.Join(wrapText(text, tableCharsForCol(colWidth)), "\n")
}

func addTableRowStyled(m core.Maroto, height float64, colWidths []int, cells []tableCell, style *props.Cell) {
	r := row.New(height)
	if style != nil {
		r = r.WithStyle(style)
	}

	for i, cell := range cells {
		width := colWidths[i]
		textProps := props.Text{Size: 8.5, Top: 1.2, Left: 1}
		if cell.bold {
			textProps.Style = fontstyle.Bold
		}
		if cell.color != nil {
			textProps.Color = cell.color
		}
		if cell.link != nil {
			textProps.Hyperlink = cell.link
			if textProps.Color == nil {
				textProps.Color = colorPtr(colorLink)
			}
		}
		r.Add(text.NewCol(width, formatTableCellText(cell.text, width), textProps))
	}
	m.AddRows(r)
}

func sanitizeReportError(msg string) string {
	msg = strings.TrimSpace(msg)
	if msg == "" {
		return msg
	}

	for _, verb := range []string{`Get "`, `Post "`, `Head "`} {
		if strings.HasPrefix(msg, verb) {
			if end := strings.Index(msg, `": `); end > 0 {
				msg = strings.TrimSpace(msg[end+3:])
			}
			break
		}
	}

	msg = strings.TrimPrefix(msg, "traceroute: ")
	for strings.Contains(msg, "external traceroute: ") {
		msg = strings.Replace(msg, "external traceroute: ", "", 1)
	}
	msg = strings.TrimPrefix(msg, "traceroute: ")

	switch {
	case strings.Contains(msg, "context deadline exceeded"),
		strings.Contains(msg, "Client.Timeout exceeded"),
		strings.Contains(msg, "i/o timeout"):
		return "request timed out"
	default:
		return msg
	}
}

func addSubheader(m core.Maroto, title string) {
	m.AddRows(row.New(2).Add(col.New(12)))
	m.AddRows(
		text.NewRow(7, title, props.Text{Size: 11, Style: fontstyle.Bold}),
	)
}

func addBulletItems(m core.Maroto, items []string, maxItems int) {
	if len(items) == 0 {
		m.AddRows(text.NewRow(5, "None detected.", props.Text{Size: 9}))
		return
	}

	for i, item := range items {
		if maxItems > 0 && i >= maxItems {
			m.AddRows(text.NewRow(5, fmt.Sprintf("... and %d more", len(items)-maxItems), props.Text{
				Size:  9,
				Style: fontstyle.Italic,
			}))
			break
		}
		lines := wrapText("• "+item, valueCharsPerLine+8)
		rowHeight := float64(len(lines)) * 4.0
		if rowHeight < 4.0 {
			rowHeight = 4.0
		}
		m.AddRows(
			text.NewRow(rowHeight, strings.Join(lines, "\n"), props.Text{Size: 9}),
		)
	}
}

func addMultilineBlock(m core.Maroto, title, body string, maxLines int) {
	body = strings.TrimSpace(body)
	if body == "" {
		return
	}

	addSubheader(m, title)

	lines := strings.Split(strings.ReplaceAll(body, "\r\n", "\n"), "\n")
	if maxLines > 0 && len(lines) > maxLines {
		lines = append(lines[:maxLines], fmt.Sprintf("... truncated (%d more lines)", len(lines)-maxLines))
	}

	for _, line := range lines {
		wrapped := wrapText(line, valueCharsPerLine+12)
		for _, part := range wrapped {
			m.AddRows(
				text.NewRow(rawLineHeight, part, props.Text{Size: 8}),
			)
		}
	}
}

func wrapText(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}

	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	var lines []string
	for _, paragraph := range strings.Split(text, "\n") {
		paragraph = strings.TrimSpace(paragraph)
		if paragraph == "" {
			lines = append(lines, "")
			continue
		}
		lines = append(lines, wrapParagraph(paragraph, width)...)
	}
	if len(lines) == 0 {
		return []string{""}
	}
	return lines
}

func wrapParagraph(text string, width int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}

	var lines []string
	current := words[0]
	for _, word := range words[1:] {
		candidate := current + " " + word
		if len(candidate) <= width {
			current = candidate
			continue
		}
		lines = append(lines, current)
		current = word
	}
	lines = append(lines, current)
	return lines
}

func lookupVal(data map[string]any, keys ...string) (any, bool) {
	if data == nil {
		return nil, false
	}
	for _, key := range keys {
		if v, ok := data[key]; ok && v != nil {
			return v, true
		}
	}
	return nil, false
}

func stringVal(data map[string]any, keys ...string) string {
	v, ok := lookupVal(data, keys...)
	if !ok {
		return "N/A"
	}
	return formatScalar(v)
}

func nestedMap(data map[string]any, key string) map[string]any {
	if data == nil {
		return nil
	}
	v, ok := data[key]
	if !ok || v == nil {
		return nil
	}
	m, ok := v.(map[string]any)
	if !ok {
		return nil
	}
	return m
}

func formatScalar(v any) string {
	switch s := v.(type) {
	case string:
		if strings.TrimSpace(s) == "" {
			return "N/A"
		}
		return truncate(s, maxFieldLen)
	case []string:
		if len(s) == 0 {
			return "N/A"
		}
		return truncate(strings.Join(s, ", "), maxFieldLen)
	case []any:
		if len(s) == 0 {
			return "N/A"
		}
		var parts []string
		for _, item := range s {
			if item == nil {
				continue
			}
			parts = append(parts, formatAnyValue(item))
		}
		if len(parts) == 0 {
			return "N/A"
		}
		return truncate(strings.Join(parts, "; "), maxFieldLen)
	case time.Time:
		return s.Format(pdfDateFormat)
	case map[string]any:
		return truncate(formatMapLines(s, ""), maxFieldLen)
	default:
		str := strings.TrimSpace(fmt.Sprintf("%v", v))
		if str == "" {
			return "N/A"
		}
		return truncate(str, maxFieldLen)
	}
}

func formatAnyValue(v any) string {
	switch s := v.(type) {
	case string:
		return strings.TrimSpace(s)
	case map[string]any:
		return formatMapLines(s, "")
	case []string:
		return strings.Join(s, ", ")
	case []any:
		var parts []string
		for _, item := range s {
			parts = append(parts, formatAnyValue(item))
		}
		return strings.Join(parts, "; ")
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", s))
	}
}

func formatMapLines(data map[string]any, prefix string) string {
	if len(data) == 0 {
		return "N/A"
	}

	keys := sortedMapKeys(data)
	var lines []string
	for _, key := range keys {
		v := data[key]
		if v == nil {
			continue
		}
		label := prefix + formatFieldLabel(key)
		switch nested := v.(type) {
		case map[string]any:
			nestedLines := formatMapLines(nested, label+".")
			if nestedLines != "N/A" {
				lines = append(lines, nestedLines)
			}
		case []any:
			if len(nested) == 0 {
				continue
			}
			lines = append(lines, fmt.Sprintf("%s: %s", label, formatAnyValue(nested)))
		default:
			formatted := formatAnyValue(v)
			if formatted == "" {
				continue
			}
			lines = append(lines, fmt.Sprintf("%s: %s", label, formatted))
		}
	}
	if len(lines) == 0 {
		return "N/A"
	}
	return strings.Join(lines, "\n")
}

func sortedMapKeys(data map[string]any) []string {
	keys := make([]string, 0, len(data))
	for key := range data {
		keys = append(keys, key)
	}
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[j] < keys[i] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	return keys
}

func formatFieldLabel(key string) string {
	key = strings.ReplaceAll(key, "_", " ")
	parts := strings.Fields(key)
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
	}
	return strings.Join(parts, " ")
}

func stringSlice(data map[string]any, keys ...string) []string {
	v, ok := lookupVal(data, keys...)
	if !ok {
		return nil
	}
	switch s := v.(type) {
	case []string:
		return s
	case []any:
		var out []string
		for _, item := range s {
			if item == nil {
				continue
			}
			out = append(out, fmt.Sprintf("%v", item))
		}
		return out
	case string:
		if strings.TrimSpace(s) == "" {
			return nil
		}
		return []string{s}
	default:
		return nil
	}
}

func formatHopRTT(hop map[string]any) string {
	if hop["timeout"] == true {
		return "timeout"
	}
	rtt := stringVal(hop, "rtt_ms", "rtt", "latency_ms")
	if rtt == "N/A" {
		return "—"
	}
	return rtt + " ms"
}

func tracerouteHopIP(hop map[string]any) string {
	if hop["timeout"] == true {
		return "*"
	}
	return stringVal(hop, "ip", "address")
}

func tracerouteLocation(hop map[string]any) string {
	country := stringVal(hop, "country")
	city := stringVal(hop, "city")
	switch {
	case country != "N/A" && city != "N/A":
		return city + ", " + country
	case country != "N/A":
		return country
	case city != "N/A":
		return city
	default:
		return "—"
	}
}

func targetHTTPSURL(host string) string {
	host = strings.TrimSpace(host)
	if host == "" {
		return ""
	}
	if strings.HasPrefix(host, "http://") || strings.HasPrefix(host, "https://") {
		return host
	}
	return "https://" + host
}

func mapSlice(data map[string]any, key string) []map[string]any {
	v, ok := data[key]
	if !ok || v == nil {
		return nil
	}
	switch s := v.(type) {
	case []map[string]any:
		return s
	case []any:
		var out []map[string]any
		for _, item := range s {
			if m, ok := item.(map[string]any); ok {
				out = append(out, m)
			}
		}
		return out
	default:
		return nil
	}
}
