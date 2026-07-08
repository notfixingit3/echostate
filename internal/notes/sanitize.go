package notes

import (
	"fmt"
	"html"
	"regexp"
	"strings"
)

var (
	scriptBlock = regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`)
	styleBlock  = regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`)
	anchorTag   = regexp.MustCompile(`(?is)<a\s+href="(https?://[^"]+)"[^>]*>(.*?)</a>`)
	anyHTMLTag  = regexp.MustCompile(`(?is)<[^>]+>`)
	bareURL     = regexp.MustCompile(`https?://[^\s<>"']+`)
)

const anchorPlaceholderPrefix = "\x00ANCHOR:"

// SanitizeBody keeps plain text, auto-safe bare URLs, and limited <a href="http(s)://..."> links.
func SanitizeBody(body string) string {
	body = strings.TrimSpace(body)
	if body == "" {
		return ""
	}

	body = scriptBlock.ReplaceAllString(body, "")
	body = styleBlock.ReplaceAllString(body, "")
	anchors := make([]string, 0)
	body = anchorTag.ReplaceAllStringFunc(body, func(match string) string {
		sub := anchorTag.FindStringSubmatch(match)
		if len(sub) < 3 {
			return html.EscapeString(stripTags(match))
		}
		href := html.EscapeString(sub[1])
		label := html.EscapeString(stripTags(sub[2]))
		anchors = append(anchors, `<a href="`+href+`" target="_blank" rel="noopener noreferrer">`+label+`</a>`)
		return fmt.Sprintf("%s%d\x00", anchorPlaceholderPrefix, len(anchors)-1)
	})

	body = html.EscapeString(stripTags(body))
	body = bareURL.ReplaceAllStringFunc(body, func(url string) string {
		escaped := html.EscapeString(url)
		return `<a href="` + escaped + `" target="_blank" rel="noopener noreferrer">` + escaped + `</a>`
	})

	for i, anchor := range anchors {
		body = strings.ReplaceAll(body, fmt.Sprintf("%s%d\x00", anchorPlaceholderPrefix, i), anchor)
	}
	return body
}

func stripTags(value string) string {
	return strings.TrimSpace(anyHTMLTag.ReplaceAllString(value, ""))
}

// PlainText returns user-visible text without HTML tags.
func PlainText(body string) string {
	return stripTags(body)
}
