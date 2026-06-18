package webhooks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/notfixingit3/echostate/internal/config"
	"github.com/notfixingit3/echostate/internal/db"
	"github.com/notfixingit3/echostate/internal/diff"
	"github.com/notfixingit3/echostate/internal/models"
)

func Dispatch(ctx context.Context, database *db.DB, targetHost string, snapshot *models.Snapshot, frontendURL string) {
	if len(snapshot.ChangeDetails) == 0 && len(snapshot.Changes) == 0 {
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

	entries := snapshot.ChangeDetails
	if len(entries) == 0 {
		for _, summary := range snapshot.Changes {
			entries = append(entries, models.ChangeDetail{Type: "generic", Severity: diff.SeverityInfo, Summary: summary})
		}
	}

	targetLink := fmt.Sprintf("%s/target?id=%s", strings.TrimRight(frontendURL, "/"), snapshot.TargetID.String())

	for _, hook := range hooks {
		filtered := filterChangesForHook(entries, hook.ID)
		if len(filtered) == 0 {
			continue
		}
		go sendWebhook(hook, targetHost, filtered, targetLink)
	}
}

func filterChangesForHook(entries []models.ChangeDetail, webhookID uuid.UUID) []models.ChangeDetail {
	rules := config.GetSettings().AlertRules
	if len(rules) == 0 {
		return entries
	}

	var matched []models.ChangeDetail
	for _, entry := range entries {
		for _, rule := range rules {
			if !rule.Enabled {
				continue
			}
			if len(rule.WebhookIDs) > 0 && !containsID(rule.WebhookIDs, webhookID.String()) {
				continue
			}
			if diff.MatchesRule(toDiffEntry(entry), rule.MatchTypes, rule.MinSeverity) {
				matched = append(matched, entry)
				break
			}
		}
	}
	return matched
}

func toDiffEntry(entry models.ChangeDetail) diff.Entry {
	return diff.Entry{
		Type:     entry.Type,
		Severity: entry.Severity,
		Summary:  entry.Summary,
		Field:    entry.Field,
		Detail:   entry.Detail,
	}
}

func containsID(ids []string, want string) bool {
	for _, id := range ids {
		if id == want {
			return true
		}
	}
	return false
}

func sendWebhook(hook models.Webhook, host string, changes []models.ChangeDetail, link string) {
	lines := make([]string, 0, len(changes))
	for _, change := range changes {
		lines = append(lines, "- "+diff.FormatSummary(toDiffEntry(change)))
	}
	changesStr := strings.Join(lines, "\n")
	if len(changesStr) > 1200 {
		changesStr = changesStr[:1200] + "\n...(truncated)"
	}

	switch hook.Type {
	case "pushover":
		_ = sendPushover(hook, host, changesStr, link)
		return
	}

	payload := map[string]any{
		"text": fmt.Sprintf("EchoState alert for %s\n%s\n%s", host, changesStr, link),
		"host": host,
		"changes": changes,
		"link": link,
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return
	}

	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest(http.MethodPost, hook.URL, bytes.NewBuffer(encoded))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	_, _ = client.Do(req)
}