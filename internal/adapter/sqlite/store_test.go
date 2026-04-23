package sqlite

import (
	"context"
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

func openTestStore(t *testing.T, ctx context.Context) *Store {
	t.Helper()

	store, err := Open(ctx, filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	return store
}
