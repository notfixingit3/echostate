package models

import (
	"time"

	"github.com/google/uuid"
)

// Target represents a host or IP address under surveillance.
type Target struct {
	ID        uuid.UUID `json:"id"`
	Host      string    `json:"host"`
	CreatedAt time.Time `json:"created_at"`
}

// Snapshot stores a single reconnaissance result for a target.
type Snapshot struct {
	ID        uuid.UUID       `json:"id"`
	TargetID  uuid.UUID       `json:"target_id"`
	ScannedAt time.Time       `json:"scanned_at"`
	LastSeen  time.Time       `json:"last_seen"`
	DataHash  string          `json:"data_hash"`
	RawData   map[string]any  `json:"raw_data"`
	Changes   []string        `json:"changes,omitempty"`
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
