package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"dst-server-ctl/internal/adapter/command"
	"dst-server-ctl/internal/domain"
)

func TestInstallRunnerServiceStartCreatesAndExecutesTasks(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 4, 24, 8, 0, 0, 0, time.UTC)
	installs := &fakeInstallationStateRepository{
		state: domain.InstallationState{
			ManagedRoot: "/srv/managed",
			CreatedAt:   now.Add(-time.Hour),
			UpdatedAt:   now.Add(-time.Hour),
		},
	}
	tasks := &fakeTaskRepository{}
	taskService := NewInstallTaskService(tasks, &fakeTaskIDGenerator{ids: []domain.TaskID{"task-1", "task-2"}})
	taskService.now = func() time.Time { return now }
	runner := &fakeInstallCommandRunner{}

	service := NewInstallRunnerService(
		domain.ManagedLayout{Root: "/srv/managed", SteamCMD: "/srv/managed/steamcmd", DST: "/srv/managed/dst", Logs: "/srv/managed/logs"},
		installs,
		tasks,
		NewInstallPlanner(),
		taskService,
		runner,
	)
	service.now = func() time.Time {
		now = now.Add(time.Minute)
		return now
	}
	service.dispatch = func(fn func()) { fn() }

	created, err := service.Start(ctx)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	if len(created) != 2 {
		t.Fatalf("created task count = %d, want 2", len(created))
	}

	allTasks, err := service.ListTasks(ctx)
	if err != nil {
		t.Fatalf("ListTasks() error = %v", err)
	}
	if len(allTasks) != 2 {
		t.Fatalf("stored task count = %d, want 2", len(allTasks))
	}
	if allTasks[0].Status != domain.TaskStatusSucceeded {
		t.Fatalf("latest task status = %q, want succeeded", allTasks[0].Status)
	}

	if installs.saved == nil {
		t.Fatal("saved installation state = nil, want updated state")
	}
	if installs.saved.SteamCMDInstalledAt == nil {
		t.Fatal("SteamCMDInstalledAt = nil, want populated")
	}
	if installs.saved.DSTInstalledAt == nil {
		t.Fatal("DSTInstalledAt = nil, want populated")
	}
	if runner.steamCMDLogPath != "/srv/managed/logs/install-task-1.log" {
		t.Fatalf("steamCMDLogPath = %q, want /srv/managed/logs/install-task-1.log", runner.steamCMDLogPath)
	}
	if runner.dstLogPath != "/srv/managed/logs/install-task-2.log" {
		t.Fatalf("dstLogPath = %q, want /srv/managed/logs/install-task-2.log", runner.dstLogPath)
	}
}

func TestInstallRunnerServiceStartRejectsWhenInstallAlreadyRunning(t *testing.T) {
	ctx := context.Background()
	service := NewInstallRunnerService(
		domain.ManagedLayout{},
		&fakeInstallationStateRepository{state: domain.InstallationState{ManagedRoot: "/srv/managed"}},
		&fakeTaskRepository{
			tasks: []domain.Task{
				{ID: "task-1", Type: domain.TaskTypeInstallDST, Status: domain.TaskStatusRunning},
			},
		},
		NewInstallPlanner(),
		NewInstallTaskService(&fakeTaskRepository{}, &fakeTaskIDGenerator{}),
		&fakeInstallCommandRunner{},
	)

	_, err := service.Start(ctx)
	if !errors.Is(err, domain.ErrInstallAlreadyInProgress) {
		t.Fatalf("Start() error = %v, want %v", err, domain.ErrInstallAlreadyInProgress)
	}
}

func TestInstallRunnerServiceStartRejectsWhenInstallNotRequired(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 4, 24, 8, 0, 0, 0, time.UTC)
	taskRepo := &fakeTaskRepository{}
	service := NewInstallRunnerService(
		domain.ManagedLayout{},
		&fakeInstallationStateRepository{
			state: domain.InstallationState{
				ManagedRoot:         "/srv/managed",
				SteamCMDInstalledAt: &now,
				DSTInstalledAt:      &now,
				CreatedAt:           now,
				UpdatedAt:           now,
			},
		},
		taskRepo,
		NewInstallPlanner(),
		NewInstallTaskService(taskRepo, &fakeTaskIDGenerator{}),
		&fakeInstallCommandRunner{},
	)

	_, err := service.Start(ctx)
	if !errors.Is(err, domain.ErrInstallNotRequired) {
		t.Fatalf("Start() error = %v, want %v", err, domain.ErrInstallNotRequired)
	}
}

func TestInstallRunnerServiceMarksTaskFailedWhenCommandFails(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 4, 24, 8, 0, 0, 0, time.UTC)
	installs := &fakeInstallationStateRepository{
		state: domain.InstallationState{
			ManagedRoot: "/srv/managed",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}
	taskRepo := &fakeTaskRepository{}
	taskService := NewInstallTaskService(taskRepo, &fakeTaskIDGenerator{ids: []domain.TaskID{"task-1", "task-2"}})
	taskService.now = func() time.Time { return now }

	service := NewInstallRunnerService(
		domain.ManagedLayout{},
		installs,
		taskRepo,
		NewInstallPlanner(),
		taskService,
		&fakeInstallCommandRunner{
			installSteamCMDErr:    errors.New("curl failed"),
			installSteamCMDResult: command.Result{Stderr: "network down"},
		},
	)
	service.now = func() time.Time {
		now = now.Add(time.Minute)
		return now
	}
	service.dispatch = func(fn func()) { fn() }

	_, err := service.Start(ctx)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	allTasks, err := service.ListTasks(ctx)
	if err != nil {
		t.Fatalf("ListTasks() error = %v", err)
	}
	if len(allTasks) != 2 {
		t.Fatalf("stored task count = %d, want 2", len(allTasks))
	}
	if allTasks[0].Status != domain.TaskStatusFailed {
		t.Fatalf("first created task status = %q, want failed", allTasks[0].Status)
	}
	if allTasks[1].Status != domain.TaskStatusPending {
		t.Fatalf("second created task status = %q, want pending", allTasks[1].Status)
	}
	if allTasks[0].Error == "" {
		t.Fatal("failed task error = empty, want message")
	}
	if installs.saved != nil {
		t.Fatalf("saved installation state = %#v, want nil", installs.saved)
	}
}

type fakeInstallCommandRunner struct {
	installSteamCMDResult command.Result
	installSteamCMDErr    error
	installDSTResult      command.Result
	installDSTErr         error
	steamCMDLogPath       string
	dstLogPath            string
}

func (r *fakeInstallCommandRunner) InstallSteamCMD(_ context.Context, _ domain.ManagedLayout, logPath string) (command.Result, error) {
	r.steamCMDLogPath = logPath
	return r.installSteamCMDResult, r.installSteamCMDErr
}

func (r *fakeInstallCommandRunner) InstallDST(_ context.Context, _ domain.ManagedLayout, logPath string) (command.Result, error) {
	r.dstLogPath = logPath
	return r.installDSTResult, r.installDSTErr
}
