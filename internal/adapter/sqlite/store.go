package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"dst-server-ctl/internal/domain"

	_ "modernc.org/sqlite"
)

const timeFormat = time.RFC3339Nano

type Store struct {
	db *sql.DB
}

func Open(ctx context.Context, path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}

	if _, err := db.ExecContext(ctx, `PRAGMA foreign_keys = ON`); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable sqlite foreign keys: %w", err)
	}

	if err := Migrate(ctx, db); err != nil {
		db.Close()
		return nil, err
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) GetInstallationState(ctx context.Context) (domain.InstallationState, error) {
	var row installationStateRow
	err := s.db.QueryRowContext(ctx, `
SELECT managed_root, steamcmd_installed_at, dst_installed_at, created_at, updated_at
FROM installation_state
WHERE id = 1`).Scan(
		&row.ManagedRoot,
		&row.SteamCMDInstalledAt,
		&row.DSTInstalledAt,
		&row.CreatedAt,
		&row.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.InstallationState{}, domain.ErrInstallationStateNotFound
	}
	if err != nil {
		return domain.InstallationState{}, fmt.Errorf("get installation state: %w", err)
	}

	state, err := row.toDomain()
	if err != nil {
		return domain.InstallationState{}, err
	}

	return state, nil
}

func (s *Store) SaveInstallationState(ctx context.Context, state domain.InstallationState) error {
	row := installationStateRowFromDomain(state)
	_, err := s.db.ExecContext(ctx, `
INSERT INTO installation_state (
	id,
	managed_root,
	steamcmd_installed_at,
	dst_installed_at,
	created_at,
	updated_at
) VALUES (1, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
	managed_root = excluded.managed_root,
	steamcmd_installed_at = excluded.steamcmd_installed_at,
	dst_installed_at = excluded.dst_installed_at,
	updated_at = excluded.updated_at`,
		row.ManagedRoot,
		row.SteamCMDInstalledAt,
		row.DSTInstalledAt,
		row.CreatedAt,
		row.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("save installation state: %w", err)
	}

	return nil
}

type installationStateRow struct {
	ManagedRoot         string
	SteamCMDInstalledAt sql.NullString
	DSTInstalledAt      sql.NullString
	CreatedAt           string
	UpdatedAt           string
}

func installationStateRowFromDomain(state domain.InstallationState) installationStateRow {
	return installationStateRow{
		ManagedRoot:         state.ManagedRoot,
		SteamCMDInstalledAt: nullableTime(state.SteamCMDInstalledAt),
		DSTInstalledAt:      nullableTime(state.DSTInstalledAt),
		CreatedAt:           state.CreatedAt.UTC().Format(timeFormat),
		UpdatedAt:           state.UpdatedAt.UTC().Format(timeFormat),
	}
}

func (r installationStateRow) toDomain() (domain.InstallationState, error) {
	createdAt, err := parseRequiredTime("created_at", r.CreatedAt)
	if err != nil {
		return domain.InstallationState{}, err
	}
	updatedAt, err := parseRequiredTime("updated_at", r.UpdatedAt)
	if err != nil {
		return domain.InstallationState{}, err
	}

	steamCMDInstalledAt, err := parseNullableTime("steamcmd_installed_at", r.SteamCMDInstalledAt)
	if err != nil {
		return domain.InstallationState{}, err
	}
	dstInstalledAt, err := parseNullableTime("dst_installed_at", r.DSTInstalledAt)
	if err != nil {
		return domain.InstallationState{}, err
	}

	return domain.InstallationState{
		ManagedRoot:         r.ManagedRoot,
		SteamCMDInstalledAt: steamCMDInstalledAt,
		DSTInstalledAt:      dstInstalledAt,
		CreatedAt:           createdAt,
		UpdatedAt:           updatedAt,
	}, nil
}

func nullableTime(value *time.Time) sql.NullString {
	if value == nil {
		return sql.NullString{}
	}

	return sql.NullString{
		String: value.UTC().Format(timeFormat),
		Valid:  true,
	}
}

func parseNullableTime(name string, value sql.NullString) (*time.Time, error) {
	if !value.Valid {
		return nil, nil
	}

	parsed, err := parseRequiredTime(name, value.String)
	if err != nil {
		return nil, err
	}

	return &parsed, nil
}

func parseRequiredTime(name string, value string) (time.Time, error) {
	parsed, err := time.Parse(timeFormat, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse %s: %w", name, err)
	}

	return parsed, nil
}
