package enrichment

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/notfixingit3/echostate/internal/config"
	"github.com/notfixingit3/echostate/internal/db"
)

const httpTimeout = 12 * time.Second

// EnrichSnapshot augments snapshot raw_data with passive third-party correlation.
func EnrichSnapshot(ctx context.Context, database *db.DB, snapshotID uuid.UUID) error {
	settings := config.GetSettings()
	if !hasEnrichmentKeys(settings) {
		return nil
	}

	var raw []byte
	err := database.Pool.QueryRow(ctx, `SELECT raw_data FROM snapshots WHERE id = $1`, snapshotID).Scan(&raw)
	if err != nil {
		return fmt.Errorf("load snapshot: %w", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return fmt.Errorf("decode snapshot: %w", err)
	}

	enrichment := map[string]any{}
	if settings.ShodanAPIKey != "" {
		if shodan, err := queryShodan(ctx, settings.ShodanAPIKey, payload); err == nil && len(shodan) > 0 {
			enrichment["shodan"] = shodan
		} else if err != nil {
			enrichment["shodan_error"] = err.Error()
		}
	}
	if settings.CensysAPIID != "" && settings.CensysAPISecret != "" {
		if censys, err := queryCensys(ctx, settings.CensysAPIID, settings.CensysAPISecret, payload); err == nil && len(censys) > 0 {
			enrichment["censys"] = censys
		} else if err != nil {
			enrichment["censys_error"] = err.Error()
		}
	}
	if settings.HIBPAPIKey != "" {
		if hibp, err := queryHIBP(ctx, settings.HIBPAPIKey, payload); err == nil && len(hibp) > 0 {
			enrichment["hibp"] = hibp
		} else if err != nil {
			enrichment["hibp_error"] = err.Error()
		}
	}
	if settings.RiskIQAPIUser != "" && settings.RiskIQAPIKey != "" {
		if riskiq, err := queryRiskIQ(ctx, settings.RiskIQAPIUser, settings.RiskIQAPIKey, payload); err == nil && len(riskiq) > 0 {
			enrichment["riskiq"] = riskiq
		} else if err != nil {
			enrichment["riskiq_error"] = err.Error()
		}
	}

	if len(enrichment) == 0 {
		return nil
	}

	payload["enrichment"] = enrichment
	updated, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	_, err = database.Pool.Exec(ctx, `UPDATE snapshots SET raw_data = $1 WHERE id = $2`, updated, snapshotID)
	if err != nil {
		return fmt.Errorf("store enrichment: %w", err)
	}
	log.Printf("enrichment: updated snapshot %s", snapshotID)
	return nil
}

func hasEnrichmentKeys(settings config.SystemSettings) bool {
	return settings.ShodanAPIKey != "" ||
		(settings.CensysAPIID != "" && settings.CensysAPISecret != "") ||
		settings.HIBPAPIKey != "" ||
		(settings.RiskIQAPIUser != "" && settings.RiskIQAPIKey != "")
}

func queryShodan(ctx context.Context, apiKey string, payload map[string]any) (map[string]any, error) {
	favicon, _ := payload["favicon"].(map[string]any)
	query := strings.TrimSpace(fmt.Sprint(favicon["mmh3"]))
	if query == "" {
		query = strings.TrimSpace(fmt.Sprint(favicon["shodan"]))
	}
	if query == "" {
		return nil, nil
	}

	endpoint := fmt.Sprintf("https://api.shodan.io/shodan/host/search?key=%s&query=http.favicon.hash:%s",
		url.QueryEscape(apiKey), url.QueryEscape(query))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	resp, err := httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("shodan HTTP %d", resp.StatusCode)
	}

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	return map[string]any{
		"query":   query,
		"total":   body["total"],
		"matches": truncateMatches(body["matches"]),
	}, nil
}

func queryCensys(ctx context.Context, apiID, apiSecret string, payload map[string]any) (map[string]any, error) {
	tlsMap, _ := payload["tls"].(map[string]any)
	jarm := strings.TrimSpace(fmt.Sprint(tlsMap["jarm"]))
	if jarm == "" {
		return nil, nil
	}

	endpoint := "https://search.censys.io/api/v2/hosts/search"
	body := map[string]any{
		"q":        fmt.Sprintf("services.jarm.fingerprint: %s", jarm),
		"per_page": 5,
	}
	encoded, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(string(encoded)))
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(apiID, apiSecret)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("censys HTTP %d", resp.StatusCode)
	}

	var parsed map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, err
	}
	return map[string]any{
		"query":  jarm,
		"result": parsed["result"],
	}, nil
}

func truncateMatches(raw any) any {
	items, ok := raw.([]any)
	if !ok || len(items) <= 5 {
		return raw
	}
	return items[:5]
}

func httpClient() *http.Client {
	return &http.Client{Timeout: httpTimeout}
}