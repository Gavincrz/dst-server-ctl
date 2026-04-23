package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"dst-server-ctl/internal/domain"
)

func TestInstallationServiceInitializeCreatesMissingState(t *testing.T) {
	ctx := context.Background()
	repo := &fakeInstallationStateRepository{err: domain.ErrInstallationStateNotFound}
	service := NewInstallationService(domain.ManagedLayout{Root: "/srv/dst-server-ctl"}, repo)
	now := time.Date(2026, 4, 23, 9, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return now }

	state, err := service.Initialize(ctx)
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	if state.ManagedRoot != "/srv/dst-server-ctl" {
		t.Fatalf("ManagedRoot = %q, want /srv/dst-server-ctl", state.ManagedRoot)
	}
	if !state.CreatedAt.Equal(now) {
		t.Fatalf("CreatedAt = %v, want %v", state.CreatedAt, now)
	}
	if repo.saved == nil {
		t.Fatal("expected Initialize() to save state")
	}
}

func TestInstallationServiceInitializeReturnsExistingState(t *testing.T) {
	ctx := context.Background()
	existing := domain.InstallationState{
		ManagedRoot: "/existing",
		CreatedAt:   time.Date(2026, 4, 23, 8, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 4, 23, 8, 0, 0, 0, time.UTC),
	}
	repo := &fakeInstallationStateRepository{state: existing}
	service := NewInstallationService(domain.ManagedLayout{Root: "/new"}, repo)

	state, err := service.Initialize(ctx)
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	if state.ManagedRoot != existing.ManagedRoot {
		t.Fatalf("ManagedRoot = %q, want %q", state.ManagedRoot, existing.ManagedRoot)
	}
	if repo.saved != nil {
		t.Fatal("expected Initialize() not to overwrite existing state")
	}
}

func TestInstallationServiceInitializeReturnsRepositoryErrors(t *testing.T) {
	ctx := context.Background()
	wantErr := errors.New("database unavailable")
	repo := &fakeInstallationStateRepository{err: wantErr}
	service := NewInstallationService(domain.ManagedLayout{Root: "/srv/dst-server-ctl"}, repo)

	_, err := service.Initialize(ctx)
	if !errors.Is(err, wantErr) {
		t.Fatalf("Initialize() error = %v, want %v", err, wantErr)
	}
}

type fakeInstallationStateRepository struct {
	state domain.InstallationState
	err   error
	saved *domain.InstallationState
}

func (r *fakeInstallationStateRepository) GetInstallationState(context.Context) (domain.InstallationState, error) {
	if r.err != nil {
		return domain.InstallationState{}, r.err
	}
	return r.state, nil
}

func (r *fakeInstallationStateRepository) SaveInstallationState(_ context.Context, state domain.InstallationState) error {
	r.saved = &state
	r.state = state
	r.err = nil
	return nil
}
