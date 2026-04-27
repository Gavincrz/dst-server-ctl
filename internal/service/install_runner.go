package service

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"dst-server-ctl/internal/adapter/command"
	"dst-server-ctl/internal/domain"
)

type InstallPlannerReader interface {
	Plan(state domain.InstallationState) domain.InstallPlan
}

type InstallCommandRunner interface {
	InstallSteamCMD(ctx context.Context, layout domain.ManagedLayout, logPath string) (command.Result, error)
	InstallDST(ctx context.Context, layout domain.ManagedLayout, logPath string) (command.Result, error)
}

type InstallRunnerService struct {
	layout      domain.ManagedLayout
	installs    InstallationStateRepository
	tasks       TaskRepository
	planner     InstallPlannerReader
	taskService *InstallTaskService
	runner      InstallCommandRunner
	now         func() time.Time
	dispatch    func(func())

	mu sync.Mutex
}

func NewInstallRunnerService(
	layout domain.ManagedLayout,
	installs InstallationStateRepository,
	tasks TaskRepository,
	planner InstallPlannerReader,
	taskService *InstallTaskService,
	runner InstallCommandRunner,
) *InstallRunnerService {
	return &InstallRunnerService{
		layout:      layout,
		installs:    installs,
		tasks:       tasks,
		planner:     planner,
		taskService: taskService,
		runner:      runner,
		now:         time.Now,
		dispatch: func(fn func()) {
			go fn()
		},
	}
}

func (s *InstallRunnerService) ListTasks(ctx context.Context) ([]domain.Task, error) {
	return s.tasks.ListTasks(ctx)
}

func (s *InstallRunnerService) Start(ctx context.Context) ([]domain.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, err := s.tasks.ListTasks(ctx)
	if err != nil {
		return nil, err
	}
	for _, task := range existing {
		if !isInstallTask(task.Type) {
			continue
		}
		if task.Status == domain.TaskStatusPending || task.Status == domain.TaskStatusRunning {
			return nil, domain.ErrInstallAlreadyInProgress
		}
	}

	state, err := s.installs.GetInstallationState(ctx)
	if err != nil {
		return nil, err
	}

	plan := s.planner.Plan(state)
	if len(plan.Steps) == 0 {
		return nil, domain.ErrInstallNotRequired
	}

	tasks, err := s.taskService.CreateTasks(ctx, plan)
	if err != nil {
		return nil, err
	}

	runTasks := append([]domain.Task(nil), tasks...)
	s.dispatch(func() {
		s.execute(context.Background(), state, runTasks)
	})

	return tasks, nil
}

func (s *InstallRunnerService) execute(ctx context.Context, state domain.InstallationState, tasks []domain.Task) {
	for _, task := range tasks {
		runningTask := task
		if err := s.markTaskRunning(ctx, &runningTask); err != nil {
			_ = s.markTaskFailed(ctx, &runningTask, err)
			return
		}

		if err := s.executeTask(ctx, &state, &runningTask); err != nil {
			_ = s.markTaskFailed(ctx, &runningTask, err)
			return
		}

		if err := s.markTaskSucceeded(ctx, &runningTask); err != nil {
			_ = s.markTaskFailed(ctx, &runningTask, err)
			return
		}
	}
}

func (s *InstallRunnerService) executeTask(ctx context.Context, state *domain.InstallationState, task *domain.Task) error {
	var (
		result command.Result
		err    error
	)

	switch task.Type {
	case domain.TaskTypeInstallSteamCMD:
		result, err = s.runner.InstallSteamCMD(ctx, s.layout, installTaskLogPath(s.layout, task.ID))
	case domain.TaskTypeInstallDST:
		result, err = s.runner.InstallDST(ctx, s.layout, installTaskLogPath(s.layout, task.ID))
	default:
		return fmt.Errorf("unsupported install task type %q", task.Type)
	}
	if err != nil {
		return installCommandError(task.Type, result, err)
	}

	finishedAt := s.now().UTC()
	switch task.Type {
	case domain.TaskTypeInstallSteamCMD:
		state.SteamCMDInstalledAt = &finishedAt
	case domain.TaskTypeInstallDST:
		state.DSTInstalledAt = &finishedAt
	}
	state.UpdatedAt = finishedAt

	if err := s.installs.SaveInstallationState(ctx, *state); err != nil {
		return fmt.Errorf("save installation state: %w", err)
	}

	return nil
}

func installTaskLogPath(layout domain.ManagedLayout, taskID domain.TaskID) string {
	return filepath.Join(layout.Logs, "install-"+string(taskID)+".log")
}

func (s *InstallRunnerService) markTaskRunning(ctx context.Context, task *domain.Task) error {
	now := s.now().UTC()
	task.Status = domain.TaskStatusRunning
	task.Error = ""
	task.StartedAt = &now
	task.FinishedAt = nil
	task.UpdatedAt = now
	return s.tasks.UpdateTask(ctx, *task)
}

func (s *InstallRunnerService) markTaskSucceeded(ctx context.Context, task *domain.Task) error {
	now := s.now().UTC()
	task.Status = domain.TaskStatusSucceeded
	task.Error = ""
	task.FinishedAt = &now
	task.UpdatedAt = now
	return s.tasks.UpdateTask(ctx, *task)
}

func (s *InstallRunnerService) markTaskFailed(ctx context.Context, task *domain.Task, reason error) error {
	now := s.now().UTC()
	task.Status = domain.TaskStatusFailed
	task.Error = reason.Error()
	task.FinishedAt = &now
	task.UpdatedAt = now
	return s.tasks.UpdateTask(ctx, *task)
}

func isInstallTask(taskType domain.TaskType) bool {
	return taskType == domain.TaskTypeInstallSteamCMD || taskType == domain.TaskTypeInstallDST
}

func installCommandError(taskType domain.TaskType, result command.Result, err error) error {
	parts := []string{fmt.Sprintf("%s failed", taskType)}

	if stderr := strings.TrimSpace(result.Stderr); stderr != "" {
		parts = append(parts, stderr)
	}
	if stdout := strings.TrimSpace(result.Stdout); stdout != "" {
		parts = append(parts, stdout)
	}

	message := strings.Join(parts, ": ")
	if message == "" {
		return err
	}

	return errors.Join(errors.New(message), err)
}
