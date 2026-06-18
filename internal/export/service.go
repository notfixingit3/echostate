package export

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/notfixingit3/echostate/internal/db"
	"github.com/notfixingit3/echostate/internal/models"
	"github.com/notfixingit3/echostate/internal/scanner"
	"github.com/notfixingit3/echostate/internal/version"
)

type Options struct {
	IncludeBlobs   bool
	IncludeConfig  bool
}

// BuildBundle reads investigation data from Postgres for portable export.
func BuildBundle(ctx context.Context, database *db.DB, opts Options) (*models.ExportBundle, error) {
	bundle := &models.ExportBundle{
		FormatVersion:    models.ExportBundleFormatVersion,
		ExportedAt:       time.Now().UTC(),
		EchoStateVersion: version.Version,
		Targets:          []models.ExportTarget{},
		Snapshots:        []models.ExportSnapshot{},
		SnapshotBlobs:    []models.ExportSnapshotBlob{},
		Collections:      []models.ExportCollection{},
		CollectionMembers: []models.ExportCollectionMember{},
		Notes:            []models.InvestigationNote{},
		GraphViews:       []models.SavedGraphView{},
		Webhooks:         []models.Webhook{},
	}

	targetRows, err := database.Pool.Query(ctx, `
		SELECT id, host, COALESCE(normalized_host, ''), tags, created_at
		FROM targets
		ORDER BY created_at ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list targets: %w", err)
	}
	defer targetRows.Close()

	for targetRows.Next() {
		var item models.ExportTarget
		if err := targetRows.Scan(&item.ID, &item.Host, &item.NormalizedHost, &item.Tags, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan target: %w", err)
		}
		if item.Tags == nil {
			item.Tags = []string{}
		}
		bundle.Targets = append(bundle.Targets, item)
	}
	if err := targetRows.Err(); err != nil {
		return nil, err
	}

	snapshotRows, err := database.Pool.Query(ctx, `
		SELECT id, target_id, scanned_at, last_seen, data_hash, raw_data, changes,
			COALESCE(change_details, '[]'), COALESCE(client_ip, ''),
			pwhois_data, pwhois_looked_up_at, pwhois_origin_as, pwhois_org_name,
			pwhois_country_code, pwhois_city, pwhois_prefix
		FROM snapshots
		ORDER BY scanned_at ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list snapshots: %w", err)
	}
	defer snapshotRows.Close()

	for snapshotRows.Next() {
		var item models.ExportSnapshot
		var changeDetailsJSON []byte
		if err := snapshotRows.Scan(
			&item.ID, &item.TargetID, &item.ScannedAt, &item.LastSeen, &item.DataHash, &item.RawData,
			&item.Changes, &changeDetailsJSON, &item.ClientIP,
			&item.PwhoisData, &item.PwhoisLookedUp, &item.PwhoisOriginAS, &item.PwhoisOrgName,
			&item.PwhoisCountry, &item.PwhoisCity, &item.PwhoisPrefix,
		); err != nil {
			return nil, fmt.Errorf("scan snapshot: %w", err)
		}
		if item.RawData == nil {
			item.RawData = map[string]any{}
		}
		if item.Changes == nil {
			item.Changes = []string{}
		}
		if len(changeDetailsJSON) > 0 {
			_ = json.Unmarshal(changeDetailsJSON, &item.ChangeDetails)
		}
		if item.ChangeDetails == nil {
			item.ChangeDetails = []models.ChangeDetail{}
		}
		bundle.Snapshots = append(bundle.Snapshots, item)
	}
	if err := snapshotRows.Err(); err != nil {
		return nil, err
	}

	if opts.IncludeBlobs {
		blobRows, err := database.Pool.Query(ctx, `
			SELECT id, snapshot_id, kind, content_type, data
			FROM snapshot_blobs
			ORDER BY created_at ASC
		`)
		if err != nil {
			return nil, fmt.Errorf("list snapshot blobs: %w", err)
		}
		defer blobRows.Close()

		for blobRows.Next() {
			var item models.ExportSnapshotBlob
			var data []byte
			if err := blobRows.Scan(&item.ID, &item.SnapshotID, &item.Kind, &item.ContentType, &data); err != nil {
				return nil, fmt.Errorf("scan snapshot blob: %w", err)
			}
			item.DataBase64 = base64.StdEncoding.EncodeToString(data)
			bundle.SnapshotBlobs = append(bundle.SnapshotBlobs, item)
		}
		if err := blobRows.Err(); err != nil {
			return nil, err
		}
	}

	collectionRows, err := database.Pool.Query(ctx, `
		SELECT id, name, description, created_at, updated_at
		FROM target_collections
		ORDER BY name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list collections: %w", err)
	}
	defer collectionRows.Close()

	for collectionRows.Next() {
		var item models.ExportCollection
		if err := collectionRows.Scan(&item.ID, &item.Name, &item.Description, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan collection: %w", err)
		}
		bundle.Collections = append(bundle.Collections, item)
	}
	if err := collectionRows.Err(); err != nil {
		return nil, err
	}

	memberRows, err := database.Pool.Query(ctx, `
		SELECT collection_id, target_id, added_at
		FROM target_collection_members
		ORDER BY added_at ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list collection members: %w", err)
	}
	defer memberRows.Close()

	for memberRows.Next() {
		var item models.ExportCollectionMember
		if err := memberRows.Scan(&item.CollectionID, &item.TargetID, &item.AddedAt); err != nil {
			return nil, fmt.Errorf("scan collection member: %w", err)
		}
		bundle.CollectionMembers = append(bundle.CollectionMembers, item)
	}
	if err := memberRows.Err(); err != nil {
		return nil, err
	}

	noteRows, err := database.Pool.Query(ctx, `
		SELECT n.id, n.title, n.body, n.status, n.target_id, n.snapshot_id, n.collection_id,
			n.graph_node_id, n.graph_node_label, n.graph_node_type, n.refs,
			n.trashed_at, n.archived_at, n.created_at, n.updated_at
		FROM investigation_notes n
		ORDER BY n.updated_at ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list notes: %w", err)
	}
	defer noteRows.Close()

	for noteRows.Next() {
		var item models.InvestigationNote
		var refsJSON []byte
		if err := noteRows.Scan(
			&item.ID, &item.Title, &item.Body, &item.Status,
			&item.TargetID, &item.SnapshotID, &item.CollectionID,
			&item.GraphNodeID, &item.GraphNodeLabel, &item.GraphNodeType, &refsJSON,
			&item.TrashedAt, &item.ArchivedAt, &item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan note: %w", err)
		}
		if err := json.Unmarshal(refsJSON, &item.References); err != nil {
			return nil, fmt.Errorf("decode note refs: %w", err)
		}
		if item.References == nil {
			item.References = []models.NoteReference{}
		}
		bundle.Notes = append(bundle.Notes, item)
	}
	if err := noteRows.Err(); err != nil {
		return nil, err
	}

	viewRows, err := database.Pool.Query(ctx, `
		SELECT id, name, description, view_mode, target_id, vantage_filter,
			snapshot_id, compare_snapshot_id, compare_mode, pinned_nodes,
			selected_node_id, created_at, updated_at
		FROM graph_views
		ORDER BY updated_at ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list graph views: %w", err)
	}
	defer viewRows.Close()

	for viewRows.Next() {
		var item models.SavedGraphView
		var pinnedJSON []byte
		if err := viewRows.Scan(
			&item.ID, &item.Name, &item.Description, &item.ViewMode, &item.TargetID,
			&item.VantageFilter, &item.SnapshotID, &item.CompareSnapshotID, &item.CompareMode,
			&pinnedJSON, &item.SelectedNodeID, &item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan graph view: %w", err)
		}
		item.PinnedNodes = map[string]models.GraphPinnedNode{}
		if len(pinnedJSON) > 0 {
			_ = json.Unmarshal(pinnedJSON, &item.PinnedNodes)
		}
		if item.PinnedNodes == nil {
			item.PinnedNodes = map[string]models.GraphPinnedNode{}
		}
		bundle.GraphViews = append(bundle.GraphViews, item)
	}
	if err := viewRows.Err(); err != nil {
		return nil, err
	}

	if opts.IncludeConfig {
		webhookRows, err := database.Pool.Query(ctx, `
			SELECT id, name, type, url, config, enabled, created_at, updated_at
			FROM webhooks
			ORDER BY name ASC
		`)
		if err != nil {
			return nil, fmt.Errorf("list webhooks: %w", err)
		}
		defer webhookRows.Close()

		for webhookRows.Next() {
			var item models.Webhook
			var configJSON []byte
			if err := webhookRows.Scan(
				&item.ID, &item.Name, &item.Type, &item.URL, &configJSON, &item.Enabled,
				&item.CreatedAt, &item.UpdatedAt,
			); err != nil {
				return nil, fmt.Errorf("scan webhook: %w", err)
			}
			if len(configJSON) > 0 {
				_ = json.Unmarshal(configJSON, &item.Config)
			}
			if item.Config == nil {
				item.Config = map[string]any{}
			}
			bundle.Webhooks = append(bundle.Webhooks, item)
		}
		if err := webhookRows.Err(); err != nil {
			return nil, err
		}

		var settingsJSON []byte
		err = database.Pool.QueryRow(ctx, `SELECT value FROM settings WHERE key = 'app_settings'`).Scan(&settingsJSON)
		if err == nil {
			settings := map[string]any{}
			if err := json.Unmarshal(settingsJSON, &settings); err == nil {
				bundle.Settings = settings
			}
		} else if err != pgx.ErrNoRows {
			return nil, fmt.Errorf("load settings: %w", err)
		}
	}

	return bundle, nil
}

