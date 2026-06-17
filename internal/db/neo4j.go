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
					MERGE (t)-[:HOSTED_ON]->(a)
				`
				_, err = tx.Run(ctx, asnQuery, map[string]any{
					"asn":       asnVal,
					"target_id": target.ID.String(),
				})
			}
		}
		
		// Add relationship for IP if it exists
		if asnMap, ok := snapshot.RawData["asn"].(map[string]any); ok {
			if ipVal, ok := asnMap["ip"].(string); ok && ipVal != "" {
				ipQuery := `
					MERGE (i:IP {address: $ip})
					WITH i
					MATCH (t:Target {id: $target_id})
					MERGE (t)-[:RESOLVES_TO]->(i)
				`
				_, err = tx.Run(ctx, ipQuery, map[string]any{
					"ip":        ipVal,
					"target_id": target.ID.String(),
				})
			}
		}

		return nil, err
	})

	return err
}
