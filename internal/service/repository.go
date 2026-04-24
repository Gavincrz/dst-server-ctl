package service

import (
	"context"

	"dst-server-ctl/internal/domain"
)

type InstallationStateRepository interface {
	GetInstallationState(ctx context.Context) (domain.InstallationState, error)
	SaveInstallationState(ctx context.Context, state domain.InstallationState) error
}

type ClusterConfigRepository interface {
	GetClusterConfig(ctx context.Context) (domain.ClusterConfig, error)
	SaveClusterConfig(ctx context.Context, config domain.ClusterConfig) error
}

type TaskRepository interface {
	CreateTask(ctx context.Context, task domain.Task) error
	GetTask(ctx context.Context, id domain.TaskID) (domain.Task, error)
	ListTasks(ctx context.Context) ([]domain.Task, error)
	UpdateTask(ctx context.Context, task domain.Task) error
}
