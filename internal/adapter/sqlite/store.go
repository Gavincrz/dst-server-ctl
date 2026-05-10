package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"maps"
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
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if _, err := db.ExecContext(ctx, `PRAGMA journal_mode = WAL`); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable sqlite wal mode: %w", err)
	}

	if _, err := db.ExecContext(ctx, `PRAGMA busy_timeout = 5000`); err != nil {
		db.Close()
		return nil, fmt.Errorf("set sqlite busy timeout: %w", err)
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

func (s *Store) GetUpdateState(ctx context.Context) (domain.UpdateState, error) {
	var row updateStateRow
	err := s.db.QueryRowContext(ctx, `
SELECT current_version, latest_version, update_available, last_checked_at, last_updated_at, last_error, created_at, updated_at
FROM update_state
WHERE id = 1`).Scan(
		&row.CurrentVersion,
		&row.LatestVersion,
		&row.UpdateAvailable,
		&row.LastCheckedAt,
		&row.LastUpdatedAt,
		&row.LastError,
		&row.CreatedAt,
		&row.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.UpdateState{}, domain.ErrUpdateStateNotFound
	}
	if err != nil {
		return domain.UpdateState{}, fmt.Errorf("get update state: %w", err)
	}

	state, err := row.toDomain()
	if err != nil {
		return domain.UpdateState{}, err
	}

	return state, nil
}

func (s *Store) SaveUpdateState(ctx context.Context, state domain.UpdateState) error {
	row := updateStateRowFromDomain(state)
	_, err := s.db.ExecContext(ctx, `
INSERT INTO update_state (
	id,
	current_version,
	latest_version,
	update_available,
	last_checked_at,
	last_updated_at,
	last_error,
	created_at,
	updated_at
) VALUES (1, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
	current_version = excluded.current_version,
	latest_version = excluded.latest_version,
	update_available = excluded.update_available,
	last_checked_at = excluded.last_checked_at,
	last_updated_at = excluded.last_updated_at,
	last_error = excluded.last_error,
	updated_at = excluded.updated_at`,
		row.CurrentVersion,
		row.LatestVersion,
		row.UpdateAvailable,
		row.LastCheckedAt,
		row.LastUpdatedAt,
		row.LastError,
		row.CreatedAt,
		row.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("save update state: %w", err)
	}

	return nil
}

func (s *Store) GetClusterConfig(ctx context.Context) (domain.ClusterConfig, error) {
	var row clusterConfigRow
	err := s.db.QueryRowContext(ctx, `
SELECT cluster_name, cluster_description, cluster_password, cluster_intention, game_mode, max_players, language, pvp, pause_when_empty, offline_cluster, lan_only_cluster, tick_rate, console_enabled, bind_ip, master_port, cluster_key, created_at, updated_at
FROM cluster_config
WHERE id = 1`).Scan(
		&row.ClusterName,
		&row.ClusterDescription,
		&row.ClusterPassword,
		&row.ClusterIntention,
		&row.GameMode,
		&row.MaxPlayers,
		&row.Language,
		&row.PVP,
		&row.PauseWhenEmpty,
		&row.OfflineCluster,
		&row.LANOnlyCluster,
		&row.TickRate,
		&row.ConsoleEnabled,
		&row.BindIP,
		&row.MasterPort,
		&row.ClusterKey,
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
	cluster_password,
	cluster_intention,
	game_mode,
	max_players,
	language,
	pvp,
	pause_when_empty,
	offline_cluster,
	lan_only_cluster,
	tick_rate,
	console_enabled,
	bind_ip,
	master_port,
	cluster_key,
	created_at,
	updated_at
) VALUES (1, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
	cluster_name = excluded.cluster_name,
	cluster_description = excluded.cluster_description,
	cluster_password = excluded.cluster_password,
	cluster_intention = excluded.cluster_intention,
	game_mode = excluded.game_mode,
	max_players = excluded.max_players,
	language = excluded.language,
	pvp = excluded.pvp,
	pause_when_empty = excluded.pause_when_empty,
	offline_cluster = excluded.offline_cluster,
	lan_only_cluster = excluded.lan_only_cluster,
	tick_rate = excluded.tick_rate,
	console_enabled = excluded.console_enabled,
	bind_ip = excluded.bind_ip,
	master_port = excluded.master_port,
	cluster_key = excluded.cluster_key,
	updated_at = excluded.updated_at`,
		row.ClusterName,
		row.ClusterDescription,
		row.ClusterPassword,
		row.ClusterIntention,
		row.GameMode,
		row.MaxPlayers,
		row.Language,
		row.PVP,
		row.PauseWhenEmpty,
		row.OfflineCluster,
		row.LANOnlyCluster,
		row.TickRate,
		row.ConsoleEnabled,
		row.BindIP,
		row.MasterPort,
		row.ClusterKey,
		row.CreatedAt,
		row.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("save cluster config: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM cluster_shards`); err != nil {
		return fmt.Errorf("clear cluster shards: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM cluster_shard_world_overrides`); err != nil {
		return fmt.Errorf("clear cluster shard world overrides: %w", err)
	}
	for _, shard := range row.Shards {
		if _, err := tx.ExecContext(ctx, `
INSERT INTO cluster_shards (name, enabled, server_port, master_server_port, authentication_port, worldgen_preset)
VALUES (?, ?, ?, ?, ?, ?)`,
			shard.Name,
			shard.Enabled,
			shard.ServerPort,
			shard.MasterServerPort,
			shard.AuthenticationPort,
			shard.WorldGenPreset,
		); err != nil {
			return fmt.Errorf("save cluster shard %s: %w", shard.Name, err)
		}
		for key, value := range shard.WorldGenOverrides {
			if _, err := tx.ExecContext(ctx, `
INSERT INTO cluster_shard_world_overrides (shard_name, override_key, override_value)
VALUES (?, ?, ?)`,
				shard.Name,
				key,
				value,
			); err != nil {
				return fmt.Errorf("save cluster shard %s world override %s: %w", shard.Name, key, err)
			}
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

func (s *Store) CreateRuntimeEvent(ctx context.Context, event domain.RuntimeEvent) error {
	row := runtimeEventRowFromDomain(event)
	_, err := s.db.ExecContext(ctx, `
INSERT INTO runtime_events (
	shard,
	kind,
	detail,
	created_at
) VALUES (?, ?, ?, ?)`,
		row.Shard,
		row.Kind,
		row.Detail,
		row.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create runtime event: %w", err)
	}

	return nil
}

func (s *Store) ListRuntimeEvents(ctx context.Context, limit int) ([]domain.RuntimeEvent, error) {
	if limit <= 0 {
		limit = 50
	}

	rows, err := s.db.QueryContext(ctx, `
SELECT id, shard, kind, detail, created_at
FROM runtime_events
ORDER BY created_at DESC, id DESC
LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("list runtime events: %w", err)
	}
	defer rows.Close()

	var events []domain.RuntimeEvent
	for rows.Next() {
		row, err := scanRuntimeEventRow(rows)
		if err != nil {
			return nil, err
		}
		event, err := row.toDomain()
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate runtime events: %w", err)
	}

	return events, nil
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

type updateStateRow struct {
	CurrentVersion  string
	LatestVersion   string
	UpdateAvailable bool
	LastCheckedAt   sql.NullString
	LastUpdatedAt   sql.NullString
	LastError       string
	CreatedAt       string
	UpdatedAt       string
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

func updateStateRowFromDomain(state domain.UpdateState) updateStateRow {
	return updateStateRow{
		CurrentVersion:  state.CurrentVersion,
		LatestVersion:   state.LatestVersion,
		UpdateAvailable: state.UpdateAvailable,
		LastCheckedAt:   nullableTime(state.LastCheckedAt),
		LastUpdatedAt:   nullableTime(state.LastUpdatedAt),
		LastError:       state.LastError,
		CreatedAt:       state.CreatedAt.UTC().Format(timeFormat),
		UpdatedAt:       state.UpdatedAt.UTC().Format(timeFormat),
	}
}

func (r updateStateRow) toDomain() (domain.UpdateState, error) {
	createdAt, err := parseRequiredTime("created_at", r.CreatedAt)
	if err != nil {
		return domain.UpdateState{}, err
	}
	updatedAt, err := parseRequiredTime("updated_at", r.UpdatedAt)
	if err != nil {
		return domain.UpdateState{}, err
	}
	lastCheckedAt, err := parseNullableTime("last_checked_at", r.LastCheckedAt)
	if err != nil {
		return domain.UpdateState{}, err
	}
	lastUpdatedAt, err := parseNullableTime("last_updated_at", r.LastUpdatedAt)
	if err != nil {
		return domain.UpdateState{}, err
	}

	return domain.UpdateState{
		CurrentVersion:  r.CurrentVersion,
		LatestVersion:   r.LatestVersion,
		UpdateAvailable: r.UpdateAvailable,
		LastCheckedAt:   lastCheckedAt,
		LastUpdatedAt:   lastUpdatedAt,
		LastError:       r.LastError,
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
	}, nil
}

type clusterConfigRow struct {
	ClusterName        string
	ClusterDescription string
	ClusterPassword    string
	ClusterIntention   string
	GameMode           string
	MaxPlayers         int
	Language           string
	PVP                bool
	PauseWhenEmpty     bool
	OfflineCluster     bool
	LANOnlyCluster     bool
	TickRate           int
	ConsoleEnabled     bool
	BindIP             string
	MasterPort         int
	ClusterKey         string
	Shards             []clusterShardRow
	CreatedAt          string
	UpdatedAt          string
}

type clusterShardRow struct {
	Name               string
	Enabled            bool
	ServerPort         int
	MasterServerPort   int
	AuthenticationPort int
	WorldGenPreset     string
	WorldGenOverrides  map[string]string
}

func clusterConfigRowFromDomain(config domain.ClusterConfig) clusterConfigRow {
	shards := make([]clusterShardRow, 0, len(config.Shards))
	for _, shard := range config.Shards {
		shards = append(shards, clusterShardRow{
			Name:               string(shard.Name),
			Enabled:            shard.Enabled,
			ServerPort:         shard.ServerPort,
			MasterServerPort:   shard.MasterServerPort,
			AuthenticationPort: shard.AuthenticationPort,
			WorldGenPreset:     shard.WorldGenPreset,
			WorldGenOverrides:  maps.Clone(shard.WorldGenOverrides),
		})
	}

	return clusterConfigRow{
		ClusterName:        config.ClusterName,
		ClusterDescription: config.ClusterDescription,
		ClusterPassword:    config.ClusterPassword,
		ClusterIntention:   config.ClusterIntention,
		GameMode:           config.GameMode,
		MaxPlayers:         config.MaxPlayers,
		Language:           config.Language,
		PVP:                config.PVP,
		PauseWhenEmpty:     config.PauseWhenEmpty,
		OfflineCluster:     config.OfflineCluster,
		LANOnlyCluster:     config.LANOnlyCluster,
		TickRate:           config.TickRate,
		ConsoleEnabled:     config.ConsoleEnabled,
		BindIP:             config.BindIP,
		MasterPort:         config.MasterPort,
		ClusterKey:         config.ClusterKey,
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
			Name:               domain.ShardName(shard.Name),
			Enabled:            shard.Enabled,
			ServerPort:         shard.ServerPort,
			MasterServerPort:   shard.MasterServerPort,
			AuthenticationPort: shard.AuthenticationPort,
			WorldGenPreset:     shard.WorldGenPreset,
			WorldGenOverrides:  maps.Clone(shard.WorldGenOverrides),
		})
	}

	return domain.ClusterConfig{
		ClusterName:        r.ClusterName,
		ClusterDescription: r.ClusterDescription,
		ClusterPassword:    r.ClusterPassword,
		ClusterIntention:   r.ClusterIntention,
		GameMode:           r.GameMode,
		MaxPlayers:         r.MaxPlayers,
		Language:           r.Language,
		PVP:                r.PVP,
		PauseWhenEmpty:     r.PauseWhenEmpty,
		OfflineCluster:     r.OfflineCluster,
		LANOnlyCluster:     r.LANOnlyCluster,
		TickRate:           r.TickRate,
		ConsoleEnabled:     r.ConsoleEnabled,
		BindIP:             r.BindIP,
		MasterPort:         r.MasterPort,
		ClusterKey:         r.ClusterKey,
		Shards:             shards,
		CreatedAt:          createdAt,
		UpdatedAt:          updatedAt,
	}, nil
}

func (s *Store) listClusterShards(ctx context.Context) ([]clusterShardRow, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT name, enabled, server_port, master_server_port, authentication_port, worldgen_preset FROM cluster_shards ORDER BY CASE name WHEN 'Master' THEN 0 WHEN 'Caves' THEN 1 ELSE 2 END, name`)
	if err != nil {
		return nil, fmt.Errorf("list cluster shards: %w", err)
	}
	defer rows.Close()

	var shards []clusterShardRow
	for rows.Next() {
		var shard clusterShardRow
		if err := rows.Scan(&shard.Name, &shard.Enabled, &shard.ServerPort, &shard.MasterServerPort, &shard.AuthenticationPort, &shard.WorldGenPreset); err != nil {
			return nil, fmt.Errorf("scan cluster shard: %w", err)
		}
		shards = append(shards, shard)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate cluster shards: %w", err)
	}
	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("close cluster shard rows: %w", err)
	}

	for i := range shards {
		overrides, err := s.listWorldOverridesByShard(ctx, shards[i].Name)
		if err != nil {
			return nil, err
		}
		shards[i].WorldGenOverrides = overrides
	}

	return shards, nil
}

func (s *Store) listWorldOverridesByShard(ctx context.Context, shardName string) (map[string]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT override_key, override_value FROM cluster_shard_world_overrides WHERE shard_name = ? ORDER BY override_key`, shardName)
	if err != nil {
		return nil, fmt.Errorf("list cluster shard world overrides for %s: %w", shardName, err)
	}
	defer rows.Close()

	overrides := map[string]string{}
	for rows.Next() {
		var key string
		var value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf("scan cluster shard world override for %s: %w", shardName, err)
		}
		overrides[key] = value
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate cluster shard world overrides for %s: %w", shardName, err)
	}

	return overrides, nil
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

type runtimeEventRow struct {
	ID        int64
	Shard     string
	Kind      string
	Detail    string
	CreatedAt string
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

func scanRuntimeEventRow(scanner rowScanner) (runtimeEventRow, error) {
	var row runtimeEventRow
	if err := scanner.Scan(
		&row.ID,
		&row.Shard,
		&row.Kind,
		&row.Detail,
		&row.CreatedAt,
	); err != nil {
		return runtimeEventRow{}, err
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

func runtimeEventRowFromDomain(event domain.RuntimeEvent) runtimeEventRow {
	return runtimeEventRow{
		ID:        event.ID,
		Shard:     string(event.Shard),
		Kind:      string(event.Kind),
		Detail:    event.Detail,
		CreatedAt: event.CreatedAt.UTC().Format(timeFormat),
	}
}

func (r runtimeEventRow) toDomain() (domain.RuntimeEvent, error) {
	createdAt, err := parseRequiredTime("created_at", r.CreatedAt)
	if err != nil {
		return domain.RuntimeEvent{}, err
	}

	return domain.RuntimeEvent{
		ID:        r.ID,
		Shard:     domain.ShardName(r.Shard),
		Kind:      domain.RuntimeEventKind(r.Kind),
		Detail:    r.Detail,
		CreatedAt: createdAt,
	}, nil
}
