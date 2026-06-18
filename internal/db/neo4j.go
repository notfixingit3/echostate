package db

import (
	"context"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/notfixingit3/echostate/internal/models"
)

type Neo4jClient struct {
	driver neo4j.DriverWithContext
}

func NewNeo4jClient(uri, user, pass string) (*Neo4jClient, error) {
	driver, err := neo4j.NewDriverWithContext(uri, neo4j.BasicAuth(user, pass, ""))
	if err != nil {
		return nil, err
	}
	return &Neo4jClient{driver: driver}, nil
}

func (c *Neo4jClient) Close(ctx context.Context) error {
	return c.driver.Close(ctx)
}

func (c *Neo4jClient) SyncSnapshot(ctx context.Context, target *models.Target, snapshot *models.Snapshot) error {
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	meta := graphSyncMeta{
		TargetID:   target.ID.String(),
		SnapshotID: snapshot.ID.String(),
		ScannedAt:  snapshot.ScannedAt.Unix(),
	}

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		// Merge Target
		targetQuery := `
			MERGE (t:Target {id: $target_id})
			SET t.host = $host, t.created_at = $created_at
		`
		_, err := tx.Run(ctx, targetQuery, map[string]any{
			"target_id":  target.ID.String(),
			"host":       target.Host,
			"created_at": target.CreatedAt.Unix(),
		})
		if err != nil {
			return nil, err
		}

		// Merge Snapshot
		snapQuery := `
			MERGE (s:Snapshot {id: $snap_id})
			SET s.scanned_at = $scanned_at, s.data_hash = $data_hash
			WITH s
			MATCH (t:Target {id: $target_id})
			MERGE (t)-[:HAS_SNAPSHOT]->(s)
		`
		_, err = tx.Run(ctx, snapQuery, map[string]any{
			"snap_id":    snapshot.ID.String(),
			"scanned_at": snapshot.ScannedAt.Unix(),
			"data_hash":  snapshot.DataHash,
			"target_id":  target.ID.String(),
		})
		if err != nil {
			return nil, err
		}

		// Add relationship for ASN if it exists
		if asnMap, ok := snapshot.RawData["asn"].(map[string]any); ok {
			if asnVal, ok := asnMap["asn"].(string); ok && asnVal != "" {
				asnQuery := `
					MERGE (a:ASN {number: $asn})
					WITH a
					MATCH (t:Target {id: $target_id})
					MERGE (t)-[r:HOSTED_ON]->(a)
					SET r.snapshot_id = $snapshot_id, r.scanned_at = $scanned_at
				`
				_, err = tx.Run(ctx, asnQuery, meta.with(map[string]any{"asn": asnVal}))
				if err != nil {
					return nil, err
				}
			}
		}
		
		// Add relationship for IP if it exists
		if asnMap, ok := snapshot.RawData["asn"].(map[string]any); ok {
			if ipVal, ok := asnMap["ip"].(string); ok && ipVal != "" {
				ipQuery := `
					MERGE (i:IP {address: $ip})
					WITH i
					MATCH (t:Target {id: $target_id})
					MERGE (t)-[r:RESOLVES_TO]->(i)
					SET r.snapshot_id = $snapshot_id, r.scanned_at = $scanned_at
				`
				_, err = tx.Run(ctx, ipQuery, meta.with(map[string]any{"ip": ipVal}))
				if err != nil {
					return nil, err
				}
			}
		}

		if tlsMap, ok := snapshot.RawData["tls"].(map[string]any); ok {
			if jarm, ok := tlsMap["jarm"].(string); ok && jarm != "" {
				jarmQuery := `
					MERGE (j:JARM {hash: $hash})
					WITH j
					MATCH (t:Target {id: $target_id})
					MERGE (t)-[r:HAS_JARM]->(j)
					SET r.snapshot_id = $snapshot_id, r.scanned_at = $scanned_at
				`
				_, err = tx.Run(ctx, jarmQuery, meta.with(map[string]any{"hash": jarm}))
				if err != nil {
					return nil, err
				}
			}

			if issuer, ok := tlsMap["issuer"].(string); ok && issuer != "" {
				issuerQuery := `
					MERGE (c:CertIssuer {name: $name})
					WITH c
					MATCH (t:Target {id: $target_id})
					MERGE (t)-[r:SIGNED_BY]->(c)
					SET r.snapshot_id = $snapshot_id, r.scanned_at = $scanned_at
				`
				_, err = tx.Run(ctx, issuerQuery, meta.with(map[string]any{"name": issuer}))
				if err != nil {
					return nil, err
				}
			}
		}

		if favMap, ok := snapshot.RawData["favicon"].(map[string]any); ok {
			if mmh3, ok := favMap["mmh3"].(string); ok && mmh3 != "" {
				favQuery := `
					MERGE (f:Favicon {mmh3: $mmh3})
					WITH f
					MATCH (t:Target {id: $target_id})
					MERGE (t)-[r:HAS_FAVICON]->(f)
					SET r.snapshot_id = $snapshot_id, r.scanned_at = $scanned_at
				`
				_, err = tx.Run(ctx, favQuery, meta.with(map[string]any{"mmh3": mmh3}))
				if err != nil {
					return nil, err
				}
			}
		}

		if err = syncBGPGraph(ctx, tx, meta, snapshot.RawData); err != nil {
			return nil, err
		}
		if err = syncPeeringGraph(ctx, tx, meta, snapshot.RawData); err != nil {
			return nil, err
		}
		if err = syncCTGraph(ctx, tx, meta, snapshot.RawData); err != nil {
			return nil, err
		}
		if err = syncDNSGraph(ctx, tx, meta, snapshot.RawData); err != nil {
			return nil, err
		}
		if err = syncTracerouteGraph(ctx, tx, meta, snapshot.RawData); err != nil {
			return nil, err
		}
		if err = syncCertSANGraph(ctx, tx, meta, snapshot.RawData); err != nil {
			return nil, err
		}

		if err = syncIntelEvents(ctx, tx, target.ID.String(), snapshot); err != nil {
			return nil, err
		}
		if enrich, ok := snapshot.RawData["enrichment"].(map[string]any); ok {
			if err = syncEnrichmentHits(ctx, tx, target.ID.String(), snapshot.ID.String(), snapshot.ScannedAt.Unix(), enrich); err != nil {
				return nil, err
			}
		}

		return nil, nil
	})

	return err
}
