package service

import (
	"context"

	"dst-server-ctl/internal/domain"
)

type InstallationStateRepository interface {
	GetInstallationState(ctx context.Context) (domain.InstallationState, error)
	SaveInstallationState(ctx context.Context, state domain.InstallationState) error
}
