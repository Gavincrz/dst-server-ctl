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
	maxLines = normalizeLogLineLimit(maxLines)

	return s.reader.ReadRecent(ctx, filepath.Join(s.layout.Logs, "install-"+string(taskID)+".log"), maxLines)
}

func (s *InstallTaskLogService) Stream(taskID domain.TaskID, maxLines int) (domain.LogStream, error) {
	if taskID == "" {
		return nil, fmt.Errorf("%w: task id is required", domain.ErrTaskNotFound)
	}

	return s.reader.OpenStream(filepath.Join(s.layout.Logs, "install-"+string(taskID)+".log"), normalizeLogLineLimit(maxLines))
}
