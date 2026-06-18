package scanner

import (
	"testing"

	"github.com/notfixingit3/echostate/internal/config"
)

func TestGetResolver_Default(t *testing.T) {
	config.UpdateSettings(config.SystemSettings{})
	resolver := getResolver()
	if resolver == nil {
		t.Fatal("expected non-nil resolver")
	}
}

func TestGetResolver_CustomDNSServers(t *testing.T) {
	config.UpdateSettings(config.SystemSettings{
		DNSServers: "8.8.8.8,1.1.1.1",
	})
	resolver := getResolver()
	if resolver == nil {
		t.Fatal("expected non-nil resolver")
	}
	if resolver.Dial == nil {
		t.Fatal("expected custom Dial function when DNS servers are configured")
	}
}