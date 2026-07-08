package db

import "testing"

func TestNormalizeSAN(t *testing.T) {
	got := normalizeSAN("*.Example.COM.")
	if got != "*.example.com" {
		t.Fatalf("normalizeSAN() = %q", got)
	}
}

func TestUnionFind(t *testing.T) {
	uf := newUnionFind([]string{"a", "b", "c", "d"})
	uf.union("a", "b")
	uf.union("c", "d")
	if uf.find("a") != uf.find("b") {
		t.Fatal("expected a and b in same set")
	}
	if uf.find("a") == uf.find("c") {
		t.Fatal("expected separate sets")
	}
	uf.union("b", "c")
	if uf.find("a") != uf.find("d") {
		t.Fatal("expected merged cluster")
	}
}
