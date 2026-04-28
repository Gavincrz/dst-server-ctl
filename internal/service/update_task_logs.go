package service

import (
	"context"
	"fmt"
	"path/filepath"

	"dst-server-ctl/internal/domain"
)

type UpdateTaskLogService struct {
	layout domain.ManagedLayout
	reader RuntimeLogReader
}

func NewUpdateTaskLogService(layout domain.ManagedLayout, reader RuntimeLogReader) *UpdateTaskLogService {
	return &UpdateTaskLogService{
		layout: layout,
		reader: reader,
	}
}

func (s *UpdateTaskLogService) Get(ctx context.Context, taskID domain.TaskID, maxLines int) ([]string, error) {
	if taskID == "" {
		return nil, fmt.Errorf("%w: task id is required", domain.ErrTaskNotFound)
	}
	if maxLines <= 0 {
		maxLines = 200
	}
	if maxLines > 500 {
		maxLines = 500
	}

	return s.reader.ReadRecent(ctx, filepath.Join(s.layout.Logs, "update-"+string(taskID)+".log"), maxLines)
}
