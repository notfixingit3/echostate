package db

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/notfixingit3/echostate/internal/models"
)

func syncIntelEvents(ctx context.Context, tx neo4j.ManagedTransaction, targetID string, snapshot *models.Snapshot) error {
	if snapshot == nil || len(snapshot.ChangeDetails) == 0 {
		return nil
	}

	snapID := snapshot.ID.String()
	detectedAt := snapshot.ScannedAt.Unix()

	for order, entry := range snapshot.ChangeDetails {
		eventID := intelEventID(snapID, entry)
		_, err := tx.Run(ctx, `
			MERGE (e:IntelEvent {id: $id})
			SET e.type = $type,
				e.severity = $severity,
				e.summary = $summary,
				e.field = $field,
				e.detail = $detail,
				e.detected_at = $detected_at,
				e.snapshot_id = $snapshot_id,
				e.source = 'diff'
			WITH e
			MATCH (s:Snapshot {id: $snapshot_id})
			MERGE (s)-[:DETECTED {order: $order}]->(e)
			WITH e
			MATCH (t:Target {id: $target_id})
			MERGE (t)-[:HAS_INTEL_EVENT]->(e)
		`, map[string]any{
			"id":          eventID,
			"type":        entry.Type,
			"severity":    entry.Severity,
			"summary":     entry.Summary,
			"field":       entry.Field,
			"detail":      entry.Detail,
			"detected_at": detectedAt,
			"snapshot_id": snapID,
			"order":       order,
			"target_id":   targetID,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func syncEnrichmentHits(ctx context.Context, tx neo4j.ManagedTransaction, targetID, snapshotID string, detectedAt int64, enrichment map[string]any) error {
	if len(enrichment) == 0 {
		return nil
	}

	for provider, raw := range enrichment {
		if strings.HasSuffix(provider, "_error") || raw == nil {
			continue
		}
		summary := enrichmentSummary(provider, raw)
		if summary == "" {
			continue
		}

		hitID := enrichmentHitID(snapshotID, provider)
		_, err := tx.Run(ctx, `
			MERGE (h:EnrichmentHit {id: $id})
			SET h.provider = $provider,
				h.summary = $summary,
				h.snapshot_id = $snapshot_id,
				h.detected_at = $detected_at,
				h.source = 'enrichment'
			WITH h
			MATCH (s:Snapshot {id: $snapshot_id})
			MERGE (s)-[:ENRICHED_BY]->(h)
			WITH h
			MATCH (t:Target {id: $target_id})
			MERGE (t)-[:HAS_ENRICHMENT_HIT]->(h)
		`, map[string]any{
			"id":          hitID,
			"provider":    provider,
			"summary":     summary,
			"snapshot_id": snapshotID,
			"detected_at": detectedAt,
			"target_id":   targetID,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// SyncEnrichment writes enrichment correlation hits for a snapshot into Neo4j.
func (c *Neo4jClient) SyncEnrichment(ctx context.Context, targetID, snapshotID string, detectedAt int64, enrichment map[string]any) error {
	if c == nil || len(enrichment) == 0 {
		return nil
	}

	session := c.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		return nil, syncEnrichmentHits(ctx, tx, targetID, snapshotID, detectedAt, enrichment)
	})
	return err
}

// GetIntelEvents returns recent diff and enrichment intel for a target.
func (c *Neo4jClient) GetIntelEvents(ctx context.Context, targetID string, limit int) ([]models.GraphIntelEvent, error) {
	if limit <= 0 {
		limit = 50
	}

	session := c.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.Run(ctx, `
		MATCH (t:Target {id: $target_id})-[rel:HAS_INTEL_EVENT|HAS_ENRICHMENT_HIT]->(e)
		RETURN e
		ORDER BY coalesce(e.detected_at, 0) DESC, e.summary ASC
		LIMIT $limit
	`, map[string]any{
		"target_id": targetID,
		"limit":     limit,
	})
	if err != nil {
		return nil, err
	}

	records, err := result.Collect(ctx)
	if err != nil {
		return nil, err
	}

	events := make([]models.GraphIntelEvent, 0, len(records))
	for _, record := range records {
		raw, _ := record.Get("e")
		node, ok := raw.(neo4j.Node)
		if !ok {
			continue
		}
		event, ok := graphIntelEventFromNode(node)
		if ok {
			events = append(events, event)
		}
	}
	return events, nil
}

func graphIntelEventFromNode(node neo4j.Node) (models.GraphIntelEvent, bool) {
	props := node.Props
	id := stringProp(props, "id")
	if id == "" {
		return models.GraphIntelEvent{}, false
	}

	label := primaryNeo4jLabel(node.Labels)
	source := stringProp(props, "source")
	if source == "" && label == "EnrichmentHit" {
		source = "enrichment"
	}
	eventType := stringProp(props, "type")
	if eventType == "" {
		if label == "EnrichmentHit" {
			eventType = "enrichment_hit"
		} else {
			eventType = "intel_event"
		}
	}
	summary := stringProp(props, "summary")
	if source == "enrichment" {
		provider := stringProp(props, "provider")
		if summary == "" && provider != "" {
			summary = provider + " enrichment"
		}
	}

	return models.GraphIntelEvent{
		ID:         id,
		Type:       eventType,
		Severity:   stringProp(props, "severity"),
		Summary:    summary,
		Field:      stringProp(props, "field"),
		Detail:     stringProp(props, "detail"),
		SnapshotID: stringProp(props, "snapshot_id"),
		DetectedAt: int64(intProp(props, "detected_at")),
		Source:     source,
	}, true
}

func intelEventID(snapshotID string, entry models.ChangeDetail) string {
	payload := snapshotID + "|" + entry.Type + "|" + entry.Field + "|" + entry.Summary
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:16])
}

func enrichmentHitID(snapshotID, provider string) string {
	sum := sha256.Sum256([]byte(snapshotID + "|" + provider))
	return "enrich:" + hex.EncodeToString(sum[:12])
}

func enrichmentSummary(provider string, raw any) string {
	switch provider {
	case "shodan":
		if m, ok := raw.(map[string]any); ok {
			if host, ok := m["host"].(map[string]any); ok {
				return fmt.Sprintf("Shodan host %s (%s)", truncateLabel(fmt.Sprint(host["ip"]), 24), truncateLabel(fmt.Sprint(host["org"]), 24))
			}
			if search, ok := m["favicon_search"].(map[string]any); ok {
				return fmt.Sprintf("Shodan: %v favicon matches", search["total"])
			}
		}
	case "censys":
		if m, ok := raw.(map[string]any); ok {
			if host, ok := m["host"].(map[string]any); ok {
				return fmt.Sprintf("Censys host %s", truncateLabel(fmt.Sprint(host["ip"]), 24))
			}
			if search, ok := m["jarm_search"].(map[string]any); ok {
				return fmt.Sprintf("Censys JARM %v", truncateLabel(fmt.Sprint(search["query"]), 24))
			}
		}
	case "wayback":
		if m, ok := raw.(map[string]any); ok {
			return fmt.Sprintf("Wayback: %v archived URLs", m["total"])
		}
	case "virustotal":
		if m, ok := raw.(map[string]any); ok {
			return fmt.Sprintf("VirusTotal: %v passive DNS records", m["total"])
		}
	case "hibp":
		if items, ok := raw.([]any); ok {
			return fmt.Sprintf("HIBP: %d email breach hit(s)", len(items))
		}
	case "riskiq":
		return "RiskIQ: passive DNS enrichment"
	default:
		encoded, err := json.Marshal(raw)
		if err == nil {
			return provider + ": " + truncateLabel(string(encoded), 80)
		}
	}
	return provider + " enrichment"
}
