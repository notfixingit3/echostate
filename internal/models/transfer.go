package models

import (
	"time"

	"github.com/google/uuid"
)

const ExportBundleFormatVersion = 1

// ExportTarget is a portable target row for backup and migration.
type ExportTarget struct {
	ID             uuid.UUID `json:"id"`
	Host           string    `json:"host"`
	NormalizedHost string    `json:"normalized_host,omitempty"`
	Tags           []string  `json:"tags"`
	CreatedAt      time.Time `json:"created_at"`
}

// ExportSnapshot is a portable snapshot row for backup and migration.
type ExportSnapshot struct {
	ID             uuid.UUID      `json:"id"`
	TargetID       uuid.UUID      `json:"target_id"`
	ScannedAt      time.Time      `json:"scanned_at"`
	LastSeen       time.Time      `json:"last_seen"`
	DataHash       string         `json:"data_hash"`
	RawData        map[string]any `json:"raw_data"`
	Changes        []string       `json:"changes,omitempty"`
	ChangeDetails  []ChangeDetail `json:"change_details,omitempty"`
	PwhoisData     map[string]any `json:"pwhois_data,omitempty"`
	PwhoisLookedUp *time.Time     `json:"pwhois_looked_up_at,omitempty"`
	PwhoisOriginAS *string        `json:"pwhois_origin_as,omitempty"`
	PwhoisOrgName  *string        `json:"pwhois_org_name,omitempty"`
	PwhoisCountry  *string        `json:"pwhois_country_code,omitempty"`
	PwhoisCity     *string        `json:"pwhois_city,omitempty"`
	PwhoisPrefix   *string        `json:"pwhois_prefix,omitempty"`
	ClientIP       string         `json:"client_ip,omitempty"`
}

// ExportSnapshotBlob carries binary snapshot assets in a JSON-safe envelope.
type ExportSnapshotBlob struct {
	ID          uuid.UUID `json:"id"`
	SnapshotID  uuid.UUID `json:"snapshot_id"`
	Kind        string    `json:"kind"`
	ContentType string    `json:"content_type"`
	DataBase64  string    `json:"data_base64"`
}

// ExportCollection is a portable collection row.
type ExportCollection struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ExportCollectionMember links a target to a collection.
type ExportCollectionMember struct {
	CollectionID uuid.UUID `json:"collection_id"`
	TargetID     uuid.UUID `json:"target_id"`
	AddedAt      time.Time `json:"added_at"`
}

// ExportBundle is the portable EchoState investigation dataset.
type ExportBundle struct {
	FormatVersion    int                      `json:"format_version"`
	ExportedAt       time.Time                `json:"exported_at"`
	EchoStateVersion string                   `json:"echostate_version"`
	Targets          []ExportTarget           `json:"targets"`
	Snapshots        []ExportSnapshot         `json:"snapshots"`
	SnapshotBlobs    []ExportSnapshotBlob     `json:"snapshot_blobs,omitempty"`
	Collections      []ExportCollection       `json:"collections"`
	CollectionMembers []ExportCollectionMember `json:"collection_members"`
	Notes            []InvestigationNote      `json:"notes"`
	GraphViews       []SavedGraphView         `json:"graph_views"`
	Webhooks         []Webhook                `json:"webhooks,omitempty"`
	Settings         map[string]any           `json:"settings,omitempty"`
}

// ImportRequest imports a portable bundle into the current install.
type ImportRequest struct {
	Conflict     string       `json:"conflict"`
	RebuildGraph bool         `json:"rebuild_graph"`
	Bundle       ExportBundle `json:"bundle"`
}

// ImportResult summarizes an import operation.
type ImportResult struct {
	TargetsImported          int      `json:"targets_imported"`
	TargetsSkipped           int      `json:"targets_skipped"`
	SnapshotsImported        int      `json:"snapshots_imported"`
	SnapshotsSkipped         int      `json:"snapshots_skipped"`
	SnapshotBlobsImported    int      `json:"snapshot_blobs_imported"`
	CollectionsImported      int      `json:"collections_imported"`
	CollectionsSkipped       int      `json:"collections_skipped"`
	CollectionMembersImported int     `json:"collection_members_imported"`
	NotesImported            int      `json:"notes_imported"`
	NotesSkipped             int      `json:"notes_skipped"`
	GraphViewsImported       int      `json:"graph_views_imported"`
	GraphViewsSkipped        int      `json:"graph_views_skipped"`
	WebhooksImported         int      `json:"webhooks_imported"`
	SettingsImported         bool     `json:"settings_imported"`
	GraphSyncQueued          int      `json:"graph_sync_queued"`
	Warnings                 []string `json:"warnings,omitempty"`
}