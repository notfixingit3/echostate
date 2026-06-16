package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/notfixingit3/echostate/internal/db"
	"github.com/notfixingit3/echostate/internal/models"
	"github.com/notfixingit3/echostate/internal/pdf"
	"github.com/notfixingit3/echostate/internal/reports"
	"github.com/notfixingit3/echostate/internal/scanner"
)

// newWorker builds the report worker used by the handler.
func newWorker(database *db.DB) *reports.Worker {
	return reports.New(database, pdf.RenderReport, 3)
}

// createReport enqueues a new PDF report for a snapshot or a host.
func (h *Handler) createReport(c *gin.Context) {
	var req models.CreateReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if (req.SnapshotID == nil && req.Host == nil) || (req.SnapshotID != nil && req.Host != nil) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "provide exactly one of snapshot_id or host"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var snapshotID uuid.UUID
	if req.SnapshotID != nil {
		exists, err := h.snapshotExists(ctx, *req.SnapshotID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify snapshot"})
			return
		}
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "snapshot not found"})
			return
		}
		snapshotID = *req.SnapshotID
	} else {
		id, err := h.findLatestSnapshotByHost(ctx, *req.Host)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to look up snapshot"})
			return
		}
		if id == uuid.Nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "no snapshot found for host"})
			return
		}
		snapshotID = id
	}

	reportID, err := h.worker.CreateReport(ctx, snapshotID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create report"})
		return
	}

	c.JSON(http.StatusAccepted, models.ReportResponse{
		ID:         reportID,
		Status:     models.ReportPending,
		SnapshotID: snapshotID,
		CreatedAt:  time.Now().UTC(),
	})
}

// getReport returns the current status of a report.
func (h *Handler) getReport(c *gin.Context) {
	reportID, ok := h.parseReportID(c)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	report, err := h.worker.GetReport(ctx, reportID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch report"})
		return
	}
	if report == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "report not found"})
		return
	}

	resp := h.toReportResponse(report)
	c.JSON(http.StatusOK, resp)
}

// downloadReport streams the generated PDF for a completed report.
func (h *Handler) downloadReport(c *gin.Context) {
	reportID, ok := h.parseReportID(c)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	report, err := h.worker.GetReport(ctx, reportID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch report"})
		return
	}
	if report == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "report not found"})
		return
	}

	if report.Status != models.ReportCompleted {
		c.JSON(http.StatusConflict, gin.H{"error": "report not ready"})
		return
	}

	h.serveReportPDF(c, report)
}

// snapshotReport returns an existing completed report for a snapshot or enqueues a new one.
func (h *Handler) snapshotReport(c *gin.Context) {
	snapshotID, ok := h.parseSnapshotID(c)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	exists, err := h.snapshotExists(ctx, snapshotID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify snapshot"})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "snapshot not found"})
		return
	}

	reportsList, err := h.worker.ListReportsForSnapshot(ctx, snapshotID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list reports"})
		return
	}

	for _, r := range reportsList {
		if r.Status == models.ReportCompleted {
			h.serveReportPDF(c, r)
			return
		}
	}

	reportID, err := h.worker.CreateReport(ctx, snapshotID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create report"})
		return
	}

	c.JSON(http.StatusAccepted, models.ReportResponse{
		ID:         reportID,
		Status:     models.ReportPending,
		SnapshotID: snapshotID,
		CreatedAt:  time.Now().UTC(),
	})
}

// serveReportPDF sets PDF headers and writes the report bytes.
func (h *Handler) serveReportPDF(c *gin.Context, report *models.Report) {
	host, err := h.snapshotHost(c.Request.Context(), report.SnapshotID)
	if err != nil {
		host = "unknown"
	}
	if host == "" {
		host = "unknown"
	}

	date := time.Now().UTC().Format("2006-01-02")
	filename := fmt.Sprintf("echostate-%s-%s.pdf", host, date)

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Data(http.StatusOK, "application/pdf", report.PDF)
}

// toReportResponse maps a Report to its API response.
func (h *Handler) toReportResponse(report *models.Report) models.ReportResponse {
	resp := models.ReportResponse{
		ID:          report.ID,
		Status:      report.Status,
		SnapshotID:  report.SnapshotID,
		CreatedAt:   report.CreatedAt,
		CompletedAt: report.CompletedAt,
	}
	if report.Status == models.ReportFailed {
		resp.Error = "report generation failed"
	}
	if report.Status == models.ReportCompleted {
		resp.DownloadURL = fmt.Sprintf("/api/reports/%s/download", report.ID)
	}
	return resp
}

// parseReportID parses a UUID route parameter named "id".
func (h *Handler) parseReportID(c *gin.Context) (uuid.UUID, bool) {
	return h.parseUUIDParam(c, "id")
}

// parseSnapshotID parses a UUID route parameter named "id" for snapshot routes.
func (h *Handler) parseSnapshotID(c *gin.Context) (uuid.UUID, bool) {
	return h.parseUUIDParam(c, "id")
}

// parseUUIDParam parses a UUID route parameter and returns 400 on failure.
func (h *Handler) parseUUIDParam(c *gin.Context, key string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(key))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return uuid.Nil, false
	}
	return id, true
}

// snapshotExists checks whether a snapshot with the given ID exists.
func (h *Handler) snapshotExists(ctx context.Context, snapshotID uuid.UUID) (bool, error) {
	var exists bool
	err := h.db.Pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM snapshots WHERE id = $1)
	`, snapshotID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check snapshot existence: %w", err)
	}
	return exists, nil
}

// snapshotHost returns the raw host for a snapshot.
func (h *Handler) snapshotHost(ctx context.Context, snapshotID uuid.UUID) (string, error) {
	var host string
	err := h.db.Pool.QueryRow(ctx, `
		SELECT t.host
		FROM snapshots s
		JOIN targets t ON t.id = s.target_id
		WHERE s.id = $1
	`, snapshotID).Scan(&host)
	if err != nil {
		return "", fmt.Errorf("fetch snapshot host: %w", err)
	}
	return host, nil
}

// findLatestSnapshotByHost finds the most recent snapshot for a normalized host.
func (h *Handler) findLatestSnapshotByHost(ctx context.Context, host string) (uuid.UUID, error) {
	normalized := scanner.NormalizeHost(host)

	var snapshotID uuid.UUID
	err := h.db.Pool.QueryRow(ctx, `
		SELECT s.id
		FROM snapshots s
		JOIN targets t ON t.id = s.target_id
		WHERE t.normalized_host = $1
		ORDER BY s.scanned_at DESC
		LIMIT 1
	`, normalized).Scan(&snapshotID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return uuid.Nil, nil
		}
		return uuid.Nil, fmt.Errorf("find latest snapshot: %w", err)
	}
	return snapshotID, nil
}
