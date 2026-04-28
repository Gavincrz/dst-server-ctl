package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"dst-server-ctl/internal/adapter/command"
	"dst-server-ctl/internal/domain"
)

func TestUpdateServiceInitializeCreatesMissingState(t *testing.T) {
	ctx := context.Background()
	repo := &fakeUpdateStateRepository{err: domain.ErrUpdateStateNotFound}
	service := NewUpdateService(domain.ManagedLayout{}, &fakeInstallationStateRepository{}, repo, &fakeTaskRepository{}, &fakeTaskIDGenerator{}, &fakeUpdateVersionReader{}, &fakeUpdateRuntimeController{})
	now := time.Date(2026, 4, 27, 1, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return now }

	state, err := service.Initialize(ctx)
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}
	if !state.CreatedAt.Equal(now) {
		t.Fatalf("CreatedAt = %v, want %v", state.CreatedAt, now)
	}
	if repo.saved == nil {
		t.Fatal("saved update state = nil, want populated state")
	}
}

func TestUpdateServiceCheckNowUpdatesVersionState(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 4, 27, 2, 0, 0, 0, time.UTC)
	installs := &fakeInstallationStateRepository{
		state: domain.InstallationState{DSTInstalledAt: ptrTime(now.Add(-time.Hour))},
	}
	updates := &fakeUpdateStateRepository{
		state: domain.UpdateState{
			CreatedAt: now.Add(-2 * time.Hour),
			UpdatedAt: now.Add(-2 * time.Hour),
		},
	}
	service := NewUpdateService(domain.ManagedLayout{}, installs, updates, &fakeTaskRepository{}, &fakeTaskIDGenerator{}, &fakeUpdateVersionReader{
		localVersion:  "100",
		remoteVersion: "101",
	}, &fakeUpdateRuntimeController{})
	service.now = func() time.Time { return now }

	state, err := service.CheckNow(ctx)
	if err != nil {
		t.Fatalf("CheckNow() error = %v", err)
	}
	if state.CurrentVersion != "100" {
		t.Fatalf("CurrentVersion = %q, want 100", state.CurrentVersion)
	}
	if state.LatestVersion != "101" {
		t.Fatalf("LatestVersion = %q, want 101", state.LatestVersion)
	}
	if !state.UpdateAvailable {
		t.Fatal("UpdateAvailable = false, want true")
	}
	if updates.saved == nil || updates.saved.LastCheckedAt == nil {
		t.Fatal("saved LastCheckedAt = nil, want populated")
	}
}

func TestUpdateServiceStartCreatesAndExecutesUpdateTask(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 4, 27, 3, 0, 0, 0, time.UTC)
	installs := &fakeInstallationStateRepository{
		state: domain.InstallationState{DSTInstalledAt: ptrTime(now.Add(-time.Hour))},
	}
	updates := &fakeUpdateStateRepository{
		state: domain.UpdateState{
			CurrentVersion:  "100",
			LatestVersion:   "101",
			UpdateAvailable: true,
			CreatedAt:       now.Add(-2 * time.Hour),
			UpdatedAt:       now.Add(-time.Hour),
		},
	}
	tasks := &fakeTaskRepository{}
	reader := &fakeUpdateVersionReader{localVersion: "101"}
	runtime := &fakeUpdateRuntimeController{}
	service := NewUpdateService(
		domain.ManagedLayout{Logs: "/srv/managed/logs"},
		installs,
		updates,
		tasks,
		&fakeTaskIDGenerator{ids: []domain.TaskID{"task-1"}},
		reader,
		runtime,
	)
	service.dispatch = func(fn func()) { fn() }
	service.now = func() time.Time {
		now = now.Add(time.Minute)
		return now
	}

	task, err := service.Start(ctx, UpdateStartOptions{})
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if task.ID != "task-1" {
		t.Fatalf("task.ID = %q, want task-1", task.ID)
	}

	stored, err := tasks.GetTask(ctx, "task-1")
	if err != nil {
		t.Fatalf("GetTask() error = %v", err)
	}
	if stored.Status != domain.TaskStatusSucceeded {
		t.Fatalf("stored status = %q, want succeeded", stored.Status)
	}
	if reader.updateLogPath != "/srv/managed/logs/update-task-1.log" {
		t.Fatalf("updateLogPath = %q, want /srv/managed/logs/update-task-1.log", reader.updateLogPath)
	}
	if updates.saved == nil || updates.saved.CurrentVersion != "101" {
		t.Fatalf("saved update state = %#v, want current version 101", updates.saved)
	}
	if updates.saved.UpdateAvailable {
		t.Fatal("UpdateAvailable = true, want false after successful update")
	}
	if runtime.stopCalls != 0 {
		t.Fatalf("stopCalls = %d, want 0", runtime.stopCalls)
	}
}

func TestUpdateServiceStartRejectsWhenNoUpdateAvailable(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 4, 27, 4, 0, 0, 0, time.UTC)
	service := NewUpdateService(
		domain.ManagedLayout{},
		&fakeInstallationStateRepository{state: domain.InstallationState{DSTInstalledAt: ptrTime(now)}},
		&fakeUpdateStateRepository{state: domain.UpdateState{CurrentVersion: "100", LatestVersion: "100"}},
		&fakeTaskRepository{},
		&fakeTaskIDGenerator{},
		&fakeUpdateVersionReader{},
		&fakeUpdateRuntimeController{},
	)

	_, err := service.Start(ctx, UpdateStartOptions{})
	if !errors.Is(err, domain.ErrUpdateNotRequired) {
		t.Fatalf("Start() error = %v, want %v", err, domain.ErrUpdateNotRequired)
	}
}

