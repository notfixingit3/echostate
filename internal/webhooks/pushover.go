package webhooks

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/notfixingit3/echostate/internal/models"
)

const pushoverMessagesURL = "https://api.pushover.net/1/messages.json"

// PushoverConfig holds Pushover-specific notification options.
type PushoverConfig struct {
	AppToken string `json:"app_token"`
	UserKey  string `json:"user_key"`
	Priority int    `json:"priority"`
	Sound    string `json:"sound,omitempty"`
	Device   string `json:"device,omitempty"`
	Title    string `json:"title,omitempty"`
	URLTitle string `json:"url_title,omitempty"`
	Retry    int    `json:"retry,omitempty"`
	Expire   int    `json:"expire,omitempty"`
}

func ParsePushoverConfig(config map[string]any) (PushoverConfig, error) {
	cfg := PushoverConfig{
		Priority: 0,
		Title:    "EchoState Alert",
		URLTitle: "View target",
	}
	if config == nil {
		return cfg, fmt.Errorf("pushover config is required")
	}

	cfg.AppToken = strings.TrimSpace(stringVal(config["app_token"]))
	cfg.UserKey = strings.TrimSpace(stringVal(config["user_key"]))
	cfg.Sound = strings.TrimSpace(stringVal(config["sound"]))
	cfg.Device = strings.TrimSpace(stringVal(config["device"]))

	if title := strings.TrimSpace(stringVal(config["title"])); title != "" {
		cfg.Title = title
	}
	if urlTitle := strings.TrimSpace(stringVal(config["url_title"])); urlTitle != "" {
		cfg.URLTitle = urlTitle
	}

	switch v := config["priority"].(type) {
	case float64:
		cfg.Priority = int(v)
	case int:
		cfg.Priority = v
	case int64:
		cfg.Priority = int(v)
	default:
		cfg.Priority = 0
	}

	switch v := config["retry"].(type) {
	case float64:
		cfg.Retry = int(v)
	case int:
		cfg.Retry = v
	case int64:
		cfg.Retry = int(v)
	}

	switch v := config["expire"].(type) {
	case float64:
		cfg.Expire = int(v)
	case int:
		cfg.Expire = v
	case int64:
		cfg.Expire = int(v)
	}

	if cfg.AppToken == "" || cfg.UserKey == "" {
		return cfg, fmt.Errorf("pushover app_token and user_key are required")
	}

	if cfg.Priority < -2 || cfg.Priority > 2 {
		return cfg, fmt.Errorf("pushover priority must be between -2 and 2")
	}

	if cfg.Priority == 2 {
		if cfg.Retry < 30 {
			return cfg, fmt.Errorf("pushover retry must be at least 30 seconds for emergency priority")
		}
		if cfg.Expire < 1 || cfg.Expire > 10800 {
			return cfg, fmt.Errorf("pushover expire must be between 1 and 10800 seconds for emergency priority")
		}
	}

	return cfg, nil
}

func sendPushover(hook models.Webhook, host, changes, link string) error {
	cfg, err := ParsePushoverConfig(hook.Config)
	if err != nil {
		return err
	}

	message := fmt.Sprintf("Changes detected for %s:\n\n%s", host, changes)
	if len(message) > 1024 {
		message = message[:1021] + "..."
	}

	form := url.Values{}
	form.Set("token", cfg.AppToken)
	form.Set("user", cfg.UserKey)
	form.Set("title", cfg.Title)
	form.Set("message", message)
	form.Set("url", link)
	form.Set("url_title", cfg.URLTitle)
	form.Set("priority", strconv.Itoa(cfg.Priority))

	if cfg.Sound != "" {
		form.Set("sound", cfg.Sound)
	}
	if cfg.Device != "" {
		form.Set("device", cfg.Device)
	}
	if cfg.Priority == 2 {
		form.Set("retry", strconv.Itoa(cfg.Retry))
		form.Set("expire", strconv.Itoa(cfg.Expire))
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodPost, pushoverMessagesURL, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("pushover API returned status %d", resp.StatusCode)
	}

	return nil
}

func stringVal(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}