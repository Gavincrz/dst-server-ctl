package service

import (
	"context"

	"dst-server-ctl/internal/domain"
)

type RuntimeHistoryService struct {
	events RuntimeEventRepository
}

func NewRuntimeHistoryService(events RuntimeEventRepository) *RuntimeHistoryService {
	return &RuntimeHistoryService{events: events}
}

func (s *RuntimeHistoryService) List(ctx context.Context, limit int) ([]domain.RuntimeEvent, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.events.ListRuntimeEvents(ctx, limit)
}
