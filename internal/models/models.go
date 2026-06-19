package models

import (
	"time"

	"github.com/google/uuid"
)

type Target struct {
	ID        uuid.UUID `json:"id"`
	Host      string    `json:"host"`
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
}

// Snapshot stores a single reconnaissance result for a target.
type Snapshot struct {
	ID              uuid.UUID      `json:"id"`
	TargetID        uuid.UUID      `json:"target_id"`
	ScannedAt       time.Time      `json:"scanned_at"`
	LastSeen        time.Time      `json:"last_seen"`
	DataHash        string         `json:"data_hash"`
	RawData         map[string]any `json:"raw_data"`
	Changes         []string       `json:"changes,omitempty"`
	ChangeDetails   []ChangeDetail `json:"change_details,omitempty"`
	PwhoisData      map[string]any `json:"pwhois_data,omitempty"`
	PwhoisLookedUp  *time.Time     `json:"pwhois_looked_up_at,omitempty"`
	PwhoisOriginAS  *string        `json:"pwhois_origin_as,omitempty"`
	PwhoisOrgName   *string        `json:"pwhois_org_name,omitempty"`
	PwhoisCountry   *string        `json:"pwhois_country_code,omitempty"`
	PwhoisCity      *string        `json:"pwhois_city,omitempty"`
	PwhoisPrefix    *string        `json:"pwhois_prefix,omitempty"`
	ClientIP        string         `json:"client_ip,omitempty"`
}

// ChangeDetail is a structured snapshot diff entry.
type ChangeDetail struct {
	Type     string `json:"type"`
	Severity string `json:"severity"`
	Summary  string `json:"summary"`
	Field    string `json:"field,omitempty"`
	Detail   string `json:"detail,omitempty"`
}

// ScanRequest is the payload accepted by POST /api/scan.
type ScanRequest struct {
	Host string `json:"host" binding:"required"`
}

// ScanJobStatus represents the lifecycle state of an async scan job.
type ScanJobStatus string

const (
	ScanJobPending   ScanJobStatus = "pending"
	ScanJobRunning   ScanJobStatus = "running"
	ScanJobCompleted ScanJobStatus = "completed"
	ScanJobFailed    ScanJobStatus = "failed"
)

func (s ScanJobStatus) IsValid() bool {
	switch s {
	case ScanJobPending, ScanJobRunning, ScanJobCompleted, ScanJobFailed:
		return true
	}
	return false
}

// ScanJob tracks an asynchronous reconnaissance scan.
type ScanJob struct {
	ID           uuid.UUID      `json:"id"`
	Host         string         `json:"host"`
	Status       ScanJobStatus  `json:"status"`
	SnapshotID   *uuid.UUID     `json:"snapshot_id,omitempty"`
	TargetID     *uuid.UUID     `json:"target_id,omitempty"`
	ErrorMessage string         `json:"error,omitempty"`
	ClientIP     string         `json:"client_ip,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	CompletedAt  *time.Time     `json:"completed_at,omitempty"`
	Snapshot     *Snapshot      `json:"snapshot,omitempty"`
}

// ScanResult is the unified reconnaissance payload stored in raw_data.
type ScanResult struct {
	Host      string         `json:"host"`
	ScannedAt time.Time      `json:"scanned_at"`
	WHOIS     map[string]any `json:"whois,omitempty"`
	ASN       map[string]any `json:"asn,omitempty"`
	Web       map[string]any `json:"web,omitempty"`
	TLS       map[string]any `json:"tls,omitempty"`
	DNS       map[string]any `json:"dns,omitempty"`
	Favicon   map[string]any `json:"favicon,omitempty"`
	Crawl     map[string]any `json:"crawl,omitempty"`
	Storage   map[string]any `json:"storage,omitempty"`
	CT          map[string]any `json:"ct,omitempty"`
	Traceroute  map[string]any `json:"traceroute,omitempty"`
	Screenshot  map[string]any `json:"screenshot,omitempty"`
	Enrichment  map[string]any `json:"enrichment,omitempty"`
	Errors      []string       `json:"errors,omitempty"`
}

// ReportStatus represents the lifecycle state of a PDF report.
type ReportStatus string

const (
	ReportPending   ReportStatus = "pending"
	ReportRunning   ReportStatus = "running"
	ReportCompleted ReportStatus = "completed"
	ReportFailed    ReportStatus = "failed"
)

func (s ReportStatus) IsValid() bool {
	switch s {
	case ReportPending, ReportRunning, ReportCompleted, ReportFailed:
		return true
	}
	return false
}

// CreateReportRequest is the payload accepted by POST /api/reports.
type CreateReportRequest struct {
	SnapshotID *uuid.UUID `json:"snapshot_id,omitempty"`
	Host       *string    `json:"host,omitempty"`
}

// ReportResponse is the JSON envelope returned by report endpoints.
type ReportResponse struct {
	ID          uuid.UUID    `json:"id"`
	Status      ReportStatus `json:"status"`
	SnapshotID  uuid.UUID    `json:"snapshot_id"`
	Host        string       `json:"host"`
	DownloadURL string       `json:"download_url,omitempty"`
	Error       string       `json:"error,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	CompletedAt *time.Time   `json:"completed_at,omitempty"`
}

