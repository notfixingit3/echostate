package notes

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSanitizeBody_AllowsAnchorAndAutolinksURL(t *testing.T) {
	input := `See <a href="https://example.com/path">Example</a> and https://ripe.net/db`
	out := SanitizeBody(input)
	require.Contains(t, out, `<a href="https://example.com/path"`)
	require.Contains(t, out, `Example</a>`)
	require.Contains(t, out, `<a href="https://ripe.net/db"`)
	require.NotContains(t, out, "<script")
}

func TestSanitizeBody_StripsDangerousTags(t *testing.T) {
	input := `<script>alert(1)</script>Hello <b>world</b>`
	out := SanitizeBody(input)
	require.Equal(t, "Hello world", stripTags(out))
	require.NotContains(t, strings.ToLower(out), "script")
}
