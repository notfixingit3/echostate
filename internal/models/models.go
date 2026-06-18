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
	PwhoisData      map[string]any `json:"pwhois_data,omitempty"`
	PwhoisLookedUp  *time.Time     `json:"pwhois_looked_up_at,omitempty"`
	PwhoisOriginAS  *string        `json:"pwhois_origin_as,omitempty"`
	PwhoisOrgName   *string        `json:"pwhois_org_name,omitempty"`
	PwhoisCountry   *string        `json:"pwhois_country_code,omitempty"`
	PwhoisCity      *string        `json:"pwhois_city,omitempty"`
	PwhoisPrefix    *string        `json:"pwhois_prefix,omitempty"`
	ClientIP        string         `json:"client_ip,omitempty"`
}

// ScanRequest is the payload accepted by POST /api/scan.
type ScanRequest struct {
	Host string `json:"host" binding:"required"`
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
	Errors    []string       `json:"errors,omitempty"`
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
