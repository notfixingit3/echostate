package scanner

import (
	"context"
	"testing"
)

func TestGatherDNS_EmptyHost(t *testing.T) {
	t.Parallel()

	key, data, err := gatherDNS(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty host")
	}
	if key != "dns" {
		t.Errorf("key = %q, want dns", key)
	}
	if data != nil {
		t.Errorf("data = %v, want nil", data)
	}
}