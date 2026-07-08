package enrichment

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/notfixingit3/echostate/internal/config"
	"github.com/notfixingit3/echostate/internal/db"
	"github.com/notfixingit3/echostate/internal/diff"
	"github.com/notfixingit3/echostate/internal/models"
	"github.com/notfixingit3/echostate/internal/webhooks"
)

// PendingPayload is stored on new snapshots while async enrichment runs.
func PendingPayload() map[string]any {
	return map[string]any{
		"status":     "pending",
		"started_at": time.Now().UTC().Format(time.RFC3339),
	}
}

// ShouldEnqueueEnrichment reports whether async enrichment should run for a host.
func ShouldEnqueueEnrichment(settings config.SystemSettings, host string) bool {
	if hasEnrichmentKeys(settings) {
		return true
	}
	return resolveHost(map[string]any{"host": host}) != ""
}

func hasEnrichmentKeys(settings config.SystemSettings) bool {
	return settings.ShodanAPIKey != "" ||
		(settings.CensysAPIID != "" && settings.CensysAPISecret != "") ||
		settings.HIBPAPIKey != "" ||
		(settings.RiskIQAPIUser != "" && settings.RiskIQAPIKey != "") ||
		settings.VirusTotalAPIKey != ""
}

// EnrichmentComplete reports whether enrichment finished (success or partial).
func EnrichmentComplete(enrichment map[string]any) bool {
	if len(enrichment) == 0 {
		return false
	}
	status := strings.TrimSpace(fmt.Sprint(enrichment["status"]))
	return status == "completed" || status == "partial"
}

// EnrichSnapshot augments snapshot raw_data with passive third-party correlation.
func EnrichSnapshot(ctx context.Context, database *db.DB, neo4j *db.Neo4jClient, snapshotID uuid.UUID) error {
	settings := config.GetSettings()

	var targetID uuid.UUID
	var scannedAt time.Time
	var raw []byte
	var changes []string
	var changeDetailsRaw []byte
	err := database.Pool.QueryRow(ctx, `
		SELECT target_id, scanned_at, raw_data, COALESCE(changes, '{}'), COALESCE(change_details, '[]')
		FROM snapshots WHERE id = $1
	`, snapshotID).Scan(&targetID, &scannedAt, &raw, &changes, &changeDetailsRaw)
	if err != nil {
		return fmt.Errorf("load snapshot: %w", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return fmt.Errorf("decode snapshot: %w", err)
	}
	if !ShouldEnqueueEnrichment(settings, stringField(payload, "host")) {
		return nil
	}

	prevEnrichment, _ := loadPreviousEnrichment(ctx, database, targetID, snapshotID)

	enrichment := map[string]any{}
	var hadError bool

	if settings.ShodanAPIKey != "" {
		if shodan, err := queryShodan(ctx, settings.ShodanAPIKey, payload); err == nil && len(shodan) > 0 {
			enrichment["shodan"] = shodan
		} else if err != nil {
			enrichment["shodan_error"] = FriendlyError(err)
			hadError = true
		}
	}
	if settings.CensysAPIID != "" && settings.CensysAPISecret != "" {
		if censys, err := queryCensys(ctx, settings.CensysAPIID, settings.CensysAPISecret, payload); err == nil && len(censys) > 0 {
			enrichment["censys"] = censys
		} else if err != nil {
			enrichment["censys_error"] = FriendlyError(err)
			hadError = true
		}
	}
	if settings.HIBPAPIKey != "" {
		if hibp, err := queryHIBP(ctx, settings.HIBPAPIKey, payload); err == nil && len(hibp) > 0 {
			enrichment["hibp"] = hibp
		} else if err != nil {
			enrichment["hibp_error"] = FriendlyError(err)
			hadError = true
		}
	}
	if settings.RiskIQAPIUser != "" && settings.RiskIQAPIKey != "" {
		if riskiq, err := queryRiskIQ(ctx, settings.RiskIQAPIUser, settings.RiskIQAPIKey, payload); err == nil && len(riskiq) > 0 {
			enrichment["riskiq"] = riskiq
		} else if err != nil {
			enrichment["riskiq_error"] = FriendlyError(err)
			hadError = true
		}
	}
	if wayback, err := queryWayback(ctx, payload); err == nil && len(wayback) > 0 {
		enrichment["wayback"] = wayback
	} else if err != nil {
		enrichment["wayback_error"] = FriendlyError(err)
		hadError = true
	}
	if settings.VirusTotalAPIKey != "" {
		if vt, err := queryVirusTotal(ctx, settings.VirusTotalAPIKey, payload); err == nil && len(vt) > 0 {
			enrichment["virustotal"] = vt
		} else if err != nil {
			enrichment["virustotal_error"] = FriendlyError(err)
			hadError = true
		}
	}

	status := "completed"
	if hadError && len(enrichment) <= 2 {
		status = "partial"
	} else if hadError {
		status = "partial"
	}
	if len(enrichment) == 0 {
		status = "completed"
	}

	enrichment["status"] = status
	enrichment["completed_at"] = time.Now().UTC().Format(time.RFC3339)

	payload["enrichment"] = enrichment
	updated, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	diffEntries := DiffEnrichment(prevEnrichment, enrichment)
	var changeDetails []models.ChangeDetail
	if len(changeDetailsRaw) > 0 {
		_ = json.Unmarshal(changeDetailsRaw, &changeDetails)
	}
	for _, entry := range diffEntries {
		changeDetails = append(changeDetails, models.ChangeDetail{
			Type:     entry.Type,
			Severity: entry.Severity,
			Summary:  entry.Summary,
			Field:    entry.Field,
			Detail:   entry.Detail,
		})
		changes = append(changes, diff.FormatSummary(entry))
	}

	detailsJSON, err := json.Marshal(changeDetails)
	if err != nil {
		return err
	}

	_, err = database.Pool.Exec(ctx, `
		UPDATE snapshots
		SET raw_data = $1, changes = $2, change_details = $3
		WHERE id = $4
	`, updated, changes, detailsJSON, snapshotID)
	if err != nil {
		return fmt.Errorf("store enrichment: %w", err)
	}
	slog.Default().Info("enrichment: updated snapshot", slog.String("component", "enrichment"), slog.String("snapshot_id", snapshotID.String()), slog.String("status", status))

	if len(diffEntries) > 0 {
		snapshot := &models.Snapshot{
			ID:            snapshotID,
			TargetID:      targetID,
			Changes:       changes,
			ChangeDetails: changeDetails,
		}
		go webhooks.Dispatch(context.Background(), database, stringField(payload, "host"), snapshot, "")
	}

	if neo4j != nil {
		go func() {
			_ = neo4j.SyncEnrichment(
				context.Background(),
				targetID.String(),
				snapshotID.String(),
				scannedAt.Unix(),
				enrichment,
			)
		}()
	}
	return nil
}

func loadPreviousEnrichment(ctx context.Context, database *db.DB, targetID, snapshotID uuid.UUID) (map[string]any, error) {
	var raw []byte
	err := database.Pool.QueryRow(ctx, `
		SELECT raw_data->'enrichment'
		FROM snapshots
		WHERE target_id = $1 AND id <> $2
		ORDER BY scanned_at DESC
		LIMIT 1
	`, targetID, snapshotID).Scan(&raw)
	if err != nil || len(raw) == 0 || string(raw) == "null" {
		return nil, err
	}
	var enrichment map[string]any
	if err := json.Unmarshal(raw, &enrichment); err != nil {
		return nil, err
	}
	return enrichment, nil
}

func httpClient() *http.Client {
	return &http.Client{Timeout: config.EnrichmentHTTPTimeout()}
}
