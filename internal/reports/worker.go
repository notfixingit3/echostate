package reports

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/notfixingit3/echostate/internal/db"
	"github.com/notfixingit3/echostate/internal/models"
)

const (
	defaultMaxConcurrency = 3
	maxPDFSize            = 10 * 1024 * 1024
	renderTimeout         = 30 * time.Second
	wakeBufferSize        = 1
	pollInterval          = 2 * time.Second
)

// Renderer renders a ScanResult into a PDF byte slice.
type Renderer func(*models.ScanResult) ([]byte, error)

// Worker processes PDF report jobs asynchronously with bounded concurrency.
type Worker struct {
	db             *db.DB
	renderer       Renderer
	maxConcurrency int

	wake   chan struct{}
	sem    chan struct{}
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc

	mu      sync.Mutex
	started bool
}

// New creates a Worker. If maxConcurrency is <= 0, it defaults to 3.
func New(db *db.DB, renderer Renderer, maxConcurrency int) *Worker {
	if maxConcurrency <= 0 {
		maxConcurrency = defaultMaxConcurrency
	}
	return &Worker{
		db:             db,
		renderer:       renderer,
		maxConcurrency: maxConcurrency,
		wake:           make(chan struct{}, wakeBufferSize),
		sem:            make(chan struct{}, maxConcurrency),
	}
}

// Start begins the background processing loop. It returns once any stale
// running reports have been marked failed and the loop has started.
func (w *Worker) Start(ctx context.Context) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.started {
		return fmt.Errorf("worker already started")
	}

	if err := w.failStaleRunning(ctx); err != nil {
		return fmt.Errorf("fail stale running reports: %w", err)
	}

	w.ctx, w.cancel = context.WithCancel(ctx)
	w.started = true

	w.wg.Add(1)
	go w.loop()

	return nil
}

// Stop gracefully shuts down the worker, waiting for in-flight jobs.
func (w *Worker) Stop() {
	w.mu.Lock()
	if !w.started {
		w.mu.Unlock()
		return
	}
	w.cancel()
	w.mu.Unlock()

	w.wg.Wait()

	w.mu.Lock()
	w.started = false
	w.mu.Unlock()
}

// CreateReport inserts a pending report for the given snapshot and wakes the
// worker. It does not wait for generation to finish.
func (w *Worker) CreateReport(ctx context.Context, snapshotID uuid.UUID) (uuid.UUID, error) {
	var reportID uuid.UUID
	err := w.db.Pool.QueryRow(ctx, `
		INSERT INTO reports (snapshot_id, status)
		VALUES ($1, $2)
		RETURNING id
	`, snapshotID, models.ReportPending).Scan(&reportID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("insert report: %w", err)
	}

	select {
	case w.wake <- struct{}{}:
	default:
	}

	return reportID, nil
}

// GetReport fetches a single report by ID.
func (w *Worker) GetReport(ctx context.Context, reportID uuid.UUID) (*models.Report, error) {
	report, err := w.scanReport(w.db.Pool.QueryRow(ctx, `
		SELECT id, snapshot_id, status, error_message, pdf, created_at, completed_at
		FROM reports
		WHERE id = $1
	`, reportID))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get report: %w", err)
	}
	return report, nil
}

// ListReportsForSnapshot returns all reports for a snapshot, newest first.
func (w *Worker) ListReportsForSnapshot(ctx context.Context, snapshotID uuid.UUID) ([]*models.Report, error) {
	rows, err := w.db.Pool.Query(ctx, `
		SELECT id, snapshot_id, status, error_message, pdf, created_at, completed_at
		FROM reports
		WHERE snapshot_id = $1
		ORDER BY created_at DESC
	`, snapshotID)
	if err != nil {
		return nil, fmt.Errorf("list reports: %w", err)
	}
	defer rows.Close()

	var reports []*models.Report
	for rows.Next() {
		report, err := w.scanReport(rows)
		if err != nil {
			return nil, fmt.Errorf("scan report: %w", err)
		}
		reports = append(reports, report)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate reports: %w", err)
	}

	return reports, nil
}

func (w *Worker) loop() {
	defer w.wg.Done()

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-w.wake:
			w.processPending(w.ctx)
		case <-time.After(pollInterval):
			w.processPending(w.ctx)
		}
	}
}