// Report is the internal model mapping to the reports database table.
type Report struct {
	ID           uuid.UUID    `json:"id"`
	SnapshotID   uuid.UUID    `json:"snapshot_id"`
	Status       ReportStatus `json:"status"`
	ErrorMessage string       `json:"error_message,omitempty"`
	PDF          []byte       `json:"-"`
	CreatedAt    time.Time    `json:"created_at"`
	CompletedAt  *time.Time   `json:"completed_at,omitempty"`
}

type PaginatedResponse[T any] struct {
	Data  []T `json:"data"`
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

type TargetSummary struct {
	ID               uuid.UUID  `json:"id"`
	Host             string     `json:"host"`
	Tags             []string   `json:"tags"`
	CreatedAt        time.Time  `json:"created_at"`
	SnapshotCount    int        `json:"snapshot_count"`
	LatestSnapshotAt *time.Time `json:"latest_snapshot_at,omitempty"`
	LatestAsn        *string    `json:"latest_asn,omitempty"`
	LatestAsName     *string    `json:"latest_as_name,omitempty"`
	LatestWebTitle   *string    `json:"latest_web_title,omitempty"`
}

// TargetDetail is the response envelope for GET /api/targets/:id.
type TargetDetail struct {
	TargetSummary
	LatestSnapshot *Snapshot `json:"latest_snapshot,omitempty"`
}

// ScreenshotEntry is a thumbnail captured during a snapshot scan.
type ScreenshotEntry struct {
	SnapshotID  uuid.UUID `json:"snapshot_id"`
	ScannedAt   time.Time `json:"scanned_at"`
	URL         string    `json:"url,omitempty"`
	Width       int       `json:"width,omitempty"`
	Height      int       `json:"height,omitempty"`
	Format      string    `json:"format,omitempty"`
	Thumbnail   string    `json:"thumbnail,omitempty"`
	Error       string    `json:"error,omitempty"`
}

type SnapshotSummary struct {
	ID              uuid.UUID  `json:"id"`
	TargetID        uuid.UUID  `json:"target_id"`
	Host            string     `json:"host"`
	ScannedAt       time.Time  `json:"scanned_at"`
	LastSeen        time.Time  `json:"last_seen"`
	DataHash        string     `json:"data_hash"`
	Changes         []string   `json:"changes,omitempty"`
	ClientIP        string     `json:"client_ip"`
	PwhoisLookedUp  *time.Time `json:"pwhois_looked_up_at,omitempty"`
	PwhoisOriginAS  *string    `json:"pwhois_origin_as,omitempty"`
	PwhoisOrgName   *string    `json:"pwhois_org_name,omitempty"`
	PwhoisCountry   *string    `json:"pwhois_country_code,omitempty"`
	PwhoisCity      *string    `json:"pwhois_city,omitempty"`
	PwhoisPrefix    *string    `json:"pwhois_prefix,omitempty"`
	Asn             *string    `json:"asn,omitempty"`
	AsName          *string    `json:"as_name,omitempty"`
	WebTitle        *string    `json:"web_title,omitempty"`
	Registrar       *string    `json:"registrar,omitempty"`
	ResolvedIP      *string    `json:"resolved_ip,omitempty"`
}

type ReportSummary struct {
	ID          uuid.UUID    `json:"id"`
	SnapshotID  uuid.UUID    `json:"snapshot_id"`
	Host        string       `json:"host"`
	Status      ReportStatus `json:"status"`
	CreatedAt   time.Time    `json:"created_at"`
	CompletedAt *time.Time   `json:"completed_at,omitempty"`
}

type TargetCollectionSummary struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	TargetCount int       `json:"target_count"`
}

type TargetCollectionDetail struct {
	TargetCollectionSummary
	Targets []TargetSummary `json:"targets"`
}

type CreateTargetCollectionRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type UpdateTargetCollectionRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type CollectionTargetRequest struct {
	TargetID uuid.UUID `json:"target_id" binding:"required"`
}

type CollectionRescanJob struct {
	TargetID uuid.UUID `json:"target_id"`
	Host     string    `json:"host"`
	JobID    uuid.UUID `json:"job_id"`
}

type CollectionRescanResponse struct {
	CollectionID uuid.UUID             `json:"collection_id"`
	Jobs         []CollectionRescanJob `json:"jobs"`
}

type NoteStatus string

const (
	NoteStatusActive   NoteStatus = "active"
	NoteStatusArchived NoteStatus = "archived"
	NoteStatusTrashed  NoteStatus = "trashed"
)

func (s NoteStatus) IsValid() bool {
	switch s {
	case NoteStatusActive, NoteStatusArchived, NoteStatusTrashed:
		return true
	}
	return false
}

