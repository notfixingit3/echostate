package scans

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/notfixingit3/echostate/internal/config"
	"github.com/notfixingit3/echostate/internal/db"
	"github.com/notfixingit3/echostate/internal/models"
	obstrace "github.com/notfixingit3/echostate/internal/observability/trace"
)

const (
	defaultMaxConcurrency = 2
	wakeBufferSize        = 1
	pollInterval          = 2 * time.Second
	scanJobTimeout        = 120 * time.Second
	cancelledByUserMsg    = "cancelled by user"
)

// ErrScanNotCancellable is returned when a scan job is already finished.
var ErrScanNotCancellable = fmt.Errorf("scan cannot be cancelled")

// Runner executes a full reconnaissance scan for a host.
type Runner interface {
	Run(ctx context.Context, host string) (*models.ScanResult, error)
}

// Persister stores scan results as snapshots.
type Persister interface {
	PersistScan(ctx context.Context, clientIP string, result *models.ScanResult) (*models.Snapshot, error)
}

// Worker processes scan jobs asynchronously with bounded concurrency.
type Worker struct {
	db             *db.DB
	runner         Runner
	persister      Persister
	maxConcurrency int

	wake   chan struct{}
	sem    chan struct{}
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc

	mu      sync.Mutex
	started bool
}

// New creates a scan worker. If maxConcurrency is <= 0, it defaults to 2.
func New(database *db.DB, runner Runner, persister Persister, maxConcurrency int) *Worker {
	if maxConcurrency <= 0 {
		maxConcurrency = defaultMaxConcurrency
	}
	return &Worker{
		db:             database,
		runner:         runner,
		persister:      persister,
		maxConcurrency: maxConcurrency,
		wake:           make(chan struct{}, wakeBufferSize),
		sem:            make(chan struct{}, maxConcurrency),
	}
}

// Start begins the background processing loop.
func (w *Worker) Start(ctx context.Context) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.started {
		return fmt.Errorf("scan worker already started")
	}

	if err := w.failStaleRunning(ctx); err != nil {
		return fmt.Errorf("fail stale running scans: %w", err)
	}

	w.ctx, w.cancel = context.WithCancel(ctx)
	w.started = true

	w.wg.Add(1)
	go w.loop()

	return nil
}

// Stop gracefully shuts down the worker.
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

// CreateJob inserts a pending scan job and wakes the worker.
func (w *Worker) CreateJob(ctx context.Context, host, clientIP string, traceparent ...string) (uuid.UUID, error) {
	var jobID uuid.UUID
	jobTraceparent := ""
	if len(traceparent) > 0 {
		jobTraceparent = traceparent[0]
	}

	err := w.db.Pool.QueryRow(ctx, `
		INSERT INTO scan_jobs (host, status, client_ip, traceparent)
		VALUES ($1, $2, $3, NULLIF($4, ''))
		RETURNING id
	`, host, models.ScanJobPending, clientIP, jobTraceparent).Scan(&jobID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("insert scan job: %w", err)
	}

	select {
	case w.wake <- struct{}{}:
	default:
	}

	return jobID, nil
}

// CancelJob marks a pending or running scan job as failed (cancelled by user).
func (w *Worker) CancelJob(ctx context.Context, jobID uuid.UUID) (*models.ScanJob, error) {
	tag, err := w.db.Pool.Exec(ctx, `
		UPDATE scan_jobs
		SET status = $1,
		    error_message = $2,
		    updated_at = NOW(),
		    completed_at = NOW()
		WHERE id = $3 AND status IN ($4, $5)
	`, models.ScanJobFailed, cancelledByUserMsg, jobID, models.ScanJobPending, models.ScanJobRunning)
	if err != nil {
		return nil, fmt.Errorf("cancel scan job: %w", err)
	}
	if tag.RowsAffected() == 0 {
		job, err := w.GetJob(ctx, jobID)
		if err != nil {
			return nil, err
		}
		if job == nil {
			return nil, nil
		}
		return nil, ErrScanNotCancellable
	}

	return w.GetJob(ctx, jobID)
}

// GetJob returns a scan job by ID.
func (w *Worker) GetJob(ctx context.Context, jobID uuid.UUID) (*models.ScanJob, error) {
	job, err := w.scanJob(w.db.Pool.QueryRow(ctx, `
		SELECT id, host, status, snapshot_id, target_id,
			COALESCE(error_message, ''), COALESCE(client_ip, ''),
			created_at, completed_at
		FROM scan_jobs
		WHERE id = $1
	`, jobID))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get scan job: %w", err)
	}
	return job, nil
}

// ProcessNext claims and runs one pending job. Intended for tests.
func (w *Worker) ProcessNext(ctx context.Context) (bool, error) {
	jobID, host, clientIP, traceparent, ok, err := w.claimNextPending(ctx)
	if err != nil || !ok {
		return ok, err
	}
	w.runJob(ctx, jobID, host, clientIP, traceparent)
	return true, nil
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

		jobID, host, clientIP, traceparent, ok, err := w.claimNextPending(ctx)
		if err != nil {
			slog.Default().Error("scan worker: claim next pending", slog.String("component", "scan_worker"), slog.String("error", err.Error()))
			<-w.sem
			return
		}
		if !ok {
			<-w.sem
			return
		}

		w.wg.Add(1)
		go func(jobID uuid.UUID, host, clientIP, traceparent string) {
			defer func() {
				<-w.sem
				w.wg.Done()
			}()
			w.runJob(ctx, jobID, host, clientIP, traceparent)
		}(jobID, host, clientIP, traceparent)
	}
}

