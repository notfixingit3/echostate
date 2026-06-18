package webhooks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/notfixingit3/echostate/internal/db"
	"github.com/notfixingit3/echostate/internal/models"
)

func Dispatch(ctx context.Context, database *db.DB, targetHost string, snapshot *models.Snapshot, frontendURL string) {
	if len(snapshot.Changes) == 0 {
		return
	}

	rows, err := database.Pool.Query(ctx, `SELECT id, name, type, url, config, enabled FROM webhooks WHERE enabled = true`)
	if err != nil {
		return
	}
	defer rows.Close()

	var hooks []models.Webhook
	for rows.Next() {
		var w models.Webhook
		var configBytes []byte
		if err := rows.Scan(&w.ID, &w.Name, &w.Type, &w.URL, &configBytes, &w.Enabled); err == nil {
			if len(configBytes) > 0 {
				_ = json.Unmarshal(configBytes, &w.Config)
			}
			hooks = append(hooks, w)
		}
	}

	if len(hooks) == 0 {
		return
	}

	changesStr := ""
	for _, c := range snapshot.Changes {
		changesStr += "- " + c + "\n"
	}
	if len(changesStr) > 500 {
		changesStr = changesStr[:500] + "...\n(Truncated)"
	}
	targetLink := fmt.Sprintf("%s/targets/%s", frontendURL, snapshot.TargetID.String())

	for _, hook := range hooks {
		go sendWebhook(hook, targetHost, changesStr, targetLink)
	}
}

func sendWebhook(hook models.Webhook, host, changes, link string) {
	var payload []byte
	var err error

	msg := fmt.Sprintf("🚨 **EchoState Alert**\nChanges detected for **%s**\n\n```\n%s```\n[View Details](%s)", host, changes, link)

	switch hook.Type {
	case "pushover":
		_ = sendPushover(hook, host, changes, link)
		return
	case "slack":
		p := map[string]string{"text": msg}
		payload, err = json.Marshal(p)
	case "discord":
		p := map[string]string{"content": msg}
		payload, err = json.Marshal(p)
	case "teams":
		p := map[string]string{"text": msg}
		payload, err = json.Marshal(p)
	default:
		return
	}

	if err != nil {
		return
	}

	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("POST", hook.URL, bytes.NewBuffer(payload))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	client.Do(req)
}
