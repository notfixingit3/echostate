package scanner

import "testing"

func TestExtractCookieName(t *testing.T) {
	t.Parallel()

	if got := extractCookieName("sessionid=abc; Path=/; HttpOnly"); got != "sessionid" {
		t.Fatalf("name = %q", got)
	}
}

func TestMergeCookieNames(t *testing.T) {
	t.Parallel()

	merged := mergeCookieNames([]string{"a", "b"}, []string{"b", "c"})
	if len(merged) != 3 || merged[0] != "a" || merged[2] != "c" {
		t.Fatalf("merged = %#v", merged)
	}
}
