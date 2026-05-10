package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"slices"
)

type Migration struct {
	Version int
	Name    string
	SQL     string
}

var migrations = []Migration{
	{
		Version: 1,
		Name:    "create_installation_state",
		SQL: `
CREATE TABLE installation_state (
	id INTEGER PRIMARY KEY CHECK (id = 1),
	managed_root TEXT NOT NULL,
	steamcmd_installed_at TEXT,
	dst_installed_at TEXT,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);`,
	},
	{
		Version: 2,
		Name:    "create_tasks",
		SQL: `
CREATE TABLE tasks (
	id TEXT PRIMARY KEY,
	type TEXT NOT NULL,
	status TEXT NOT NULL,
	detail TEXT NOT NULL,
	error TEXT NOT NULL,
	started_at TEXT,
	finished_at TEXT,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);

CREATE INDEX tasks_created_at_idx ON tasks(created_at);`,
	},
	{
		Version: 3,
		Name:    "create_cluster_config",
		SQL: `
CREATE TABLE cluster_config (
	id INTEGER PRIMARY KEY CHECK (id = 1),
	cluster_name TEXT NOT NULL,
	cluster_description TEXT NOT NULL,
	game_mode TEXT NOT NULL,
	max_players INTEGER NOT NULL,
	language TEXT NOT NULL,
	pvp INTEGER NOT NULL,
	pause_when_empty INTEGER NOT NULL,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);

CREATE TABLE cluster_shards (
	name TEXT PRIMARY KEY,
	enabled INTEGER NOT NULL
);`,
	},
	{
		Version: 4,
		Name:    "create_runtime_events",
		SQL: `
CREATE TABLE runtime_events (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	shard TEXT NOT NULL,
	kind TEXT NOT NULL,
	detail TEXT NOT NULL,
	created_at TEXT NOT NULL
);

CREATE INDEX runtime_events_created_at_idx ON runtime_events(created_at DESC, id DESC);`,
	},
	{
		Version: 5,
		Name:    "create_update_state",
		SQL: `
CREATE TABLE update_state (
	id INTEGER PRIMARY KEY CHECK (id = 1),
	current_version TEXT NOT NULL,
	latest_version TEXT NOT NULL,
	update_available INTEGER NOT NULL,
	last_checked_at TEXT,
	last_updated_at TEXT,
	last_error TEXT NOT NULL,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);`,
	},
	{
		Version: 6,
		Name:    "expand_cluster_config_network_fields",
		SQL: `
ALTER TABLE cluster_config ADD COLUMN cluster_password TEXT NOT NULL DEFAULT '';
ALTER TABLE cluster_config ADD COLUMN cluster_intention TEXT NOT NULL DEFAULT 'cooperative';
ALTER TABLE cluster_config ADD COLUMN offline_cluster INTEGER NOT NULL DEFAULT 0;
ALTER TABLE cluster_config ADD COLUMN lan_only_cluster INTEGER NOT NULL DEFAULT 0;
ALTER TABLE cluster_config ADD COLUMN tick_rate INTEGER NOT NULL DEFAULT 15;
ALTER TABLE cluster_config ADD COLUMN console_enabled INTEGER NOT NULL DEFAULT 1;
ALTER TABLE cluster_config ADD COLUMN bind_ip TEXT NOT NULL DEFAULT '127.0.0.1';
ALTER TABLE cluster_config ADD COLUMN master_port INTEGER NOT NULL DEFAULT 10888;
ALTER TABLE cluster_config ADD COLUMN cluster_key TEXT NOT NULL DEFAULT 'dst-server-ctl';

ALTER TABLE cluster_shards ADD COLUMN server_port INTEGER NOT NULL DEFAULT 10999;
ALTER TABLE cluster_shards ADD COLUMN master_server_port INTEGER NOT NULL DEFAULT 27016;
ALTER TABLE cluster_shards ADD COLUMN authentication_port INTEGER NOT NULL DEFAULT 8766;

UPDATE cluster_shards
SET
	server_port = CASE name WHEN 'Master' THEN 10999 WHEN 'Caves' THEN 11000 ELSE server_port END,
	master_server_port = CASE name WHEN 'Master' THEN 27016 WHEN 'Caves' THEN 27017 ELSE master_server_port END,
	authentication_port = CASE name WHEN 'Master' THEN 8766 WHEN 'Caves' THEN 8767 ELSE authentication_port END;`,
	},
	{
		Version: 7,
		Name:    "add_worldgen_config",
		SQL: `
ALTER TABLE cluster_shards ADD COLUMN worldgen_preset TEXT NOT NULL DEFAULT 'SURVIVAL_TOGETHER';

UPDATE cluster_shards
SET worldgen_preset = CASE name WHEN 'Master' THEN 'SURVIVAL_TOGETHER' WHEN 'Caves' THEN 'DST_CAVE' ELSE worldgen_preset END;

CREATE TABLE cluster_shard_world_overrides (
	shard_name TEXT NOT NULL,
	override_key TEXT NOT NULL,
	override_value TEXT NOT NULL,
	PRIMARY KEY (shard_name, override_key)
);`,
	},
}

func Migrate(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS schema_migrations (
	version INTEGER PRIMARY KEY,
	name TEXT NOT NULL,
	applied_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);`); err != nil {
		return fmt.Errorf("create schema migrations table: %w", err)
	}

	applied, err := appliedMigrations(ctx, db)
	if err != nil {
		return err
	}

	for _, migration := range migrations {
		if slices.Contains(applied, migration.Version) {
			continue
		}

		if err := applyMigration(ctx, db, migration); err != nil {
			return err
		}
	}

	return nil
}

func appliedMigrations(ctx context.Context, db *sql.DB) ([]int, error) {
	rows, err := db.QueryContext(ctx, `SELECT version FROM schema_migrations ORDER BY version`)
	if err != nil {
		return nil, fmt.Errorf("list applied migrations: %w", err)
	}
	defer rows.Close()

	var versions []int
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("scan applied migration: %w", err)
		}
		versions = append(versions, version)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate applied migrations: %w", err)
	}

	return versions, nil
}

func applyMigration(ctx context.Context, db *sql.DB, migration Migration) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migration %d: %w", migration.Version, err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, migration.SQL); err != nil {
		return fmt.Errorf("apply migration %d %s: %w", migration.Version, migration.Name, err)
	}
	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO schema_migrations (version, name) VALUES (?, ?)`,
		migration.Version,
		migration.Name,
	); err != nil {
		return fmt.Errorf("record migration %d %s: %w", migration.Version, migration.Name, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %d %s: %w", migration.Version, migration.Name, err)
	}

	return nil
}
