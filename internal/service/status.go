package service

import "dst-server-ctl/internal/domain"

type StatusService struct {
	version string
}

func NewStatusService(version string) *StatusService {
	return &StatusService{version: version}
}

func (s *StatusService) Status() domain.Status {
	return domain.Status{
		Version: s.version,
		Status:  domain.ServerStatusUnknown,
	}
}
