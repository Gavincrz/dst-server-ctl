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

const updateCheckInterval = 6 * time.Hour

type UpdateVersionReader interface {
	LocalVersion(ctx context.Context, layout domain.ManagedLayout) (string, error)
	RemoteVersion(ctx context.Context, layout domain.ManagedLayout, logPath string) (string, command.Result, error)
	UpdateDST(ctx context.Context, layout domain.ManagedLayout, logPath string) (command.Result, error)
}

type UpdateRuntimeController interface {
	Status(ctx context.Context) (domain.RuntimeStatus, error)
	Stop(ctx context.Context) error
}

type UpdateStartOptions struct {
	AllowStop bool
}

type UpdateService struct {
	layout   domain.ManagedLayout
	installs InstallationStateRepository
	updates  UpdateStateRepository
	tasks    TaskRepository
	ids      TaskIDGenerator
	reader   UpdateVersionReader
	runtime  UpdateRuntimeController
	now      func() time.Time
	dispatch func(func())

	mu sync.Mutex
}

func NewUpdateService(
	layout domain.ManagedLayout,
	installs InstallationStateRepository,
	updates UpdateStateRepository,
	tasks TaskRepository,
	ids TaskIDGenerator,
	reader UpdateVersionReader,
	runtime UpdateRuntimeController,
) *UpdateService {
	return &UpdateService{
		layout:   layout,
		installs: installs,
		updates:  updates,
		tasks:    tasks,
		ids:      ids,
		reader:   reader,
		runtime:  runtime,
		now:      time.Now,
		dispatch: func(fn func()) {
			go fn()
		},
	}
}

func (s *UpdateService) Initialize(ctx context.Context) (domain.UpdateState, error) {
	state, err := s.updates.GetUpdateState(ctx)
	if err == nil {
		return state, nil
	}
	if !errors.Is(err, domain.ErrUpdateStateNotFound) {
		return domain.UpdateState{}, err
	}

	now := s.now().UTC()
	state = domain.UpdateState{
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.updates.SaveUpdateState(ctx, state); err != nil {
		return domain.UpdateState{}, err
	}

	return state, nil
}

func (s *UpdateService) Status(ctx context.Context) (domain.UpdateState, error) {
	return s.updates.GetUpdateState(ctx)
}

func (s *UpdateService) ListTasks(ctx context.Context) ([]domain.Task, error) {
	allTasks, err := s.tasks.ListTasks(ctx)
	if err != nil {
		return nil, err
	}

	filtered := make([]domain.Task, 0, len(allTasks))
	for _, task := range allTasks {
		if isUpdateTask(task.Type) {
			filtered = append(filtered, task)
		}
	}
	return filtered, nil
}

func (s *UpdateService) CheckNow(ctx context.Context) (domain.UpdateState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.checkNowLocked(ctx)
}

func (s *UpdateService) Start(ctx context.Context, options UpdateStartOptions) (domain.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ensureDSTInstalled(ctx); err != nil {
		return domain.Task{}, err
	}

	state, err := s.updates.GetUpdateState(ctx)
	if err != nil {
		return domain.Task{}, err
	}
	if s.hasActiveUpdateTaskLocked(ctx) {
		return domain.Task{}, domain.ErrUpdateAlreadyInProgress
	}
	if !state.UpdateAvailable {
		return domain.Task{}, domain.ErrUpdateNotRequired
	}
	if err := s.ensureUpdateCanStopRuntime(ctx, options); err != nil {
		return domain.Task{}, err
	}

	task, err := s.createTask(ctx, domain.TaskTypeUpdateDST, "Update Don't Starve Together dedicated server app 343050")
	if err != nil {
		return domain.Task{}, err
	}

	s.dispatch(func() {
		s.executeTask(context.Background(), state, task)
	})

	return task, nil
}

func (s *UpdateService) ensureUpdateCanStopRuntime(ctx context.Context, options UpdateStartOptions) error {
	if s.runtime == nil {
		return nil
	}

	status, err := s.runtime.Status(ctx)
	if err != nil {
		return err
	}
	if status.Status != domain.ServerStatusRunning {
		return nil
	}
	if !options.AllowStop {
		return domain.ErrUpdateRequiresServerStop
	}
	if err := s.runtime.Stop(ctx); err != nil && !errors.Is(err, domain.ErrServerNotRunning) {
		return fmt.Errorf("stop runtime for update: %w", err)
	}

	return nil
}

func (s *UpdateService) StartPeriodicChecks(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = updateCheckInterval
	}

	s.dispatch(func() {
		s.runScheduledCheck(ctx)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.runScheduledCheck(ctx)
			}
		}
	})
}

func (s *UpdateService) runScheduledCheck(ctx context.Context) {
	checkCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	_, err := s.CheckNow(checkCtx)
	switch {
	case err == nil:
		return
	case errors.Is(err, domain.ErrDSTNotInstalled):
		return
	case errors.Is(err, domain.ErrUpdateAlreadyInProgress):
		return
	case errors.Is(err, context.Canceled):
		return
	case errors.Is(err, context.DeadlineExceeded):
		return
	default:
		return
	}
}

