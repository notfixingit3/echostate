package scanner

import "testing"

func TestParseASPath(t *testing.T) {
	got := parseASPath("3257 15169")
	if len(got) != 2 || got[0] != "3257" || got[1] != "15169" {
		t.Fatalf("parseASPath() = %#v", got)
	}
}
