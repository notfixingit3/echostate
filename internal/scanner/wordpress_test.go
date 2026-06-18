package scanner

import (
	"reflect"
	"testing"
)

func TestExtractWordPressPluginFromURL(t *testing.T) {
	slug, version, ok := extractWordPressPluginFromURL(
		"https://example.com/wp-content/plugins/woocommerce/assets/js/frontend.min.js?ver=8.9.1",
	)
	if !ok || slug != "woocommerce" || version != "8.9.1" {
		t.Fatalf("extractWordPressPluginFromURL() = (%q, %q, %v)", slug, version, ok)
	}

	_, _, ok = extractWordPressPluginFromURL("https://example.com/wp-includes/js/jquery.js")
	if ok {
		t.Fatal("expected non-plugin URL to be ignored")
	}
}

func TestExtractWordPressThemeFromURL(t *testing.T) {
	slug, version, ok := extractWordPressThemeFromURL(
		"https://example.com/wp-content/themes/twentytwentyfour/style.css?ver=1.2",
	)
	if !ok || slug != "twentytwentyfour" || version != "1.2" {
		t.Fatalf("extractWordPressThemeFromURL() = (%q, %q, %v)", slug, version, ok)
	}
}

func TestNormalizeWordPressItems(t *testing.T) {
	got := normalizeWordPressItems([]wordpressItem{
		{Slug: "Akismet", Version: ""},
		{Slug: "woocommerce", Version: "8.9.1"},
		{Slug: "woocommerce", Version: ""},
	})

	want := []map[string]any{
		{"slug": "akismet"},
		{"slug": "woocommerce", "version": "8.9.1"},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("normalizeWordPressItems() = %#v, want %#v", got, want)
	}
}