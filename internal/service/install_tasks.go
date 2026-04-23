package service

import (
	"context"
	"fmt"
	"time"

	"dst-server-ctl/internal/domain"
)

type TaskIDGenerator interface {
	NewTaskID() (domain.TaskID, error)
}

type InstallTaskService struct {
	repo TaskRepository
	ids  TaskIDGenerator
	now  func() time.Time
}

func NewInstallTaskService(repo TaskRepository, ids TaskIDGenerator) *InstallTaskService {
	return &InstallTaskService{
		repo: repo,
		ids:  ids,
		now:  time.Now,
	}
}

func (s *InstallTaskService) CreateTasks(ctx context.Context, plan domain.InstallPlan) ([]domain.Task, error) {
	tasks := make([]domain.Task, 0, len(plan.Steps))

	for _, step := range plan.Steps {
		now := s.now().UTC()
		id, err := s.ids.NewTaskID()
		if err != nil {
			return nil, fmt.Errorf("create install task id: %w", err)
		}
		task := domain.Task{
			ID:        id,
			Type:      step.Type,
			Status:    domain.TaskStatusPending,
			Detail:    step.Description,
			CreatedAt: now,
			UpdatedAt: now,
		}

		if task.ID == "" {
			return nil, fmt.Errorf("create install task: empty task id")
		}
		if err := s.repo.CreateTask(ctx, task); err != nil {
			return nil, err
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}
