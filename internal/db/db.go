package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB wraps a pgx connection pool.
type DB struct {
	Pool *pgxpool.Pool
}

// Connect establishes a connection pool to PostgreSQL.
func Connect(databaseURL string) (*DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}

	config.MaxConns = 10
	config.MinConns = 2

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &DB{Pool: pool}, nil
}

// Close shuts down the connection pool.
func (db *DB) Close() {
	db.Pool.Close()
}

// Migrate runs the embedded schema migrations.
func Migrate(db *DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	_, err := db.Pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS targets (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			host TEXT NOT NULL UNIQUE,
			normalized_host TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS snapshots (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			target_id UUID NOT NULL REFERENCES targets(id) ON DELETE CASCADE,
			scanned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			last_seen TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			data_hash TEXT NOT NULL,
			raw_data JSONB NOT NULL,
			changes TEXT[] DEFAULT '{}',
			client_ip TEXT
		);

		CREATE INDEX IF NOT EXISTS idx_snapshots_target_id_scanned_at
			ON snapshots(target_id, scanned_at DESC);

		CREATE INDEX IF NOT EXISTS idx_snapshots_data_hash
			ON snapshots(data_hash);

		ALTER TABLE targets ADD COLUMN IF NOT EXISTS normalized_host TEXT;

		CREATE INDEX IF NOT EXISTS idx_targets_normalized_host
			ON targets(normalized_host);

		CREATE TABLE IF NOT EXISTS reports (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			snapshot_id UUID NOT NULL REFERENCES snapshots(id) ON DELETE CASCADE,
			status TEXT NOT NULL DEFAULT 'pending',
			error_message TEXT,
			pdf BYTEA,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			completed_at TIMESTAMPTZ
		);

		CREATE INDEX IF NOT EXISTS idx_reports_snapshot_id_status
			ON reports(snapshot_id, status);

		ALTER TABLE snapshots ADD COLUMN IF NOT EXISTS pwhois_data JSONB;
		ALTER TABLE snapshots ADD COLUMN IF NOT EXISTS pwhois_looked_up_at TIMESTAMPTZ;
		ALTER TABLE snapshots ADD COLUMN IF NOT EXISTS pwhois_origin_as TEXT;
		ALTER TABLE snapshots ADD COLUMN IF NOT EXISTS pwhois_org_name TEXT;
		ALTER TABLE snapshots ADD COLUMN IF NOT EXISTS pwhois_country_code TEXT;
		ALTER TABLE snapshots ADD COLUMN IF NOT EXISTS pwhois_city TEXT;
		ALTER TABLE snapshots ADD COLUMN IF NOT EXISTS pwhois_prefix TEXT;

		CREATE INDEX IF NOT EXISTS idx_snapshots_pwhois_country_code
			ON snapshots(pwhois_country_code);

		CREATE INDEX IF NOT EXISTS idx_snapshots_pwhois_origin_as
			ON snapshots(pwhois_origin_as);

		CREATE TABLE IF NOT EXISTS webhooks (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name TEXT NOT NULL,
			type TEXT NOT NULL,
			url TEXT NOT NULL,
			enabled BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		ALTER TABLE targets ADD COLUMN IF NOT EXISTS tags TEXT[] DEFAULT '{}';
		CREATE INDEX IF NOT EXISTS idx_targets_tags ON targets USING GIN (tags);

		CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value JSONB NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`)
	if err != nil {
		return fmt.Errorf("execute migrations: %w", err)
	}

	return nil
}
