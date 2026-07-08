package scanner

import "testing"

func TestNormalizeJSAssets(t *testing.T) {
	assets, hints := normalizeJSAssets(jsAssetIntel{
		Assets: []jsAsset{
			{URL: "https://cdn.example.com/jquery-3.7.1.min.js"},
			{URL: "https://cdn.example.com/app.bundle.js"},
		},
		Hints: []string{"analytics: Google Tag Manager"},
	})

	if len(assets) != 2 {
		t.Fatalf("assets len = %d, want 2", len(assets))
	}
	if assets[0]["hint"] != "lib: jQuery" {
		t.Fatalf("first hint = %v, want lib: jQuery", assets[0]["hint"])
	}

	foundGTM := false
	for _, hint := range hints {
		if hint == "analytics: Google Tag Manager" {
			foundGTM = true
		}
	}
	if !foundGTM {
		t.Fatalf("hints = %#v, want GTM hint", hints)
	}
}

func TestHintForAssetURL(t *testing.T) {
	if got := hintForAssetURL("https://site.test/_next/static/chunks/main.js"); got != "framework: Next.js" {
		t.Fatalf("hintForAssetURL() = %q", got)
	}
}
