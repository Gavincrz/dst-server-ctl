package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"dst-server-ctl/internal/domain"
)

func TestInstallTaskServiceCreatesPendingTasksFromPlan(t *testing.T) {
	ctx := context.Background()
	repo := &fakeTaskRepository{}
	service := NewInstallTaskService(repo, &fakeTaskIDGenerator{ids: []domain.TaskID{"task-1", "task-2"}})
	now := time.Date(2026, 4, 23, 9, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return now }

	tasks, err := service.CreateTasks(ctx, domain.InstallPlan{
		Steps: []domain.InstallStep{
			{Type: domain.TaskTypeInstallSteamCMD, Description: "Install SteamCMD"},
			{Type: domain.TaskTypeInstallDST, Description: "Install DST"},
		},
	})
	if err != nil {
		t.Fatalf("CreateTasks() error = %v", err)
	}

	if len(tasks) != 2 {
		t.Fatalf("task count = %d, want 2", len(tasks))
	}
	if tasks[0].ID != "task-1" {
		t.Fatalf("first ID = %q, want task-1", tasks[0].ID)
	}
	if tasks[0].Status != domain.TaskStatusPending {
		t.Fatalf("first status = %q, want pending", tasks[0].Status)
	}
	if !tasks[0].CreatedAt.Equal(now) {
		t.Fatalf("CreatedAt = %v, want %v", tasks[0].CreatedAt, now)
	}
	if len(repo.tasks) != 2 {
		t.Fatalf("saved task count = %d, want 2", len(repo.tasks))
	}
}

func TestInstallTaskServiceReturnsTaskIDErrors(t *testing.T) {
	ctx := context.Background()
	wantErr := errors.New("random source failed")
	service := NewInstallTaskService(&fakeTaskRepository{}, &fakeTaskIDGenerator{err: wantErr})

	_, err := service.CreateTasks(ctx, domain.InstallPlan{
		Steps: []domain.InstallStep{{Type: domain.TaskTypeInstallDST, Description: "Install DST"}},
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("CreateTasks() error = %v, want %v", err, wantErr)
	}
}

type fakeTaskIDGenerator struct {
	ids []domain.TaskID
	err error
}

func (g *fakeTaskIDGenerator) NewTaskID() (domain.TaskID, error) {
	if g.err != nil {
		return "", g.err
	}
	if len(g.ids) == 0 {
		return "", nil
	}
	id := g.ids[0]
	g.ids = g.ids[1:]
	return id, nil
}

type fakeTaskRepository struct {
	tasks []domain.Task
}

func (r *fakeTaskRepository) CreateTask(_ context.Context, task domain.Task) error {
	r.tasks = append(r.tasks, task)
	return nil
}

func (r *fakeTaskRepository) GetTask(_ context.Context, id domain.TaskID) (domain.Task, error) {
	for _, task := range r.tasks {
		if task.ID == id {
			return task, nil
		}
	}
	return domain.Task{}, domain.ErrTaskNotFound
}

func (r *fakeTaskRepository) ListTasks(context.Context) ([]domain.Task, error) {
	return r.tasks, nil
}

func (r *fakeTaskRepository) UpdateTask(_ context.Context, task domain.Task) error {
	for i := range r.tasks {
		if r.tasks[i].ID == task.ID {
			r.tasks[i] = task
			return nil
		}
	}
	return domain.ErrTaskNotFound
}