func (w *Worker) claimNextPending(ctx context.Context) (jobID uuid.UUID, host, clientIP, traceparent string, ok bool, err error) {
	tx, err := w.db.Pool.Begin(ctx)
	if err != nil {
		return uuid.Nil, "", "", "", false, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		SELECT id, host, client_ip, COALESCE(traceparent, '')
		FROM scan_jobs
		WHERE status = $1
		ORDER BY created_at ASC
		LIMIT 1
		FOR UPDATE SKIP LOCKED
	`, models.ScanJobPending).Scan(&jobID, &host, &clientIP, &traceparent)
	if err != nil {
		if err == pgx.ErrNoRows {
			return uuid.Nil, "", "", "", false, nil
		}
		return uuid.Nil, "", "", "", false, fmt.Errorf("select pending scan: %w", err)
	}

	_, err = tx.Exec(ctx, `
		UPDATE scan_jobs
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`, models.ScanJobRunning, jobID)
	if err != nil {
		return uuid.Nil, "", "", "", false, fmt.Errorf("mark scan running: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, "", "", "", false, fmt.Errorf("commit claim: %w", err)
	}

	return jobID, host, clientIP, traceparent, true, nil
}

func (w *Worker) runJob(parent context.Context, jobID uuid.UUID, host, clientIP, traceparent string) {
	if parent == nil {
		parent = context.Background()
	}

	ctx := obstrace.ContextFromTraceparent(traceparent)
	ctx, span := obstrace.Tracer().Start(ctx, "scan_worker_process")
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, scanJobTimeoutFromSettings())
	defer cancel()
	stopParentCancel := context.AfterFunc(parent, cancel)
	defer stopParentCancel()

	result, err := w.runner.Run(ctx, host)
	if err != nil {
		w.failJob(ctx, jobID, fmt.Sprintf("scan failed: %v", err))
		return
	}

	if cancelled, err := w.jobWasCancelled(ctx, jobID); err != nil {
		slog.Default().Error("scan worker: check cancelled", slog.String("component", "scan_worker"), slog.String("job_id", jobID.String()), slog.String("error", err.Error()))
	} else if cancelled {
		return
	}

	snapshot, err := w.persister.PersistScan(ctx, clientIP, result)
	if err != nil {
		w.failJob(ctx, jobID, fmt.Sprintf("persist snapshot: %v", err))
		return
	}

	_, err = w.db.Pool.Exec(ctx, `
		UPDATE scan_jobs
		SET status = $1,
		    snapshot_id = $2,
		    target_id = $3,
		    error_message = NULL,
		    updated_at = NOW(),
		    completed_at = NOW()
		WHERE id = $4
	`, models.ScanJobCompleted, snapshot.ID, snapshot.TargetID, jobID)
	if err != nil {
		slog.Default().Error("scan worker: mark completed", slog.String("component", "scan_worker"), slog.String("job_id", jobID.String()), slog.String("error", err.Error()))
	}
}

func (w *Worker) failJob(ctx context.Context, jobID uuid.UUID, message string) {
	_, err := w.db.Pool.Exec(ctx, `
		UPDATE scan_jobs
		SET status = $1,
		    error_message = $2,
		    updated_at = NOW(),
		    completed_at = NOW()
		WHERE id = $3
	`, models.ScanJobFailed, message, jobID)
	if err != nil {
		slog.Default().Error("scan worker: mark failed", slog.String("component", "scan_worker"), slog.String("job_id", jobID.String()), slog.String("error", err.Error()))
	}
}

func (w *Worker) jobWasCancelled(ctx context.Context, jobID uuid.UUID) (bool, error) {
	var status models.ScanJobStatus
	var message string
	err := w.db.Pool.QueryRow(ctx, `
		SELECT status, COALESCE(error_message, '')
		FROM scan_jobs
		WHERE id = $1
	`, jobID).Scan(&status, &message)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return status == models.ScanJobFailed && message == cancelledByUserMsg, nil
}

func (w *Worker) failStaleRunning(ctx context.Context) error {
	_, err := w.db.Pool.Exec(ctx, `
		UPDATE scan_jobs
		SET status = $1,
		    error_message = 'interrupted by restart',
		    updated_at = NOW(),
		    completed_at = NOW()
		WHERE status = $2
	`, models.ScanJobFailed, models.ScanJobRunning)
	return err
}

func (w *Worker) scanJob(row pgx.Row) (*models.ScanJob, error) {
	var job models.ScanJob
	err := row.Scan(
		&job.ID,
		&job.Host,
		&job.Status,
		&job.SnapshotID,
		&job.TargetID,
		&job.ErrorMessage,
		&job.ClientIP,
		&job.CreatedAt,
		&job.CompletedAt,
	)
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func scanJobTimeoutFromSettings() time.Duration {
	sec := config.GetSettings().ScanJobTimeoutSec
	if sec > 0 {
		return time.Duration(sec) * time.Second
	}
	return scanJobTimeout
}
