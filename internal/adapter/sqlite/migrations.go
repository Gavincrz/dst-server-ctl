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
