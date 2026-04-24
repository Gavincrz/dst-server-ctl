package service

import (
	"time"

	"dst-server-ctl/internal/domain"
)

type StatusService struct {
	version   string
	startedAt time.Time
	now       func() time.Time
}

func NewStatusService(version string) *StatusService {
	now := time.Now().UTC()
	return &StatusService{
		version:   version,
		startedAt: now,
		now:       time.Now,
	}
}

func (s *StatusService) Status() domain.Status {
	return domain.Status{
		Version:   s.version,
		Status:    domain.ServerStatusRunning,
		StartedAt: &s.startedAt,
	}
}
