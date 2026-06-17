package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

// --- ReportStatus constants ---

func TestReportStatus_Constants(t *testing.T) {
	tests := []struct {
		name string
		got  ReportStatus
		want string
	}{
		{"ReportPending", ReportPending, "pending"},
		{"ReportRunning", ReportRunning, "running"},
		{"ReportCompleted", ReportCompleted, "completed"},
		{"ReportFailed", ReportFailed, "failed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.got) != tt.want {
				t.Errorf("ReportStatus(%q) = %q, want %q", tt.name, string(tt.got), tt.want)
			}
		})
	}
}

func TestReportStatus_IsStringType(t *testing.T) {
	// Verify ReportStatus is a distinct string type and constants are comparable.
	var s ReportStatus = "pending"
	if s != ReportPending {
		t.Errorf(`ReportStatus("pending") != ReportPending`)
	}
}

// --- ReportResponse JSON round-trip ---

func TestReportResponse_JSONRoundTrip(t *testing.T) {
	id := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)
	completedAt := now.Add(5 * time.Minute)

	tests := []struct {
		name string
		in   ReportResponse
	}{
		{
			name: "full response with download URL",
			in: ReportResponse{
				ID:          id,
				Status:      ReportCompleted,
				SnapshotID:  uuid.New(),
				DownloadURL: "https://example.com/report.pdf",
				Error:       "",
				CreatedAt:   now,
				CompletedAt: &completedAt,
			},
		},
		{
			name: "pending response without download URL",
			in: ReportResponse{
				ID:          uuid.New(),
				Status:      ReportPending,
				SnapshotID:  uuid.New(),
				DownloadURL: "",
				Error:       "",
				CreatedAt:   now,
				CompletedAt: nil,
			},
		},
		{
			name: "failed response with error",
			in: ReportResponse{
				ID:          uuid.New(),
				Status:      ReportFailed,
				SnapshotID:  uuid.New(),
				DownloadURL: "",
				Error:       "scan failed: timeout",
				CreatedAt:   now,
				CompletedAt: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.in)
			if err != nil {
				t.Fatalf("Marshal error: %v", err)
			}

			var out ReportResponse
			if err := json.Unmarshal(data, &out); err != nil {
				t.Fatalf("Unmarshal error: %v", err)
			}

			// Compare fields individually for readable failure messages.
			if out.ID != tt.in.ID {
				t.Errorf("ID mismatch: got %v, want %v", out.ID, tt.in.ID)
			}
			if out.Status != tt.in.Status {
				t.Errorf("Status mismatch: got %q, want %q", out.Status, tt.in.Status)
			}
			if out.SnapshotID != tt.in.SnapshotID {
				t.Errorf("SnapshotID mismatch: got %v, want %v", out.SnapshotID, tt.in.SnapshotID)
			}
			if out.DownloadURL != tt.in.DownloadURL {
				t.Errorf("DownloadURL mismatch: got %q, want %q", out.DownloadURL, tt.in.DownloadURL)
			}
			if out.Error != tt.in.Error {
				t.Errorf("Error mismatch: got %q, want %q", out.Error, tt.in.Error)
			}
			if !out.CreatedAt.Equal(tt.in.CreatedAt) {
				t.Errorf("CreatedAt mismatch: got %v, want %v", out.CreatedAt, tt.in.CreatedAt)
			}
			if (out.CompletedAt == nil) != (tt.in.CompletedAt == nil) {
				t.Errorf("CompletedAt nil mismatch: got %v, want %v", out.CompletedAt, tt.in.CompletedAt)
			}
			if out.CompletedAt != nil && tt.in.CompletedAt != nil && !out.CompletedAt.Equal(*tt.in.CompletedAt) {
				t.Errorf("CompletedAt mismatch: got %v, want %v", *out.CompletedAt, *tt.in.CompletedAt)
			}
		})
	}
}

// --- ScanRequest JSON round-trip ---

func TestScanRequest_JSONRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		in   ScanRequest
	}{
		{
			name: "valid host",
			in:   ScanRequest{Host: "example.com"},
		},
		{
			name: "IP address host",
			in:   ScanRequest{Host: "192.168.1.1"},
		},
		{
			name: "empty host",
			in:   ScanRequest{Host: ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.in)
			if err != nil {
				t.Fatalf("Marshal error: %v", err)
			}

			var out ScanRequest
			if err := json.Unmarshal(data, &out); err != nil {
				t.Fatalf("Unmarshal error: %v", err)
			}

			if out.Host != tt.in.Host {
				t.Errorf("Host mismatch: got %q, want %q", out.Host, tt.in.Host)
			}
		})
	}
}