func (s *UpdateService) checkNowLocked(ctx context.Context) (domain.UpdateState, error) {
	if err := s.ensureDSTInstalled(ctx); err != nil {
		return domain.UpdateState{}, err
	}
	if s.hasActiveUpdateTaskLocked(ctx) {
		return domain.UpdateState{}, domain.ErrUpdateAlreadyInProgress
	}

	state, err := s.updates.GetUpdateState(ctx)
	if err != nil {
		return domain.UpdateState{}, err
	}

	currentVersion, err := s.reader.LocalVersion(ctx, s.layout)
	if err != nil {
		return domain.UpdateState{}, s.saveCheckFailure(ctx, state, err)
	}

	latestVersion, result, err := s.reader.RemoteVersion(ctx, s.layout, updateCheckLogPath(s.layout))
	if err != nil {
		return domain.UpdateState{}, s.saveCheckFailure(ctx, state, updateCommandError("check_dst_update", result, err))
	}

	now := s.now().UTC()
	state.CurrentVersion = currentVersion
	state.LatestVersion = latestVersion
	state.UpdateAvailable = currentVersion != "" && latestVersion != "" && currentVersion != latestVersion
	state.LastCheckedAt = &now
	state.LastError = ""
	state.UpdatedAt = now

	if err := s.updates.SaveUpdateState(ctx, state); err != nil {
		return domain.UpdateState{}, fmt.Errorf("save update state: %w", err)
	}

	return state, nil
}

func (s *UpdateService) executeTask(ctx context.Context, state domain.UpdateState, task domain.Task) {
	runningTask := task
	if err := s.markTaskRunning(ctx, &runningTask); err != nil {
		_ = s.markTaskFailed(ctx, &runningTask, err)
		return
	}

	result, err := s.reader.UpdateDST(ctx, s.layout, updateTaskLogPath(s.layout, task.ID))
	if err != nil {
		_ = s.markTaskFailed(ctx, &runningTask, updateCommandError(string(task.Type), result, err))
		return
	}

	currentVersion, err := s.reader.LocalVersion(ctx, s.layout)
	if err != nil {
		_ = s.markTaskFailed(ctx, &runningTask, err)
		return
	}

	now := s.now().UTC()
	state.CurrentVersion = currentVersion
	state.LatestVersion = currentVersion
	state.UpdateAvailable = false
	state.LastCheckedAt = &now
	state.LastUpdatedAt = &now
	state.LastError = ""
	state.UpdatedAt = now

	if err := s.updates.SaveUpdateState(ctx, state); err != nil {
		_ = s.markTaskFailed(ctx, &runningTask, fmt.Errorf("save update state: %w", err))
		return
	}

	if err := s.markTaskSucceeded(ctx, &runningTask); err != nil {
		_ = s.markTaskFailed(ctx, &runningTask, err)
	}
}

func (s *UpdateService) ensureDSTInstalled(ctx context.Context) error {
	state, err := s.installs.GetInstallationState(ctx)
	if err != nil {
		return err
	}
	if state.DSTInstalledAt == nil {
		return domain.ErrDSTNotInstalled
	}
	return nil
}

func (s *UpdateService) hasActiveUpdateTaskLocked(ctx context.Context) bool {
	tasks, err := s.tasks.ListTasks(ctx)
	if err != nil {
		return false
	}
	for _, task := range tasks {
		if !isUpdateTask(task.Type) {
			continue
		}
		if task.Status == domain.TaskStatusPending || task.Status == domain.TaskStatusRunning {
			return true
		}
	}
	return false
}

func (s *UpdateService) createTask(ctx context.Context, taskType domain.TaskType, detail string) (domain.Task, error) {
	now := s.now().UTC()
	id, err := s.ids.NewTaskID()
	if err != nil {
		return domain.Task{}, fmt.Errorf("create update task id: %w", err)
	}
	if id == "" {
		return domain.Task{}, fmt.Errorf("create update task: empty task id")
	}

	task := domain.Task{
		ID:        id,
		Type:      taskType,
		Status:    domain.TaskStatusPending,
		Detail:    detail,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.tasks.CreateTask(ctx, task); err != nil {
		return domain.Task{}, err
	}
	return task, nil
}

func (s *UpdateService) markTaskRunning(ctx context.Context, task *domain.Task) error {
	now := s.now().UTC()
	task.Status = domain.TaskStatusRunning
	task.Error = ""
	task.StartedAt = &now
	task.FinishedAt = nil
	task.UpdatedAt = now
	return s.tasks.UpdateTask(ctx, *task)
}

func (s *UpdateService) markTaskSucceeded(ctx context.Context, task *domain.Task) error {
	now := s.now().UTC()
	task.Status = domain.TaskStatusSucceeded
	task.Error = ""
	task.FinishedAt = &now
	task.UpdatedAt = now
	return s.tasks.UpdateTask(ctx, *task)
}

func (s *UpdateService) markTaskFailed(ctx context.Context, task *domain.Task, reason error) error {
	now := s.now().UTC()
	task.Status = domain.TaskStatusFailed
	task.Error = reason.Error()
	task.FinishedAt = &now
	task.UpdatedAt = now
	return s.tasks.UpdateTask(ctx, *task)
}

func (s *UpdateService) saveCheckFailure(ctx context.Context, state domain.UpdateState, reason error) error {
	now := s.now().UTC()
	state.LastCheckedAt = &now
	state.LastError = reason.Error()
	state.UpdatedAt = now
	if err := s.updates.SaveUpdateState(ctx, state); err != nil {
		return fmt.Errorf("save update state after failure: %w", err)
	}
	return reason
}

func isUpdateTask(taskType domain.TaskType) bool {
	return taskType == domain.TaskTypeUpdateCheckDST || taskType == domain.TaskTypeUpdateDST
}

func updateTaskLogPath(layout domain.ManagedLayout, taskID domain.TaskID) string {
	return filepath.Join(layout.Logs, "update-"+string(taskID)+".log")
}

func updateCheckLogPath(layout domain.ManagedLayout) string {
	return filepath.Join(layout.Logs, "update-check.log")
}

func updateCommandError(label string, result command.Result, err error) error {
	parts := []string{fmt.Sprintf("%s failed", label)}
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
