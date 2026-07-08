package scheduler

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/notfixingit3/echostate/internal/config"
	"github.com/notfixingit3/echostate/internal/db"
	"github.com/notfixingit3/echostate/internal/scans"
)

// Scheduler enqueues rescans for stale targets on an interval.
type Scheduler struct {
	db         *db.DB
	scanWorker *scans.Worker

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func New(database *db.DB, scanWorker *scans.Worker) *Scheduler {
	return &Scheduler{db: database, scanWorker: scanWorker}
}

func (s *Scheduler) Start(ctx context.Context) {
	s.ctx, s.cancel = context.WithCancel(ctx)
	s.wg.Add(1)
	go s.loop()
}

func (s *Scheduler) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
	s.wg.Wait()
}

func (s *Scheduler) loop() {
	defer s.wg.Done()

	s.tick()
	for {
		settings := config.GetSettings()
		interval := time.Duration(settings.ScheduleIntervalMinutes) * time.Minute
		if interval <= 0 {
			interval = time.Hour
		}
		timer := time.NewTimer(interval)
		select {
		case <-s.ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			s.tick()
		}
	}
}

func (s *Scheduler) tick() {
	settings := config.GetSettings()
	if !settings.ScheduleEnabled {
		return
	}

	ctx, cancel := context.WithTimeout(s.ctx, 2*time.Minute)
	defer cancel()

	hosts, err := s.staleTargets(ctx, settings.ScheduleStaleHours, settings.ScheduleTags)
	if err != nil {
		slog.Default().Error("scheduler: list stale targets", slog.String("component", "scheduler"), slog.String("error", err.Error()))
		return
	}

	for _, host := range hosts {
		if _, err := s.scanWorker.CreateJob(ctx, host, "scheduler"); err != nil {
			slog.Default().Error("scheduler: enqueue", slog.String("component", "scheduler"), slog.String("host", host), slog.String("error", err.Error()))
		}
	}
}

func (s *Scheduler) staleTargets(ctx context.Context, staleHours int, tags []string) ([]string, error) {
	if staleHours <= 0 {
		staleHours = 24
	}

	query := `
		SELECT t.host
		FROM targets t
		LEFT JOIN LATERAL (
			SELECT scanned_at
			FROM snapshots
			WHERE target_id = t.id
			ORDER BY scanned_at DESC
			LIMIT 1
		) latest ON true
		WHERE (
			latest.scanned_at IS NULL
			OR latest.scanned_at < NOW() - ($1::text || ' hours')::interval
		)
		AND NOT EXISTS (
			SELECT 1 FROM scan_jobs j
			WHERE j.host = t.host
			  AND j.status IN ('pending', 'running')
		)
	`
	args := []any{staleHours}

	if len(tags) > 0 {
		query += ` AND t.tags && $2::text[]`
		args = append(args, tags)
	}

	query += ` ORDER BY latest.scanned_at NULLS FIRST LIMIT 25`

	rows, err := s.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hosts []string
	for rows.Next() {
		var host string
		if err := rows.Scan(&host); err != nil {
			return nil, err
		}
		hosts = append(hosts, host)
	}
	return hosts, rows.Err()
}
