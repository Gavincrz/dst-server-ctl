package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"dst-server-ctl/internal/domain"
)

func TestOpenMigratesDatabase(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t, ctx)
	defer store.Close()

	var count int
	if err := store.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations`).Scan(&count); err != nil {
		t.Fatalf("count migrations error = %v", err)
	}
	if count != len(migrations) {
		t.Fatalf("migration count = %d, want %d", count, len(migrations))
	}
}

func TestOpenConfiguresSQLiteForControllerWorkload(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t, ctx)
	defer store.Close()

	stats := store.db.Stats()
	if stats.MaxOpenConnections != 1 {
		t.Fatalf("MaxOpenConnections = %d, want 1", stats.MaxOpenConnections)
	}

	var journalMode string
	if err := store.db.QueryRowContext(ctx, `PRAGMA journal_mode`).Scan(&journalMode); err != nil {
		t.Fatalf("PRAGMA journal_mode error = %v", err)
	}
	if journalMode != "wal" {
		t.Fatalf("journal_mode = %q, want wal", journalMode)
	}

	var busyTimeout int
	if err := store.db.QueryRowContext(ctx, `PRAGMA busy_timeout`).Scan(&busyTimeout); err != nil {
		t.Fatalf("PRAGMA busy_timeout error = %v", err)
	}
	if busyTimeout != 5000 {
		t.Fatalf("busy_timeout = %d, want 5000", busyTimeout)
	}
}

func TestListTasksWaitsForWriterInsteadOfReturningBusy(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "state.db")
	store, err := Open(ctx, dbPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer store.Close()

	task := domain.Task{
		ID:        "task-1",
		Type:      domain.TaskTypeInstallDST,
		Status:    domain.TaskStatusPending,
		Detail:    "Install DST",
		CreatedAt: time.Date(2026, 4, 25, 1, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 4, 25, 1, 0, 0, 0, time.UTC),
	}
	if err := store.CreateTask(ctx, task); err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	locker, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("sql.Open(locker) error = %v", err)
	}
	defer locker.Close()

	if _, err := locker.ExecContext(ctx, `PRAGMA busy_timeout = 5000`); err != nil {
		t.Fatalf("set locker busy_timeout error = %v", err)
	}

	tx, err := locker.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("locker.BeginTx() error = %v", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `UPDATE tasks SET detail = detail WHERE id = ?`, string(task.ID)); err != nil {
		t.Fatalf("lock task row error = %v", err)
	}

	done := make(chan error, 1)
	go func() {
		_, err := store.ListTasks(ctx)
		done <- err
	}()

	time.Sleep(200 * time.Millisecond)
	if err := tx.Commit(); err != nil {
		t.Fatalf("locker.Commit() error = %v", err)
	}

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("ListTasks() error = %v, want nil", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("ListTasks() did not complete after writer released lock")
	}
}

func TestInstallationStateRepositoryRoundTripsState(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t, ctx)
	defer store.Close()

	steamCMDInstalledAt := time.Date(2026, 4, 23, 8, 0, 0, 0, time.UTC)
	state := domain.InstallationState{
		ManagedRoot:         "/srv/dst-server-ctl",
		SteamCMDInstalledAt: &steamCMDInstalledAt,
		CreatedAt:           time.Date(2026, 4, 23, 7, 0, 0, 0, time.UTC),
		UpdatedAt:           time.Date(2026, 4, 23, 8, 30, 0, 0, time.UTC),
	}

	if err := store.SaveInstallationState(ctx, state); err != nil {
		t.Fatalf("SaveInstallationState() error = %v", err)
	}

	got, err := store.GetInstallationState(ctx)
	if err != nil {
		t.Fatalf("GetInstallationState() error = %v", err)
	}

	if got.ManagedRoot != state.ManagedRoot {
		t.Fatalf("ManagedRoot = %q, want %q", got.ManagedRoot, state.ManagedRoot)
	}
	if got.SteamCMDInstalledAt == nil || !got.SteamCMDInstalledAt.Equal(steamCMDInstalledAt) {
		t.Fatalf("SteamCMDInstalledAt = %v, want %v", got.SteamCMDInstalledAt, steamCMDInstalledAt)
	}
	if got.DSTInstalledAt != nil {
		t.Fatalf("DSTInstalledAt = %v, want nil", got.DSTInstalledAt)
	}
	if !got.CreatedAt.Equal(state.CreatedAt) {
		t.Fatalf("CreatedAt = %v, want %v", got.CreatedAt, state.CreatedAt)
	}
	if !got.UpdatedAt.Equal(state.UpdatedAt) {
		t.Fatalf("UpdatedAt = %v, want %v", got.UpdatedAt, state.UpdatedAt)
	}
}

func TestGetInstallationStateReturnsNotFound(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t, ctx)
	defer store.Close()

	_, err := store.GetInstallationState(ctx)
	if !errors.Is(err, domain.ErrInstallationStateNotFound) {
		t.Fatalf("GetInstallationState() error = %v, want ErrInstallationStateNotFound", err)
	}
}

func TestUpdateStateRepositoryRoundTripsState(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t, ctx)
	defer store.Close()

	checkedAt := time.Date(2026, 4, 27, 5, 0, 0, 0, time.UTC)
	updatedAt := checkedAt.Add(time.Hour)
	state := domain.UpdateState{
		CurrentVersion:  "100",
		LatestVersion:   "101",
		UpdateAvailable: true,
		LastCheckedAt:   &checkedAt,
		LastUpdatedAt:   &updatedAt,
		LastError:       "",
		CreatedAt:       checkedAt.Add(-time.Hour),
		UpdatedAt:       updatedAt,
	}

	if err := store.SaveUpdateState(ctx, state); err != nil {
		t.Fatalf("SaveUpdateState() error = %v", err)
	}

	got, err := store.GetUpdateState(ctx)
	if err != nil {
		t.Fatalf("GetUpdateState() error = %v", err)
	}
	if got.CurrentVersion != state.CurrentVersion || got.LatestVersion != state.LatestVersion {
		t.Fatalf("versions = %#v, want %#v", got, state)
	}
	if !got.UpdateAvailable {
		t.Fatal("UpdateAvailable = false, want true")
	}
	if got.LastCheckedAt == nil || !got.LastCheckedAt.Equal(checkedAt) {
		t.Fatalf("LastCheckedAt = %v, want %v", got.LastCheckedAt, checkedAt)
	}
}

func TestGetUpdateStateReturnsNotFound(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t, ctx)
	defer store.Close()

	_, err := store.GetUpdateState(ctx)
	if !errors.Is(err, domain.ErrUpdateStateNotFound) {
		t.Fatalf("GetUpdateState() error = %v, want ErrUpdateStateNotFound", err)
	}
}

func TestClusterConfigRepositoryRoundTripsConfig(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t, ctx)
	defer store.Close()

	config := domain.ClusterConfig{
		ClusterName:        "Managed DST",
		ClusterDescription: "Test cluster",
		GameMode:           "survival",
		MaxPlayers:         8,
		Language:           "en",
		PVP:                true,
		PauseWhenEmpty:     false,
		Shards: []domain.ShardConfig{
			{Name: domain.ShardMaster, Enabled: true},
			{Name: domain.ShardCaves, Enabled: false},
		},
		CreatedAt: time.Date(2026, 4, 24, 9, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 4, 24, 10, 0, 0, 0, time.UTC),
	}

	if err := store.SaveClusterConfig(ctx, config); err != nil {
		t.Fatalf("SaveClusterConfig() error = %v", err)
	}

	got, err := store.GetClusterConfig(ctx)
	if err != nil {
		t.Fatalf("GetClusterConfig() error = %v", err)
	}

	if got.ClusterName != config.ClusterName {
		t.Fatalf("ClusterName = %q, want %q", got.ClusterName, config.ClusterName)
	}
	if got.GameMode != config.GameMode {
		t.Fatalf("GameMode = %q, want %q", got.GameMode, config.GameMode)
	}
	if len(got.Shards) != 2 {
		t.Fatalf("shard count = %d, want 2", len(got.Shards))
	}
	if got.Shards[1].Name != domain.ShardCaves || got.Shards[1].Enabled {
		t.Fatalf("Caves shard = %#v, want disabled caves shard", got.Shards[1])
	}
}

func TestGetClusterConfigReturnsNotFound(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t, ctx)
	defer store.Close()

	_, err := store.GetClusterConfig(ctx)
	if !errors.Is(err, domain.ErrClusterConfigNotFound) {
		t.Fatalf("GetClusterConfig() error = %v, want ErrClusterConfigNotFound", err)
	}
}

func TestTaskRepositoryCreatesListsGetsAndUpdatesTasks(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t, ctx)
	defer store.Close()

	createdAt := time.Date(2026, 4, 23, 9, 0, 0, 0, time.UTC)
	task := domain.Task{
		ID:        "task-1",
		Type:      domain.TaskTypeInstallDST,
		Status:    domain.TaskStatusPending,
		Detail:    "Install DST",
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	}

	if err := store.CreateTask(ctx, task); err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	got, err := store.GetTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("GetTask() error = %v", err)
	}
	if got.ID != task.ID {
		t.Fatalf("ID = %q, want %q", got.ID, task.ID)
	}
	if got.Status != domain.TaskStatusPending {
		t.Fatalf("Status = %q, want pending", got.Status)
	}

	startedAt := createdAt.Add(time.Minute)
	finishedAt := createdAt.Add(2 * time.Minute)
	got.Status = domain.TaskStatusSucceeded
	got.StartedAt = &startedAt
	got.FinishedAt = &finishedAt
	got.UpdatedAt = finishedAt
	if err := store.UpdateTask(ctx, got); err != nil {
		t.Fatalf("UpdateTask() error = %v", err)
	}

	tasks, err := store.ListTasks(ctx)
	if err != nil {
		t.Fatalf("ListTasks() error = %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("task count = %d, want 1", len(tasks))
	}
	if tasks[0].Status != domain.TaskStatusSucceeded {
		t.Fatalf("Status = %q, want succeeded", tasks[0].Status)
	}
	if tasks[0].StartedAt == nil || !tasks[0].StartedAt.Equal(startedAt) {
		t.Fatalf("StartedAt = %v, want %v", tasks[0].StartedAt, startedAt)
	}
}

func TestGetTaskReturnsNotFound(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t, ctx)
	defer store.Close()

	_, err := store.GetTask(ctx, "missing")
	if !errors.Is(err, domain.ErrTaskNotFound) {
		t.Fatalf("GetTask() error = %v, want ErrTaskNotFound", err)
	}
}

func TestRuntimeEventRepositoryCreatesAndListsEvents(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t, ctx)
	defer store.Close()

	first := domain.RuntimeEvent{
		Shard:     domain.ShardMaster,
		Kind:      domain.RuntimeEventStarted,
		Detail:    "Master shard started",
		CreatedAt: time.Date(2026, 4, 25, 2, 0, 0, 0, time.UTC),
	}
	second := domain.RuntimeEvent{
		Shard:     domain.ShardMaster,
		Kind:      domain.RuntimeEventExited,
		Detail:    "Master shard exited",
		CreatedAt: time.Date(2026, 4, 25, 2, 5, 0, 0, time.UTC),
	}

	if err := store.CreateRuntimeEvent(ctx, first); err != nil {
		t.Fatalf("CreateRuntimeEvent(first) error = %v", err)
	}
	if err := store.CreateRuntimeEvent(ctx, second); err != nil {
		t.Fatalf("CreateRuntimeEvent(second) error = %v", err)
	}

	events, err := store.ListRuntimeEvents(ctx, 10)
	if err != nil {
		t.Fatalf("ListRuntimeEvents() error = %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("event count = %d, want 2", len(events))
	}
	if events[0].Kind != domain.RuntimeEventExited {
		t.Fatalf("first kind = %q, want exited", events[0].Kind)
	}
	if events[1].Kind != domain.RuntimeEventStarted {
		t.Fatalf("second kind = %q, want started", events[1].Kind)
	}
}

func openTestStore(t *testing.T, ctx context.Context) *Store {
	t.Helper()

	store, err := Open(ctx, filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	return store
}