type ImportOptions struct {
	Conflict     string
	RebuildGraph bool
}

type GraphSyncer interface {
	SyncSnapshot(ctx context.Context, target *models.Target, snapshot *models.Snapshot) error
}

// ImportBundle upserts portable investigation data into Postgres.
func ImportBundle(
	ctx context.Context,
	database *db.DB,
	neo4j GraphSyncer,
	bundle *models.ExportBundle,
	opts ImportOptions,
) (*models.ImportResult, error) {
	if bundle == nil {
		return nil, fmt.Errorf("bundle is required")
	}
	if bundle.FormatVersion != models.ExportBundleFormatVersion {
		return nil, fmt.Errorf("unsupported bundle format version %d", bundle.FormatVersion)
	}

	conflict := strings.ToLower(strings.TrimSpace(opts.Conflict))
	if conflict == "" {
		conflict = "overwrite"
	}
	if conflict != "skip" && conflict != "overwrite" {
		return nil, fmt.Errorf("invalid conflict strategy")
	}

	result := &models.ImportResult{}
	targetIDMap := map[uuid.UUID]uuid.UUID{}
	importedSnapshotIDs := []uuid.UUID{}

	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	for _, target := range bundle.Targets {
		normalized := target.NormalizedHost
		if normalized == "" {
			normalized = scanner.NormalizeHost(target.Host)
		}
		tags := target.Tags
		if tags == nil {
			tags = []string{}
		}

		var existingID uuid.UUID
		err := tx.QueryRow(ctx, `SELECT id FROM targets WHERE id = $1`, target.ID).Scan(&existingID)
		if err == nil {
			if conflict == "skip" {
				targetIDMap[target.ID] = existingID
				result.TargetsSkipped++
				continue
			}
			_, err = tx.Exec(ctx, `
				UPDATE targets
				SET host = $2, normalized_host = $3, tags = $4, created_at = $5
				WHERE id = $1
			`, target.ID, target.Host, normalized, tags, target.CreatedAt)
			if err != nil {
				return nil, fmt.Errorf("update target %s: %w", target.ID, err)
			}
			targetIDMap[target.ID] = target.ID
			result.TargetsImported++
			continue
		}
		if err != pgx.ErrNoRows {
			return nil, fmt.Errorf("lookup target %s: %w", target.ID, err)
		}

		var hostOwner uuid.UUID
		err = tx.QueryRow(ctx, `SELECT id FROM targets WHERE host = $1`, target.Host).Scan(&hostOwner)
		if err == nil {
			if conflict == "skip" {
				targetIDMap[target.ID] = hostOwner
				result.TargetsSkipped++
				result.Warnings = append(result.Warnings,
					fmt.Sprintf("target %s skipped: host %q already exists as %s", target.ID, target.Host, hostOwner))
				continue
			}
			_, err = tx.Exec(ctx, `
				UPDATE targets
				SET normalized_host = $2, tags = $3
				WHERE id = $1
			`, hostOwner, normalized, tags)
			if err != nil {
				return nil, fmt.Errorf("update existing host target %s: %w", hostOwner, err)
			}
			targetIDMap[target.ID] = hostOwner
			result.TargetsImported++
			continue
		}
		if err != pgx.ErrNoRows {
			return nil, fmt.Errorf("lookup host %q: %w", target.Host, err)
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO targets (id, host, normalized_host, tags, created_at)
			VALUES ($1, $2, $3, $4, $5)
		`, target.ID, target.Host, normalized, tags, target.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("insert target %s: %w", target.ID, err)
		}
		targetIDMap[target.ID] = target.ID
		result.TargetsImported++
	}

	for _, snapshot := range bundle.Snapshots {
		targetID, ok := targetIDMap[snapshot.TargetID]
		if !ok {
			result.SnapshotsSkipped++
			continue
		}

		var existingID uuid.UUID
		lookupErr := tx.QueryRow(ctx, `SELECT id FROM snapshots WHERE id = $1`, snapshot.ID).Scan(&existingID)
		if lookupErr == nil && conflict == "skip" {
			result.SnapshotsSkipped++
			importedSnapshotIDs = append(importedSnapshotIDs, existingID)
			continue
		}
		if lookupErr != nil && lookupErr != pgx.ErrNoRows {
			return nil, fmt.Errorf("lookup snapshot %s: %w", snapshot.ID, lookupErr)
		}

		rawData := snapshot.RawData
		if rawData == nil {
			rawData = map[string]any{}
		}
		changes := snapshot.Changes
		if changes == nil {
			changes = []string{}
		}
		changeDetails := snapshot.ChangeDetails
		if changeDetails == nil {
			changeDetails = []models.ChangeDetail{}
		}
		changeDetailsJSON, err := json.Marshal(changeDetails)
		if err != nil {
			return nil, fmt.Errorf("marshal change details: %w", err)
		}

		if lookupErr == nil {
			_, err = tx.Exec(ctx, `
				UPDATE snapshots
				SET target_id = $2, scanned_at = $3, last_seen = $4, data_hash = $5, raw_data = $6,
					changes = $7, change_details = $8, client_ip = $9,
					pwhois_data = $10, pwhois_looked_up_at = $11, pwhois_origin_as = $12,
					pwhois_org_name = $13, pwhois_country_code = $14, pwhois_city = $15, pwhois_prefix = $16
				WHERE id = $1
			`,
				snapshot.ID, targetID, snapshot.ScannedAt, snapshot.LastSeen, snapshot.DataHash, rawData,
				changes, changeDetailsJSON, snapshot.ClientIP,
				snapshot.PwhoisData, snapshot.PwhoisLookedUp, snapshot.PwhoisOriginAS,
				snapshot.PwhoisOrgName, snapshot.PwhoisCountry, snapshot.PwhoisCity, snapshot.PwhoisPrefix,
			)
		} else {
			_, err = tx.Exec(ctx, `
				INSERT INTO snapshots (
					id, target_id, scanned_at, last_seen, data_hash, raw_data, changes, change_details,
					client_ip, pwhois_data, pwhois_looked_up_at, pwhois_origin_as, pwhois_org_name,
					pwhois_country_code, pwhois_city, pwhois_prefix
				)
				VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
			`,
				snapshot.ID, targetID, snapshot.ScannedAt, snapshot.LastSeen, snapshot.DataHash, rawData,
				changes, changeDetailsJSON, snapshot.ClientIP,
				snapshot.PwhoisData, snapshot.PwhoisLookedUp, snapshot.PwhoisOriginAS,
				snapshot.PwhoisOrgName, snapshot.PwhoisCountry, snapshot.PwhoisCity, snapshot.PwhoisPrefix,
			)
		}
		if err != nil {
			return nil, fmt.Errorf("upsert snapshot %s: %w", snapshot.ID, err)
		}
		importedSnapshotIDs = append(importedSnapshotIDs, snapshot.ID)
		result.SnapshotsImported++
	}

	for _, blob := range bundle.SnapshotBlobs {
		data, err := base64.StdEncoding.DecodeString(blob.DataBase64)
		if err != nil {
			return nil, fmt.Errorf("decode blob %s: %w", blob.ID, err)
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO snapshot_blobs (id, snapshot_id, kind, content_type, data)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (snapshot_id, kind) DO UPDATE
			SET content_type = EXCLUDED.content_type, data = EXCLUDED.data
		`, blob.ID, blob.SnapshotID, blob.Kind, blob.ContentType, data)
		if err != nil {
			return nil, fmt.Errorf("upsert snapshot blob %s: %w", blob.ID, err)
		}
		result.SnapshotBlobsImported++
	}

	collectionIDMap := map[uuid.UUID]uuid.UUID{}
	for _, collection := range bundle.Collections {
		var existingID uuid.UUID
		err := tx.QueryRow(ctx, `SELECT id FROM target_collections WHERE id = $1`, collection.ID).Scan(&existingID)
		if err == nil {
			if conflict == "skip" {
				collectionIDMap[collection.ID] = existingID
				result.CollectionsSkipped++
				continue
			}
			_, err = tx.Exec(ctx, `
				UPDATE target_collections
				SET name = $2, description = $3, created_at = $4, updated_at = $5
				WHERE id = $1
			`, collection.ID, collection.Name, collection.Description, collection.CreatedAt, collection.UpdatedAt)
			if err != nil {
				return nil, fmt.Errorf("update collection %s: %w", collection.ID, err)
			}
			collectionIDMap[collection.ID] = collection.ID
			result.CollectionsImported++
			continue
		}
		if err != pgx.ErrNoRows {
			return nil, fmt.Errorf("lookup collection %s: %w", collection.ID, err)
		}

		var nameOwner uuid.UUID
		err = tx.QueryRow(ctx, `SELECT id FROM target_collections WHERE name = $1`, collection.Name).Scan(&nameOwner)
		if err == nil {
			if conflict == "skip" {
				collectionIDMap[collection.ID] = nameOwner
				result.CollectionsSkipped++
				continue
			}
			_, err = tx.Exec(ctx, `
				UPDATE target_collections
				SET description = $2, updated_at = $3
				WHERE id = $1
			`, nameOwner, collection.Description, collection.UpdatedAt)
			if err != nil {
				return nil, fmt.Errorf("update collection by name %s: %w", nameOwner, err)
			}
			collectionIDMap[collection.ID] = nameOwner
			result.CollectionsImported++
			continue
		}
		if err != pgx.ErrNoRows {
			return nil, fmt.Errorf("lookup collection name %q: %w", collection.Name, err)
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO target_collections (id, name, description, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5)
		`, collection.ID, collection.Name, collection.Description, collection.CreatedAt, collection.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("insert collection %s: %w", collection.ID, err)
		}
		collectionIDMap[collection.ID] = collection.ID
		result.CollectionsImported++
	}

	for _, member := range bundle.CollectionMembers {
		collectionID, ok := collectionIDMap[member.CollectionID]
		if !ok {
			continue
		}
		targetID, ok := targetIDMap[member.TargetID]
		if !ok {
			continue
		}
		_, err := tx.Exec(ctx, `
			INSERT INTO target_collection_members (collection_id, target_id, added_at)
			VALUES ($1, $2, $3)
			ON CONFLICT (collection_id, target_id) DO UPDATE SET added_at = EXCLUDED.added_at
		`, collectionID, targetID, member.AddedAt)
		if err != nil {
			return nil, fmt.Errorf("upsert collection member: %w", err)
		}
		result.CollectionMembersImported++
	}

	for _, note := range bundle.Notes {
		var existingID uuid.UUID
		lookupErr := tx.QueryRow(ctx, `SELECT id FROM investigation_notes WHERE id = $1`, note.ID).Scan(&existingID)
		if lookupErr == nil && conflict == "skip" {
			result.NotesSkipped++
			continue
		}
		if lookupErr != nil && lookupErr != pgx.ErrNoRows {
			return nil, fmt.Errorf("lookup note %s: %w", note.ID, lookupErr)
		}

		targetID := remapOptionalUUID(note.TargetID, targetIDMap)
		snapshotID := note.SnapshotID
		collectionID := remapOptionalUUID(note.CollectionID, collectionIDMap)
		refs := note.References
		if refs == nil {
			refs = []models.NoteReference{}
		}
		refsJSON, err := json.Marshal(refs)
		if err != nil {
			return nil, fmt.Errorf("marshal note refs: %w", err)
		}

		if lookupErr == nil {
			_, err = tx.Exec(ctx, `
				UPDATE investigation_notes
				SET title = $2, body = $3, status = $4, target_id = $5, snapshot_id = $6,
					collection_id = $7, graph_node_id = $8, graph_node_label = $9, graph_node_type = $10,
					refs = $11, trashed_at = $12, archived_at = $13, created_at = $14, updated_at = $15
				WHERE id = $1
			`,
				note.ID, note.Title, note.Body, note.Status, targetID, snapshotID, collectionID,
				note.GraphNodeID, note.GraphNodeLabel, note.GraphNodeType, refsJSON,
				note.TrashedAt, note.ArchivedAt, note.CreatedAt, note.UpdatedAt,
			)
		} else {
			_, err = tx.Exec(ctx, `
				INSERT INTO investigation_notes (
					id, title, body, status, target_id, snapshot_id, collection_id,
					graph_node_id, graph_node_label, graph_node_type, refs,
					trashed_at, archived_at, created_at, updated_at
				)
				VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
			`,
				note.ID, note.Title, note.Body, note.Status, targetID, snapshotID, collectionID,
				note.GraphNodeID, note.GraphNodeLabel, note.GraphNodeType, refsJSON,
				note.TrashedAt, note.ArchivedAt, note.CreatedAt, note.UpdatedAt,
			)
		}
		if err != nil {
			return nil, fmt.Errorf("upsert note %s: %w", note.ID, err)
		}
		result.NotesImported++
	}

	for _, view := range bundle.GraphViews {
		var existingID uuid.UUID
		lookupErr := tx.QueryRow(ctx, `SELECT id FROM graph_views WHERE id = $1`, view.ID).Scan(&existingID)
		if lookupErr == nil && conflict == "skip" {
			result.GraphViewsSkipped++
			continue
		}
		if lookupErr != nil && lookupErr != pgx.ErrNoRows {
			return nil, fmt.Errorf("lookup graph view %s: %w", view.ID, lookupErr)
		}

		targetID := remapOptionalUUID(view.TargetID, targetIDMap)
		pinned := view.PinnedNodes
		if pinned == nil {
			pinned = map[string]models.GraphPinnedNode{}
		}
		pinnedJSON, err := json.Marshal(pinned)
		if err != nil {
			return nil, fmt.Errorf("marshal pinned nodes: %w", err)
		}

		if lookupErr == nil {
			_, err = tx.Exec(ctx, `
				UPDATE graph_views
				SET name = $2, description = $3, view_mode = $4, target_id = $5, vantage_filter = $6,
					snapshot_id = $7, compare_snapshot_id = $8, compare_mode = $9,
					pinned_nodes = $10, selected_node_id = $11, created_at = $12, updated_at = $13
				WHERE id = $1
			`,
				view.ID, view.Name, view.Description, view.ViewMode, targetID, view.VantageFilter,
				view.SnapshotID, view.CompareSnapshotID, view.CompareMode, pinnedJSON,
				view.SelectedNodeID, view.CreatedAt, view.UpdatedAt,
			)
		} else {
			_, err = tx.Exec(ctx, `
				INSERT INTO graph_views (
					id, name, description, view_mode, target_id, vantage_filter,
					snapshot_id, compare_snapshot_id, compare_mode, pinned_nodes,
					selected_node_id, created_at, updated_at
				)
				VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
			`,
				view.ID, view.Name, view.Description, view.ViewMode, targetID, view.VantageFilter,
				view.SnapshotID, view.CompareSnapshotID, view.CompareMode, pinnedJSON,
				view.SelectedNodeID, view.CreatedAt, view.UpdatedAt,
			)
		}
		if err != nil {
			return nil, fmt.Errorf("upsert graph view %s: %w", view.ID, err)
		}
		result.GraphViewsImported++
	}

	for _, webhook := range bundle.Webhooks {
		config := webhook.Config
		if config == nil {
			config = map[string]any{}
		}
		configJSON, err := json.Marshal(config)
		if err != nil {
			return nil, fmt.Errorf("marshal webhook config: %w", err)
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO webhooks (id, name, type, url, config, enabled, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (id) DO UPDATE
			SET name = EXCLUDED.name, type = EXCLUDED.type, url = EXCLUDED.url,
				config = EXCLUDED.config, enabled = EXCLUDED.enabled, updated_at = EXCLUDED.updated_at
		`, webhook.ID, webhook.Name, webhook.Type, webhook.URL, configJSON, webhook.Enabled, webhook.CreatedAt, webhook.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("upsert webhook %s: %w", webhook.ID, err)
		}
		result.WebhooksImported++
	}

	if bundle.Settings != nil {
		settingsJSON, err := json.Marshal(bundle.Settings)
		if err != nil {
			return nil, fmt.Errorf("marshal settings: %w", err)
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO settings (key, value, updated_at)
			VALUES ('app_settings', $1, NOW())
			ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()
		`, settingsJSON)
		if err != nil {
			return nil, fmt.Errorf("upsert settings: %w", err)
		}
		result.SettingsImported = true
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit import: %w", err)
	}

	if opts.RebuildGraph && neo4j != nil && len(importedSnapshotIDs) > 0 {
		for _, snapshotID := range importedSnapshotIDs {
			snapshot, target, err := loadSnapshotForSync(ctx, database, snapshotID)
			if err != nil || snapshot == nil || target == nil {
				continue
			}
			go func(target models.Target, snapshot models.Snapshot) {
				_ = neo4j.SyncSnapshot(context.Background(), &target, &snapshot)
			}(*target, *snapshot)
			result.GraphSyncQueued++
		}
	}

	return result, nil
}

func remapOptionalUUID(id *uuid.UUID, mapping map[uuid.UUID]uuid.UUID) *uuid.UUID {
	if id == nil {
		return nil
	}
	mapped, ok := mapping[*id]
	if !ok {
		return id
	}
	return &mapped
}

func loadSnapshotForSync(ctx context.Context, database *db.DB, snapshotID uuid.UUID) (*models.Snapshot, *models.Target, error) {
	var snapshot models.Snapshot
	var changeDetailsJSON []byte
	err := database.Pool.QueryRow(ctx, `
		SELECT s.id, s.target_id, s.scanned_at, s.last_seen, s.data_hash, s.raw_data, s.changes,
			COALESCE(s.change_details, '[]'), COALESCE(s.client_ip, ''),
			s.pwhois_data, s.pwhois_looked_up_at, s.pwhois_origin_as, s.pwhois_org_name,
			s.pwhois_country_code, s.pwhois_city, s.pwhois_prefix
		FROM snapshots s
		WHERE s.id = $1
	`, snapshotID).Scan(
		&snapshot.ID, &snapshot.TargetID, &snapshot.ScannedAt, &snapshot.LastSeen,
		&snapshot.DataHash, &snapshot.RawData, &snapshot.Changes, &changeDetailsJSON, &snapshot.ClientIP,
		&snapshot.PwhoisData, &snapshot.PwhoisLookedUp, &snapshot.PwhoisOriginAS,
		&snapshot.PwhoisOrgName, &snapshot.PwhoisCountry, &snapshot.PwhoisCity, &snapshot.PwhoisPrefix,
	)
	if err != nil {
		return nil, nil, err
	}
	if len(changeDetailsJSON) > 0 {
		_ = json.Unmarshal(changeDetailsJSON, &snapshot.ChangeDetails)
	}

	var target models.Target
	err = database.Pool.QueryRow(ctx, `
		SELECT id, host, tags, created_at
		FROM targets
		WHERE id = $1
	`, snapshot.TargetID).Scan(&target.ID, &target.Host, &target.Tags, &target.CreatedAt)
	if err != nil {
		return &snapshot, nil, err
	}
	return &snapshot, &target, nil
}