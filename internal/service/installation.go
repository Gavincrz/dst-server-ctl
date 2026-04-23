package service

import (
	"context"
	"errors"
	"time"

	"dst-server-ctl/internal/domain"
)

type InstallationService struct {
	layout domain.ManagedLayout
	repo   InstallationStateRepository
	now    func() time.Time
}

func NewInstallationService(layout domain.ManagedLayout, repo InstallationStateRepository) *InstallationService {
	return &InstallationService{
		layout: layout,
		repo:   repo,
		now:    time.Now,
	}
}

func (s *InstallationService) Initialize(ctx context.Context) (domain.InstallationState, error) {
	state, err := s.repo.GetInstallationState(ctx)
	if err == nil {
		return state, nil
	}
	if !errors.Is(err, domain.ErrInstallationStateNotFound) {
		return domain.InstallationState{}, err
	}

	now := s.now().UTC()
	state = domain.InstallationState{
		ManagedRoot: s.layout.Root,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.repo.SaveInstallationState(ctx, state); err != nil {
		return domain.InstallationState{}, err
	}

	return state, nil
}

func (s *InstallationService) Status(ctx context.Context) (domain.InstallationState, error) {
	return s.repo.GetInstallationState(ctx)
}
