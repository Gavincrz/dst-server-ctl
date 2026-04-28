package service

import (
	"context"
	"path/filepath"

	"dst-server-ctl/internal/domain"
)

type UpdateCheckLogService struct {
	layout domain.ManagedLayout
	reader RuntimeLogReader
}

func NewUpdateCheckLogService(layout domain.ManagedLayout, reader RuntimeLogReader) *UpdateCheckLogService {
	return &UpdateCheckLogService{
		layout: layout,
		reader: reader,
	}
}

func (s *UpdateCheckLogService) Get(ctx context.Context, maxLines int) ([]string, error) {
	if maxLines <= 0 {
		maxLines = 200
	}
	if maxLines > 500 {
		maxLines = 500
	}

	return s.reader.ReadRecent(ctx, filepath.Join(s.layout.Logs, "update-check.log"), maxLines)
}
