package scanner

import (
	"context"
	"testing"
)

func TestGatherTLS_EmptyHost(t *testing.T) {
	t.Parallel()

	key, data, err := gatherTLS(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty host")
	}
	if key != "tls" {
		t.Errorf("key = %q, want tls", key)
	}
	if data != nil {
		t.Errorf("data = %v, want nil", data)
	}
}
