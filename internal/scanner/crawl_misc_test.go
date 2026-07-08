package scanner

import "testing"

func TestSplitNonEmptyLines(t *testing.T) {
	t.Parallel()

	content := "# comment\n\nline-one\n  line-two  \r\n# another\nline-three"
	lines := splitNonEmptyLines(content)
	want := []string{"line-one", "line-two", "line-three"}
	if len(lines) != len(want) {
		t.Fatalf("lines = %#v, want %#v", lines, want)
	}
	for i := range want {
		if lines[i] != want[i] {
			t.Fatalf("lines[%d] = %q, want %q", i, lines[i], want[i])
		}
	}
}