func TestUpdateServiceStartRequiresStopConfirmationWhenRuntimeRunning(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 4, 27, 4, 30, 0, 0, time.UTC)
	runtime := &fakeUpdateRuntimeController{
		status: domain.RuntimeStatus{Status: domain.ServerStatusRunning},
	}
	service := NewUpdateService(
		domain.ManagedLayout{},
		&fakeInstallationStateRepository{state: domain.InstallationState{DSTInstalledAt: ptrTime(now)}},
		&fakeUpdateStateRepository{state: domain.UpdateState{CurrentVersion: "100", LatestVersion: "101", UpdateAvailable: true}},
		&fakeTaskRepository{},
		&fakeTaskIDGenerator{},
		&fakeUpdateVersionReader{},
		runtime,
	)

	_, err := service.Start(ctx, UpdateStartOptions{})
	if !errors.Is(err, domain.ErrUpdateRequiresServerStop) {
		t.Fatalf("Start() error = %v, want %v", err, domain.ErrUpdateRequiresServerStop)
	}
	if runtime.stopCalls != 0 {
		t.Fatalf("stopCalls = %d, want 0", runtime.stopCalls)
	}
}

func TestUpdateServiceStartStopsRuntimeAfterConfirmation(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 4, 27, 5, 0, 0, 0, time.UTC)
	updates := &fakeUpdateStateRepository{
		state: domain.UpdateState{
			CurrentVersion:  "100",
			LatestVersion:   "101",
			UpdateAvailable: true,
			CreatedAt:       now.Add(-2 * time.Hour),
			UpdatedAt:       now.Add(-time.Hour),
		},
	}
	reader := &fakeUpdateVersionReader{localVersion: "101"}
	runtime := &fakeUpdateRuntimeController{
		status: domain.RuntimeStatus{Status: domain.ServerStatusRunning},
	}
	service := NewUpdateService(
		domain.ManagedLayout{Logs: "/srv/managed/logs"},
		&fakeInstallationStateRepository{state: domain.InstallationState{DSTInstalledAt: ptrTime(now.Add(-time.Hour))}},
		updates,
		&fakeTaskRepository{},
		&fakeTaskIDGenerator{ids: []domain.TaskID{"task-2"}},
		reader,
		runtime,
	)
	service.dispatch = func(fn func()) { fn() }
	service.now = func() time.Time {
		now = now.Add(time.Minute)
		return now
	}

	task, err := service.Start(ctx, UpdateStartOptions{AllowStop: true})
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if task.ID != "task-2" {
		t.Fatalf("task.ID = %q, want task-2", task.ID)
	}
	if runtime.stopCalls != 1 {
		t.Fatalf("stopCalls = %d, want 1", runtime.stopCalls)
	}
	if updates.saved == nil || updates.saved.LastUpdatedAt == nil {
		t.Fatalf("saved update state = %#v, want persisted update result", updates.saved)
	}
}

type fakeUpdateStateRepository struct {
	state domain.UpdateState
	err   error
	saved *domain.UpdateState
}

func (r *fakeUpdateStateRepository) GetUpdateState(context.Context) (domain.UpdateState, error) {
	if r.err != nil {
		return domain.UpdateState{}, r.err
	}
	return r.state, nil
}

func (r *fakeUpdateStateRepository) SaveUpdateState(_ context.Context, state domain.UpdateState) error {
	r.saved = &state
	r.state = state
	r.err = nil
	return nil
}

type fakeUpdateVersionReader struct {
	localVersion  string
	localErr      error
	remoteVersion string
	remoteResult  command.Result
	remoteErr     error
	updateResult  command.Result
	updateErr     error
	updateLogPath string
}

type fakeUpdateRuntimeController struct {
	status    domain.RuntimeStatus
	statusErr error
	stopErr   error
	stopCalls int
}

func (r *fakeUpdateRuntimeController) Status(context.Context) (domain.RuntimeStatus, error) {
	if r.statusErr != nil {
		return domain.RuntimeStatus{}, r.statusErr
	}
	return r.status, nil
}

func (r *fakeUpdateRuntimeController) Stop(context.Context) error {
	r.stopCalls++
	return r.stopErr
}

func (r *fakeUpdateVersionReader) LocalVersion(context.Context, domain.ManagedLayout) (string, error) {
	if r.localErr != nil {
		return "", r.localErr
	}
	return r.localVersion, nil
}

func (r *fakeUpdateVersionReader) RemoteVersion(context.Context, domain.ManagedLayout) (string, command.Result, error) {
	if r.remoteErr != nil {
		return "", r.remoteResult, r.remoteErr
	}
	return r.remoteVersion, r.remoteResult, nil
}

func (r *fakeUpdateVersionReader) UpdateDST(_ context.Context, _ domain.ManagedLayout, logPath string) (command.Result, error) {
	r.updateLogPath = logPath
	if r.updateErr != nil {
		return r.updateResult, r.updateErr
	}
	return r.updateResult, nil
}

func ptrTime(value time.Time) *time.Time {
	return &value
}
