package service

import (
	"context"
	"fmt"
	"path/filepath"

	"dst-server-ctl/internal/domain"
)

type InstallTaskLogService struct {
	layout domain.ManagedLayout
	reader RuntimeLogReader
}

func NewInstallTaskLogService(layout domain.ManagedLayout, reader RuntimeLogReader) *InstallTaskLogService {
	return &InstallTaskLogService{
		layout: layout,
		reader: reader,
	}
}

func (s *InstallTaskLogService) Get(ctx context.Context, taskID domain.TaskID, maxLines int) ([]string, error) {
	if taskID == "" {
		return nil, fmt.Errorf("%w: task id is required", domain.ErrTaskNotFound)
	}
	if maxLines <= 0 {
		maxLines = 200
	}
	if maxLines > 500 {
		maxLines = 500
	}

	return s.reader.ReadRecent(ctx, filepath.Join(s.layout.Logs, "install-"+string(taskID)+".log"), maxLines)
}