func (w *Worker) processPending(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case w.sem <- struct{}{}:
		}

		reportID, snapshotID, ok, err := w.claimNextPending(ctx)
		if err != nil {
			<-w.sem
			return
		}
		if !ok {
			<-w.sem
			return
		}

		w.wg.Add(1)
		go func() {
			defer func() {
				<-w.sem
				w.wg.Done()
			}()
			w.runJob(reportID, snapshotID)
		}()
	}
}

// claimNextPending atomically selects and locks the next pending report,
// updates its status to running, and returns the report and snapshot IDs.
// If no pending report exists, ok is false.
func (w *Worker) claimNextPending(ctx context.Context) (reportID uuid.UUID, snapshotID uuid.UUID, ok bool, err error) {
	tx, err := w.db.Pool.Begin(ctx)
	if err != nil {
		return uuid.Nil, uuid.Nil, false, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		SELECT id, snapshot_id
		FROM reports
		WHERE status = $1
		ORDER BY created_at ASC
		LIMIT 1
		FOR UPDATE SKIP LOCKED
	`, models.ReportPending).Scan(&reportID, &snapshotID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return uuid.Nil, uuid.Nil, false, nil
		}
		return uuid.Nil, uuid.Nil, false, fmt.Errorf("select pending report: %w", err)
	}

	_, err = tx.Exec(ctx, `
		UPDATE reports
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`, models.ReportRunning, reportID)
	if err != nil {
		return uuid.Nil, uuid.Nil, false, fmt.Errorf("mark report running: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, uuid.Nil, false, fmt.Errorf("commit claim: %w", err)
	}

	return reportID, snapshotID, true, nil
}

func (w *Worker) runJob(reportID uuid.UUID, snapshotID uuid.UUID) {
	ctx, cancel := context.WithTimeout(w.ctx, renderTimeout)
	defer cancel()

	var raw json.RawMessage
	err := w.db.Pool.QueryRow(ctx, `
		SELECT raw_data FROM snapshots WHERE id = $1
	`, snapshotID).Scan(&raw)
	if err != nil {
		msg := fmt.Sprintf("snapshot not found: %v", err)
		if err == pgx.ErrNoRows {
			msg = "snapshot not found"
		}
		w.failReport(ctx, reportID, msg)
		return
	}

	var result models.ScanResult
	if err := json.Unmarshal(raw, &result); err != nil {
		w.failReport(ctx, reportID, fmt.Sprintf("decode snapshot data: %v", err))
		return
	}

	pdfBytes, err := w.renderer(&result)
	if err != nil {
		w.failReport(ctx, reportID, fmt.Sprintf("render PDF: %v", err))
		return
	}

	if len(pdfBytes) > maxPDFSize {
		w.failReport(ctx, reportID, "generated PDF exceeds max size")
		return
	}

	_, err = w.db.Pool.Exec(ctx, `
		UPDATE reports
		SET status = $1, pdf = $2, completed_at = NOW(), updated_at = NOW()
		WHERE id = $3
	`, models.ReportCompleted, pdfBytes, reportID)
	if err != nil {
		w.failReport(ctx, reportID, fmt.Sprintf("store completed report: %v", err))
		return
	}
}

func (w *Worker) failReport(ctx context.Context, reportID uuid.UUID, message string) {
	_, err := w.db.Pool.Exec(ctx, `
		UPDATE reports
		SET status = $1, error_message = $2, updated_at = NOW()
		WHERE id = $3
	`, models.ReportFailed, message, reportID)
	if err != nil {
		// Nothing useful to do at this point; the worker context may be cancelled.
		_ = err
	}
}

func (w *Worker) failStaleRunning(ctx context.Context) error {
	_, err := w.db.Pool.Exec(ctx, `
		UPDATE reports
		SET status = $1,
		    error_message = $2,
		    updated_at = NOW()
		WHERE status = $3
	`, models.ReportFailed, "worker restarted before completion", models.ReportRunning)
	if err != nil {
		return fmt.Errorf("mark stale running reports failed: %w", err)
	}
	return nil
}

func (w *Worker) scanReport(row pgx.Row) (*models.Report, error) {
	var report models.Report
	var errorMessage *string
	var completedAt *time.Time

	err := row.Scan(
		&report.ID,
		&report.SnapshotID,
		&report.Status,
		&errorMessage,
		&report.PDF,
		&report.CreatedAt,
		&completedAt,
	)
	if err != nil {
		return nil, err
	}

	if errorMessage != nil {
		report.ErrorMessage = *errorMessage
	}
	report.CompletedAt = completedAt

	return &report, nil
}
