package scanner

import "testing"

func TestBuildBGPPathProfile(t *testing.T) {
	profile := buildBGPPathProfile([]string{
		"3257 15169",
		"174 3356 15169",
		"6939 15169",
	})
	if profile == nil {
		t.Fatal("expected profile")
	}
	if profile["stability"] != "diverse" {
		t.Fatalf("stability = %v, want diverse", profile["stability"])
	}
	if profile["primary_path"] != "3257 15169" {
		t.Fatalf("primary_path = %v", profile["primary_path"])
	}
}
