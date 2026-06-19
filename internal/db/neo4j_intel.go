package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

const (
	graphSecurityContactLimit = 20
	graphCAARecordLimit       = 20
	graphWaybackURLLimit      = 30
)

func syncCrawlIntelGraph(ctx context.Context, tx neo4j.ManagedTransaction, meta graphSyncMeta, rawData map[string]any) error {
	crawlMap, ok := rawData["crawl"].(map[string]any)
	if !ok {
		return nil
	}

	securityTxt, ok := crawlMap["security_txt"].(map[string]any)
	if !ok {
		return nil
	}

	contacts, ok := securityTxt["contacts"].([]any)
	if !ok || len(contacts) == 0 {
		return nil
	}

	count := 0
	for _, item := range contacts {
		if count >= graphSecurityContactLimit {
			break
		}
		contact := strings.TrimSpace(fmt.Sprint(item))
		if contact == "" {
			continue
		}

		_, err := tx.Run(ctx, `
			MERGE (c:SecurityContact {address: $address})
			SET c.source = $source
			WITH c
			MATCH (t:Target {id: $target_id})
			MERGE (t)-[r:HAS_SECURITY_CONTACT]->(c)
			SET r.snapshot_id = $snapshot_id, r.scanned_at = $scanned_at
		`, meta.with(map[string]any{
			"address": contact,
			"source":  stringProp(securityTxt, "source"),
		}))
		if err != nil {
			return err
		}
		count++
	}
	return nil
}

func syncMailDNSIntelGraph(ctx context.Context, tx neo4j.ManagedTransaction, meta graphSyncMeta, rawData map[string]any) error {
	dnsMap, ok := rawData["dns"].(map[string]any)
	if !ok {
		return nil
	}

	var err error

	if mtaSts, ok := dnsMap["MTA_STS"].(map[string]any); ok {
		mode := strings.ToLower(stringProp(mtaSts, "mode"))
		if mode != "" {
			_, err = tx.Run(ctx, `
				MERGE (m:MTASTSPolicy {mode: $mode})
				SET m.policy_url = $policy_url, m.version = $version
				WITH m
				MATCH (t:Target {id: $target_id})
				MERGE (t)-[r:HAS_MTA_STS]->(m)
				SET r.snapshot_id = $snapshot_id, r.scanned_at = $scanned_at
			`, meta.with(map[string]any{
				"mode":       mode,
				"policy_url": stringProp(mtaSts, "policy_url"),
				"version":    stringProp(mtaSts, "version"),
			}))
			if err != nil {
				return err
			}
		}
	}

	if dnssec, ok := dnsMap["DNSSEC"].(map[string]any); ok {
		status := strings.ToLower(stringProp(dnssec, "status"))
		if status != "" {
			_, err = tx.Run(ctx, `
				MERGE (d:DNSSECStatus {status: $status})
				SET d.zone = $zone, d.signed = $signed
				WITH d
				MATCH (t:Target {id: $target_id})
				MERGE (t)-[r:HAS_DNSSEC]->(d)
				SET r.snapshot_id = $snapshot_id, r.scanned_at = $scanned_at
			`, meta.with(map[string]any{
				"status": status,
				"zone":   stringProp(dnssec, "zone"),
				"signed": boolProp(dnssec, "signed"),
			}))
			if err != nil {
				return err
			}
		}
	}

	if caaAny, ok := dnsMap["CAA"].([]any); ok {
		count := 0
		for _, item := range caaAny {
			if count >= graphCAARecordLimit {
				break
			}
			record, ok := item.(map[string]any)
			if !ok {
				continue
			}
			tag := stringProp(record, "tag")
			value := stringProp(record, "value")
			if tag == "" || value == "" {
				continue
			}
			recordKey := fmt.Sprintf("%s:%s", tag, value)
			_, err = tx.Run(ctx, `
				MERGE (c:CAARecord {record: $record})
				SET c.tag = $tag, c.value = $value, c.flags = $flags
				WITH c
				MATCH (t:Target {id: $target_id})
				MERGE (t)-[r:HAS_CAA]->(c)
				SET r.snapshot_id = $snapshot_id, r.scanned_at = $scanned_at
			`, meta.with(map[string]any{
				"record": recordKey,
				"tag":    tag,
				"value":  value,
				"flags":  intProp(record, "flags"),
			}))
			if err != nil {
				return err
			}
			count++
		}
	}

	if bimi, ok := dnsMap["BIMI"].(map[string]any); ok {
		record := stringProp(bimi, "record")
		if record != "" {
			_, err = tx.Run(ctx, `
				MERGE (b:BIMIRecord {record: $record})
				SET b.logo_url = $logo_url, b.authority = $authority
				WITH b
				MATCH (t:Target {id: $target_id})
				MERGE (t)-[r:HAS_BIMI]->(b)
				SET r.snapshot_id = $snapshot_id, r.scanned_at = $scanned_at
			`, meta.with(map[string]any{
				"record":    record,
				"logo_url":  stringProp(bimi, "logo_url"),
				"authority": stringProp(bimi, "authority"),
			}))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func syncWaybackGraph(ctx context.Context, tx neo4j.ManagedTransaction, meta graphSyncMeta, rawData map[string]any) error {
	enrichment, ok := rawData["enrichment"].(map[string]any)
	if !ok {
		return nil
	}
	wayback, ok := enrichment["wayback"].(map[string]any)
	if !ok {
		return nil
	}
	urlsAny, ok := wayback["urls"].([]any)
	if !ok || len(urlsAny) == 0 {
		return nil
	}

	count := 0
	for _, item := range urlsAny {
		if count >= graphWaybackURLLimit {
			break
		}
		row, ok := item.(map[string]any)
		if !ok {
			continue
		}
		url := stringProp(row, "url")
		if url == "" {
			continue
		}
		timestamp := stringProp(row, "timestamp")

		_, err := tx.Run(ctx, `
			MERGE (w:WaybackURL {url: $url})
			SET w.timestamp = $timestamp
			WITH w
			MATCH (t:Target {id: $target_id})
			MERGE (t)-[r:ARCHIVED_AT]->(w)
			SET r.snapshot_id = $snapshot_id, r.scanned_at = $scanned_at
		`, meta.with(map[string]any{
			"url":       url,
			"timestamp": timestamp,
		}))
		if err != nil {
			return err
		}
		count++
	}
	return nil
}