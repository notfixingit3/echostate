package scanner

import "testing"

func TestParseCAARData(t *testing.T) {
	t.Parallel()

	record, ok := parseCAARData([]byte{0x00, 0x05, 'i', 's', 's', 'u', 'e', '"', 'c', 'a', '.', 'e', 'x', 'a', 'm', 'p', 'l', 'e', '.', 'c', 'o', 'm', '"'})
	if !ok {
		t.Fatal("parseCAARData() failed")
	}
	if record["tag"] != "issue" {
		t.Fatalf("tag = %#v", record["tag"])
	}
}