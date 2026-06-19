package pdf

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrapText(t *testing.T) {
	lines := wrapText("one two three four five six seven eight nine ten", 20)
	assert.Greater(t, len(lines), 1)
	assert.Contains(t, strings.Join(lines, " "), "one")
}

func TestJoinDNSRecords_UppercaseKeys(t *testing.T) {
	got := joinDNSRecords(map[string]any{
		"A": []string{"93.184.216.34", "2606:2800:220:1:248:1893:25c8:1946"},
	})
	assert.Contains(t, got, "93.184.216.34")
}

func TestSanitizeReportError(t *testing.T) {
	raw := `traceroute: external: external traceroute: HTTP 404`
	assert.Equal(t, "external: HTTP 404", sanitizeReportError(raw))
}

func TestTableRowHeightForCells_GrowsWithWrappedText(t *testing.T) {
	colWidths := []int{2, 10}
	cells := []tableCell{
		{text: "HIGH"},
		{text: strings.Repeat("word ", 40)},
	}
	height := tableRowHeightForCells(colWidths, cells)
	assert.Greater(t, height, tableRowHeight)
}

func TestFormatMapLines(t *testing.T) {
	got := formatMapLines(map[string]any{
		"hijack_risk": "low",
		"notes":       []any{"ok"},
	}, "")
	assert.Contains(t, got, "Hijack Risk: low")
	assert.Contains(t, got, "Notes: ok")
}