type NoteReference struct {
	Type  string `json:"type"`
	Label string `json:"label,omitempty"`
	URL   string `json:"url,omitempty"`
	ID    string `json:"id,omitempty"`
}

type InvestigationNote struct {
	ID             uuid.UUID       `json:"id"`
	Title          string          `json:"title"`
	Body           string          `json:"body"`
	Status         NoteStatus      `json:"status"`
	TargetID       *uuid.UUID      `json:"target_id,omitempty"`
	SnapshotID     *uuid.UUID      `json:"snapshot_id,omitempty"`
	CollectionID   *uuid.UUID      `json:"collection_id,omitempty"`
	GraphNodeID    *string         `json:"graph_node_id,omitempty"`
	GraphNodeLabel *string         `json:"graph_node_label,omitempty"`
	GraphNodeType  *string         `json:"graph_node_type,omitempty"`
	References     []NoteReference `json:"references"`
	TrashedAt      *time.Time      `json:"trashed_at,omitempty"`
	ArchivedAt     *time.Time      `json:"archived_at,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	TargetHost     *string         `json:"target_host,omitempty"`
	CollectionName *string         `json:"collection_name,omitempty"`
}

type CreateInvestigationNoteRequest struct {
	Title          string          `json:"title"`
	Body           string          `json:"body" binding:"required"`
	TargetID       *uuid.UUID      `json:"target_id"`
	SnapshotID     *uuid.UUID      `json:"snapshot_id"`
	CollectionID   *uuid.UUID      `json:"collection_id"`
	GraphNodeID    *string         `json:"graph_node_id"`
	GraphNodeLabel *string         `json:"graph_node_label"`
	GraphNodeType  *string         `json:"graph_node_type"`
	References     []NoteReference `json:"references"`
}

type UpdateInvestigationNoteRequest struct {
	Title          string          `json:"title"`
	Body           string          `json:"body" binding:"required"`
	TargetID       *uuid.UUID      `json:"target_id"`
	SnapshotID     *uuid.UUID      `json:"snapshot_id"`
	CollectionID   *uuid.UUID      `json:"collection_id"`
	GraphNodeID    *string         `json:"graph_node_id"`
	GraphNodeLabel *string         `json:"graph_node_label"`
	GraphNodeType  *string         `json:"graph_node_type"`
	References     []NoteReference `json:"references"`
}

type GraphPinnedNode struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type SavedGraphView struct {
	ID                  uuid.UUID                  `json:"id"`
	Name                string                     `json:"name"`
	Description         string                     `json:"description"`
	ViewMode            string                     `json:"view_mode"`
	TargetID            *uuid.UUID                 `json:"target_id,omitempty"`
	VantageFilter       string                     `json:"vantage_filter"`
	SnapshotID          *uuid.UUID                 `json:"snapshot_id,omitempty"`
	CompareSnapshotID   *uuid.UUID                 `json:"compare_snapshot_id,omitempty"`
	CompareMode         string                     `json:"compare_mode"`
	PinnedNodes         map[string]GraphPinnedNode `json:"pinned_nodes"`
	SelectedNodeID      *string                    `json:"selected_node_id,omitempty"`
	CreatedAt           time.Time                  `json:"created_at"`
	UpdatedAt           time.Time                  `json:"updated_at"`
	TargetHost          *string                    `json:"target_host,omitempty"`
}

type CreateSavedGraphViewRequest struct {
	Name              string                     `json:"name" binding:"required"`
	Description       string                     `json:"description"`
	ViewMode          string                     `json:"view_mode"`
	TargetID          *uuid.UUID                 `json:"target_id"`
	VantageFilter     string                     `json:"vantage_filter"`
	SnapshotID        *uuid.UUID                 `json:"snapshot_id"`
	CompareSnapshotID *uuid.UUID                 `json:"compare_snapshot_id"`
	CompareMode       string                     `json:"compare_mode"`
	PinnedNodes       map[string]GraphPinnedNode `json:"pinned_nodes"`
	SelectedNodeID    *string                    `json:"selected_node_id"`
}

type UpdateSavedGraphViewRequest struct {
	Name              string                     `json:"name" binding:"required"`
	Description       string                     `json:"description"`
	ViewMode          string                     `json:"view_mode"`
	TargetID          *uuid.UUID                 `json:"target_id"`
	VantageFilter     string                     `json:"vantage_filter"`
	SnapshotID        *uuid.UUID                 `json:"snapshot_id"`
	CompareSnapshotID *uuid.UUID                 `json:"compare_snapshot_id"`
	CompareMode       string                     `json:"compare_mode"`
	PinnedNodes       map[string]GraphPinnedNode `json:"pinned_nodes"`
	SelectedNodeID    *string                    `json:"selected_node_id"`
}

type Webhook struct {
	ID        uuid.UUID      `json:"id"`
	Name      string         `json:"name"`
	Type      string         `json:"type"`
	URL       string         `json:"url"`
	Config    map[string]any `json:"config,omitempty"`
	Enabled   bool           `json:"enabled"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}
