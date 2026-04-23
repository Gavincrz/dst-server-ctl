package sqlite

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"dst-server-ctl/internal/domain"
)

func TestOpenMigratesDatabase(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t, ctx)
	defer store.Close()

	var count int
	if err := store.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations`).Scan(&count); err != nil {
		t.Fatalf("count migrations error = %v", err)
	}
	if count != len(migrations) {
		t.Fatalf("migration count = %d, want %d", count, len(migrations))
	}
}

func TestInstallationStateRepositoryRoundTripsState(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t, ctx)
	defer store.Close()

	steamCMDInstalledAt := time.Date(2026, 4, 23, 8, 0, 0, 0, time.UTC)
	state := domain.InstallationState{
		ManagedRoot:         "/srv/dst-server-ctl",
		SteamCMDInstalledAt: &steamCMDInstalledAt,
		CreatedAt:           time.Date(2026, 4, 23, 7, 0, 0, 0, time.UTC),
		UpdatedAt:           time.Date(2026, 4, 23, 8, 30, 0, 0, time.UTC),
	}

	if err := store.SaveInstallationState(ctx, state); err != nil {
		t.Fatalf("SaveInstallationState() error = %v", err)
	}

	got, err := store.GetInstallationState(ctx)
	if err != nil {
		t.Fatalf("GetInstallationState() error = %v", err)
	}

	if got.ManagedRoot != state.ManagedRoot {
		t.Fatalf("ManagedRoot = %q, want %q", got.ManagedRoot, state.ManagedRoot)
	}
	if got.SteamCMDInstalledAt == nil || !got.SteamCMDInstalledAt.Equal(steamCMDInstalledAt) {
		t.Fatalf("SteamCMDInstalledAt = %v, want %v", got.SteamCMDInstalledAt, steamCMDInstalledAt)
	}
	if got.DSTInstalledAt != nil {
		t.Fatalf("DSTInstalledAt = %v, want nil", got.DSTInstalledAt)
	}
	if !got.CreatedAt.Equal(state.CreatedAt) {
		t.Fatalf("CreatedAt = %v, want %v", got.CreatedAt, state.CreatedAt)
	}
	if !got.UpdatedAt.Equal(state.UpdatedAt) {
		t.Fatalf("UpdatedAt = %v, want %v", got.UpdatedAt, state.UpdatedAt)
	}
}

func TestGetInstallationStateReturnsNotFound(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t, ctx)
	defer store.Close()

	_, err := store.GetInstallationState(ctx)
	if !errors.Is(err, domain.ErrInstallationStateNotFound) {
		t.Fatalf("GetInstallationState() error = %v, want ErrInstallationStateNotFound", err)
	}
}

func openTestStore(t *testing.T, ctx context.Context) *Store {
	t.Helper()

	store, err := Open(ctx, filepath.Join(t.TempDir(), "state.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	return store
}
