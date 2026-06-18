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

		ALTER TABLE webhooks ADD COLUMN IF NOT EXISTS config JSONB NOT NULL DEFAULT '{}';

		ALTER TABLE targets ADD COLUMN IF NOT EXISTS tags TEXT[] DEFAULT '{}';
		CREATE INDEX IF NOT EXISTS idx_targets_tags ON targets USING GIN (tags);

		CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value JSONB NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS scan_jobs (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			host TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			snapshot_id UUID REFERENCES snapshots(id) ON DELETE SET NULL,
			target_id UUID REFERENCES targets(id) ON DELETE SET NULL,
			error_message TEXT,
			client_ip TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			completed_at TIMESTAMPTZ
		);

		CREATE INDEX IF NOT EXISTS idx_scan_jobs_status_created
			ON scan_jobs(status, created_at);

		ALTER TABLE snapshots ADD COLUMN IF NOT EXISTS change_details JSONB DEFAULT '[]';

		CREATE TABLE IF NOT EXISTS snapshot_blobs (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			snapshot_id UUID NOT NULL REFERENCES snapshots(id) ON DELETE CASCADE,
			kind TEXT NOT NULL,
			content_type TEXT NOT NULL,
			data BYTEA NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			UNIQUE (snapshot_id, kind)
		);

		CREATE TABLE IF NOT EXISTS target_collections (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name TEXT NOT NULL UNIQUE,
			description TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS target_collection_members (
			collection_id UUID NOT NULL REFERENCES target_collections(id) ON DELETE CASCADE,
			target_id UUID NOT NULL REFERENCES targets(id) ON DELETE CASCADE,
			added_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			PRIMARY KEY (collection_id, target_id)
		);

		CREATE INDEX IF NOT EXISTS idx_target_collection_members_target
			ON target_collection_members(target_id);

		CREATE TABLE IF NOT EXISTS investigation_notes (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			title TEXT NOT NULL DEFAULT '',
			body TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'archived', 'trashed')),
			target_id UUID REFERENCES targets(id) ON DELETE SET NULL,
			snapshot_id UUID REFERENCES snapshots(id) ON DELETE SET NULL,
			collection_id UUID REFERENCES target_collections(id) ON DELETE SET NULL,
			graph_node_id TEXT,
			graph_node_label TEXT,
			graph_node_type TEXT,
			refs JSONB NOT NULL DEFAULT '[]',
			trashed_at TIMESTAMPTZ,
			archived_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			CHECK (
				target_id IS NOT NULL OR
				snapshot_id IS NOT NULL OR
				collection_id IS NOT NULL OR
				graph_node_id IS NOT NULL
			)
		);

		CREATE INDEX IF NOT EXISTS idx_investigation_notes_status_updated
			ON investigation_notes(status, updated_at DESC);

		CREATE INDEX IF NOT EXISTS idx_investigation_notes_target
			ON investigation_notes(target_id)
			WHERE target_id IS NOT NULL;

		CREATE INDEX IF NOT EXISTS idx_investigation_notes_collection
			ON investigation_notes(collection_id)
			WHERE collection_id IS NOT NULL;

		CREATE INDEX IF NOT EXISTS idx_investigation_notes_trashed_at
			ON investigation_notes(trashed_at)
			WHERE status = 'trashed';

		CREATE TABLE IF NOT EXISTS graph_views (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name TEXT NOT NULL UNIQUE,
			description TEXT NOT NULL DEFAULT '',
			view_mode TEXT NOT NULL DEFAULT 'infra',
			target_id UUID REFERENCES targets(id) ON DELETE SET NULL,
			vantage_filter TEXT NOT NULL DEFAULT 'all',
			snapshot_id UUID REFERENCES snapshots(id) ON DELETE SET NULL,
			compare_snapshot_id UUID REFERENCES snapshots(id) ON DELETE SET NULL,
			compare_mode TEXT NOT NULL DEFAULT 'previous',
			pinned_nodes JSONB NOT NULL DEFAULT '{}',
			selected_node_id TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_graph_views_updated
			ON graph_views(updated_at DESC);

		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			display_name TEXT NOT NULL,
			role TEXT NOT NULL CHECK (role IN ('admin', 'scanner')),
			disabled BOOLEAN NOT NULL DEFAULT FALSE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS enrollment_codes (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			code_hash TEXT NOT NULL,
			purpose TEXT NOT NULL DEFAULT 'initial',
			expires_at TIMESTAMPTZ NOT NULL,
			used_at TIMESTAMPTZ,
			created_by UUID REFERENCES users(id) ON DELETE SET NULL,
			attempts INT NOT NULL DEFAULT 0,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS idx_enrollment_codes_hash
			ON enrollment_codes(code_hash);

		CREATE TABLE IF NOT EXISTS webauthn_credentials (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			credential_id BYTEA NOT NULL UNIQUE,
			public_key BYTEA NOT NULL,
			sign_count BIGINT NOT NULL DEFAULT 0,
			nickname TEXT NOT NULL DEFAULT '',
			transport TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			last_used_at TIMESTAMPTZ
		);

		CREATE TABLE IF NOT EXISTS sessions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			token_hash TEXT NOT NULL UNIQUE,
			expires_at TIMESTAMPTZ NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			user_agent TEXT NOT NULL DEFAULT '',
			ip TEXT NOT NULL DEFAULT ''
		);

		CREATE INDEX IF NOT EXISTS idx_sessions_token_hash
			ON sessions(token_hash);

		CREATE TABLE IF NOT EXISTS enrollment_sessions (
			token_hash TEXT PRIMARY KEY,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			expires_at TIMESTAMPTZ NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS webauthn_challenges (
			id TEXT PRIMARY KEY,
			user_id UUID REFERENCES users(id) ON DELETE CASCADE,
			kind TEXT NOT NULL,
			challenge_data JSONB NOT NULL,
			expires_at TIMESTAMPTZ NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`)
	if err != nil {
		return fmt.Errorf("execute migrations: %w", err)
	}

	return nil
}
