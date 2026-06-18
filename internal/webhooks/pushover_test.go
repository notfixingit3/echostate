package webhooks

import (
	"testing"

	"github.com/notfixingit3/echostate/internal/models"
)

func TestParsePushoverConfig_RequiredFields(t *testing.T) {
	t.Parallel()

	_, err := ParsePushoverConfig(map[string]any{
		"app_token": "abc",
	})
	if err == nil {
		t.Fatal("expected error for missing user_key")
	}
}

func TestParsePushoverConfig_EmergencyRequiresRetryExpire(t *testing.T) {
	t.Parallel()

	_, err := ParsePushoverConfig(map[string]any{
		"app_token": "abc123",
		"user_key":  "user123",
		"priority":  2,
	})
	if err == nil {
		t.Fatal("expected error for missing retry/expire")
	}
}

func TestParsePushoverConfig_ValidEmergency(t *testing.T) {
	t.Parallel()

	cfg, err := ParsePushoverConfig(map[string]any{
		"app_token": "abc123",
		"user_key":  "user123",
		"priority":  2,
		"retry":     60,
		"expire":    1800,
		"sound":     "siren",
		"device":    "iphone",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Retry != 60 || cfg.Expire != 1800 || cfg.Sound != "siren" {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}

func TestParsePushoverConfig_Defaults(t *testing.T) {
	t.Parallel()

	cfg, err := ParsePushoverConfig(map[string]any{
		"app_token": "abc123",
		"user_key":  "user123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Priority != 0 {
		t.Fatalf("priority = %d, want 0", cfg.Priority)
	}
	if cfg.Title != "EchoState Alert" {
		t.Fatalf("title = %q", cfg.Title)
	}
	if cfg.URLTitle != "View target" {
		t.Fatalf("url_title = %q", cfg.URLTitle)
	}
}

func TestSendPushover_InvalidConfig(t *testing.T) {
	t.Parallel()

	err := sendPushover(models.Webhook{Type: "pushover"}, "example.com", "- changed asn", "http://localhost/target")
	if err == nil {
		t.Fatal("expected error for invalid config")
	}
}