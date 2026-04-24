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

func (s *Store) GetClusterConfig(ctx context.Context) (domain.ClusterConfig, error) {
	var row clusterConfigRow
	err := s.db.QueryRowContext(ctx, `
SELECT cluster_name, cluster_description, game_mode, max_players, language, pvp, pause_when_empty, created_at, updated_at
FROM cluster_config
WHERE id = 1`).Scan(
		&row.ClusterName,
		&row.ClusterDescription,
		&row.GameMode,
		&row.MaxPlayers,
		&row.Language,
		&row.PVP,
		&row.PauseWhenEmpty,
		&row.CreatedAt,
		&row.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ClusterConfig{}, domain.ErrClusterConfigNotFound
	}
	if err != nil {
		return domain.ClusterConfig{}, fmt.Errorf("get cluster config: %w", err)
	}

	shards, err := s.listClusterShards(ctx)
	if err != nil {
		return domain.ClusterConfig{}, err
	}
	row.Shards = shards

	config, err := row.toDomain()
	if err != nil {
		return domain.ClusterConfig{}, err
	}

	return config, nil
}

func (s *Store) SaveClusterConfig(ctx context.Context, config domain.ClusterConfig) error {
	row := clusterConfigRowFromDomain(config)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin save cluster config: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
INSERT INTO cluster_config (
	id,
	cluster_name,
	cluster_description,
	game_mode,
	max_players,
	language,
	pvp,
	pause_when_empty,
	created_at,
	updated_at
) VALUES (1, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
	cluster_name = excluded.cluster_name,
	cluster_description = excluded.cluster_description,
	game_mode = excluded.game_mode,
	max_players = excluded.max_players,
	language = excluded.language,
	pvp = excluded.pvp,
	pause_when_empty = excluded.pause_when_empty,
	updated_at = excluded.updated_at`,
		row.ClusterName,
		row.ClusterDescription,
		row.GameMode,
		row.MaxPlayers,
		row.Language,
		row.PVP,
		row.PauseWhenEmpty,
		row.CreatedAt,
		row.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("save cluster config: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM cluster_shards`); err != nil {
		return fmt.Errorf("clear cluster shards: %w", err)
	}
	for _, shard := range row.Shards {
		if _, err := tx.ExecContext(ctx, `
INSERT INTO cluster_shards (name, enabled)
VALUES (?, ?)`,
			shard.Name,
			shard.Enabled,
		); err != nil {
			return fmt.Errorf("save cluster shard %s: %w", shard.Name, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit cluster config: %w", err)
	}

	return nil
}

func (s *Store) CreateTask(ctx context.Context, task domain.Task) error {
	row := taskRowFromDomain(task)
	_, err := s.db.ExecContext(ctx, `
INSERT INTO tasks (
	id,
	type,
	status,
	detail,
	error,
	started_at,
	finished_at,
	created_at,
	updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		row.ID,
		row.Type,
		row.Status,
		row.Detail,
		row.Error,
		row.StartedAt,
		row.FinishedAt,
		row.CreatedAt,
		row.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create task: %w", err)
	}

	return nil
}

func (s *Store) GetTask(ctx context.Context, id domain.TaskID) (domain.Task, error) {
	row, err := s.getTaskRow(ctx, `SELECT id, type, status, detail, error, started_at, finished_at, created_at, updated_at FROM tasks WHERE id = ?`, string(id))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Task{}, domain.ErrTaskNotFound
	}
	if err != nil {
		return domain.Task{}, err
	}

	return row.toDomain()
}

func (s *Store) ListTasks(ctx context.Context) ([]domain.Task, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, type, status, detail, error, started_at, finished_at, created_at, updated_at FROM tasks ORDER BY created_at DESC, id DESC`)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []domain.Task
	for rows.Next() {
		row, err := scanTaskRow(rows)
		if err != nil {
			return nil, err
		}
		task, err := row.toDomain()
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tasks: %w", err)
	}

	return tasks, nil
}

func (s *Store) UpdateTask(ctx context.Context, task domain.Task) error {
	row := taskRowFromDomain(task)
	result, err := s.db.ExecContext(ctx, `
UPDATE tasks SET
	type = ?,
	status = ?,
	detail = ?,
	error = ?,
	started_at = ?,
	finished_at = ?,
	updated_at = ?
WHERE id = ?`,
		row.Type,
		row.Status,
		row.Detail,
		row.Error,
		row.StartedAt,
		row.FinishedAt,
		row.UpdatedAt,
		row.ID,
	)
	if err != nil {
		return fmt.Errorf("update task: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check updated task count: %w", err)
	}
	if affected == 0 {
		return domain.ErrTaskNotFound
	}

	return nil
}

func (s *Store) getTaskRow(ctx context.Context, query string, args ...any) (taskRow, error) {
	row := s.db.QueryRowContext(ctx, query, args...)
	task, err := scanTaskRow(row)
	if err != nil {
		return taskRow{}, err
	}
	return task, nil
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

type clusterConfigRow struct {
	ClusterName        string
	ClusterDescription string
	GameMode           string
	MaxPlayers         int
	Language           string
	PVP                bool
	PauseWhenEmpty     bool
	Shards             []clusterShardRow
	CreatedAt          string
	UpdatedAt          string
}

type clusterShardRow struct {
	Name    string
	Enabled bool
}

func clusterConfigRowFromDomain(config domain.ClusterConfig) clusterConfigRow {
	shards := make([]clusterShardRow, 0, len(config.Shards))
	for _, shard := range config.Shards {
		shards = append(shards, clusterShardRow{
			Name:    string(shard.Name),
			Enabled: shard.Enabled,
		})
	}

	return clusterConfigRow{
		ClusterName:        config.ClusterName,
		ClusterDescription: config.ClusterDescription,
		GameMode:           config.GameMode,
		MaxPlayers:         config.MaxPlayers,
		Language:           config.Language,
		PVP:                config.PVP,
		PauseWhenEmpty:     config.PauseWhenEmpty,
		Shards:             shards,
		CreatedAt:          config.CreatedAt.UTC().Format(timeFormat),
		UpdatedAt:          config.UpdatedAt.UTC().Format(timeFormat),
	}
}

func (r clusterConfigRow) toDomain() (domain.ClusterConfig, error) {
	createdAt, err := parseRequiredTime("created_at", r.CreatedAt)
	if err != nil {
		return domain.ClusterConfig{}, err
	}
	updatedAt, err := parseRequiredTime("updated_at", r.UpdatedAt)
	if err != nil {
		return domain.ClusterConfig{}, err
	}

	shards := make([]domain.ShardConfig, 0, len(r.Shards))
	for _, shard := range r.Shards {
		shards = append(shards, domain.ShardConfig{
			Name:    domain.ShardName(shard.Name),
			Enabled: shard.Enabled,
		})
	}

	return domain.ClusterConfig{
		ClusterName:        r.ClusterName,
		ClusterDescription: r.ClusterDescription,
		GameMode:           r.GameMode,
		MaxPlayers:         r.MaxPlayers,
		Language:           r.Language,
		PVP:                r.PVP,
		PauseWhenEmpty:     r.PauseWhenEmpty,
		Shards:             shards,
		CreatedAt:          createdAt,
		UpdatedAt:          updatedAt,
	}, nil
}

func (s *Store) listClusterShards(ctx context.Context) ([]clusterShardRow, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT name, enabled FROM cluster_shards ORDER BY CASE name WHEN 'Master' THEN 0 WHEN 'Caves' THEN 1 ELSE 2 END, name`)
	if err != nil {
		return nil, fmt.Errorf("list cluster shards: %w", err)
	}
	defer rows.Close()

	var shards []clusterShardRow
	for rows.Next() {
		var shard clusterShardRow
		if err := rows.Scan(&shard.Name, &shard.Enabled); err != nil {
			return nil, fmt.Errorf("scan cluster shard: %w", err)
		}
		shards = append(shards, shard)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate cluster shards: %w", err)
	}

	return shards, nil
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

type taskRow struct {
	ID         string
	Type       string
	Status     string
	Detail     string
	Error      string
	StartedAt  sql.NullString
	FinishedAt sql.NullString
	CreatedAt  string
	UpdatedAt  string
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanTaskRow(scanner rowScanner) (taskRow, error) {
	var row taskRow
	if err := scanner.Scan(
		&row.ID,
		&row.Type,
		&row.Status,
		&row.Detail,
		&row.Error,
		&row.StartedAt,
		&row.FinishedAt,
		&row.CreatedAt,
		&row.UpdatedAt,
	); err != nil {
		return taskRow{}, err
	}
	return row, nil
}

func taskRowFromDomain(task domain.Task) taskRow {
	return taskRow{
		ID:         string(task.ID),
		Type:       string(task.Type),
		Status:     string(task.Status),
		Detail:     task.Detail,
		Error:      task.Error,
		StartedAt:  nullableTime(task.StartedAt),
		FinishedAt: nullableTime(task.FinishedAt),
		CreatedAt:  task.CreatedAt.UTC().Format(timeFormat),
		UpdatedAt:  task.UpdatedAt.UTC().Format(timeFormat),
	}
}

func (r taskRow) toDomain() (domain.Task, error) {
	startedAt, err := parseNullableTime("started_at", r.StartedAt)
	if err != nil {
		return domain.Task{}, err
	}
	finishedAt, err := parseNullableTime("finished_at", r.FinishedAt)
	if err != nil {
		return domain.Task{}, err
	}
	createdAt, err := parseRequiredTime("created_at", r.CreatedAt)
	if err != nil {
		return domain.Task{}, err
	}
	updatedAt, err := parseRequiredTime("updated_at", r.UpdatedAt)
	if err != nil {
		return domain.Task{}, err
	}

	return domain.Task{
		ID:         domain.TaskID(r.ID),
		Type:       domain.TaskType(r.Type),
		Status:     domain.TaskStatus(r.Status),
		Detail:     r.Detail,
		Error:      r.Error,
		StartedAt:  startedAt,
		FinishedAt: finishedAt,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}, nil
}
