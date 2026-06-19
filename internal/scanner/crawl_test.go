package scanner

import "testing"

func TestParseSitemapXML(t *testing.T) {
	t.Parallel()

	xml := `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url><loc>https://example.com/</loc></url>
  <url><loc>https://example.com/about</loc></url>
</urlset>`

	urls := parseSitemapURLSet(xml)
	if len(urls) != 2 {
		t.Fatalf("got %d urls, want 2", len(urls))
	}
}

func TestClassifySitemap_IndexReturnsChildSitemaps(t *testing.T) {
	t.Parallel()

	xml := `<?xml version="1.0" encoding="UTF-8"?>
<sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <sitemap><loc>https://example.com/sitemap-pages.xml</loc></sitemap>
  <sitemap><loc>https://example.com/sitemap-posts.xml</loc></sitemap>
</sitemapindex>`

	pageURLs, childSitemaps := classifySitemap(xml)
	if len(pageURLs) != 0 {
		t.Fatalf("pageURLs = %#v, want empty", pageURLs)
	}
	if len(childSitemaps) != 2 {
		t.Fatalf("childSitemaps = %#v, want 2 entries", childSitemaps)
	}
}

func TestClassifySitemap_URLSetReturnsPages(t *testing.T) {
	t.Parallel()

	xml := `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url><loc>https://example.com/pricing</loc></url>
</urlset>`

	pageURLs, childSitemaps := classifySitemap(xml)
	if len(childSitemaps) != 0 {
		t.Fatalf("childSitemaps = %#v, want empty", childSitemaps)
	}
	if len(pageURLs) != 1 || pageURLs[0] != "https://example.com/pricing" {
		t.Fatalf("pageURLs = %#v, want pricing page", pageURLs)
	}
}