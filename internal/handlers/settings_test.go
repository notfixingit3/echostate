package handlers

import "testing"

func TestMergeSecretPreservesMaskedValue(t *testing.T) {
	current := "supersecretkey12345"
	masked := "supe****"

	if got := mergeSecret(masked, current); got != current {
		t.Fatalf("expected preserved secret %q, got %q", current, got)
	}
}

func TestMergeSecretAcceptsNewValue(t *testing.T) {
	current := "old-key-value-here"
	incoming := "brand-new-secret-key"

	if got := mergeSecret(incoming, current); got != incoming {
		t.Fatalf("expected new secret %q, got %q", incoming, got)
	}
}

func TestMergeSecretPreservesVirusTotalKey(t *testing.T) {
	current := "vt-secret-api-key-abcdef"
	masked := "vt-s****"

	if got := mergeSecret(masked, current); got != current {
		t.Fatalf("expected VirusTotal key preserved as %q, got %q", current, got)
	}
}