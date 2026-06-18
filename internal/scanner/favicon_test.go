package scanner

import (
	"encoding/base64"
	"testing"

	"github.com/twmb/murmur3"
)

func TestShodanFaviconHash(t *testing.T) {
	t.Parallel()

	data := []byte{0x00, 0x01, 0x02, 0x03}
	b64 := base64.StdEncoding.EncodeToString(data)
	got := int32(murmur3.Sum32([]byte(b64)))
	if got == 0 {
		t.Fatal("expected non-zero hash")
	}
}

func TestDetectBuckets(t *testing.T) {
	t.Parallel()

	content := `
		<img src="https://my-bucket.s3.amazonaws.com/logo.png">
		<a href="https://assets.blob.core.windows.net/public/file.pdf">Azure</a>
		<script src="https://storage.googleapis.com/static-bucket/app.js"></script>
	`

	buckets := detectBuckets(content)
	if len(buckets) < 3 {
		t.Fatalf("expected at least 3 buckets, got %d", len(buckets))
	}
}

func TestParseRobotsTxt(t *testing.T) {
	t.Parallel()

	robots := `User-agent: *
Disallow: /admin/
Allow: /public/
Sitemap: https://example.com/sitemap.xml
`

	parsed := parseRobotsTxt(robots)
	disallow, _ := parsed["disallow"].([]string)
	if len(disallow) != 1 || disallow[0] != "/admin/" {
		t.Fatalf("unexpected disallow: %v", parsed["disallow"])
	}
}

func TestAssessHijackRisk(t *testing.T) {
	t.Parallel()

	risk, _ := assessHijackRisk("15169", "8.8.8.0/24", []string{"15169"}, map[string]any{})
	if risk != "low" {
		t.Fatalf("risk = %q, want low", risk)
	}

	risk, notes := assessHijackRisk("15169", "8.8.8.0/24", []string{"64512"}, map[string]any{})
	if risk != "high" {
		t.Fatalf("risk = %q, want high", risk)
	}
	if len(notes) == 0 {
		t.Fatal("expected notes")
	}
}