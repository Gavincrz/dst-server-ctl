package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"dst-server-ctl/internal/domain"
)

func TestClusterConfigServiceInitializeCreatesMissingConfig(t *testing.T) {
	ctx := context.Background()
	repo := &fakeClusterConfigRepository{err: domain.ErrClusterConfigNotFound}
	writer := &fakeClusterFilesWriter{}
	service := NewClusterConfigService(repo, writer)
	now := time.Date(2026, 4, 24, 9, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return now }

	config, err := service.Initialize(ctx)
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	if config.ClusterName != "DST Server" {
		t.Fatalf("ClusterName = %q, want %q", config.ClusterName, "DST Server")
	}
	if len(config.Shards) != 2 {
		t.Fatalf("shard count = %d, want 2", len(config.Shards))
	}
	if repo.saved == nil {
		t.Fatal("expected Initialize() to save config")
	}
	if writer.written == nil {
		t.Fatal("expected Initialize() to write cluster files")
	}
}

func TestClusterConfigServiceInitializeReturnsExistingConfig(t *testing.T) {
	ctx := context.Background()
	existing := domain.ClusterConfig{
		ClusterName:      "Existing",
		ClusterIntention: "cooperative",
		GameMode:         "endless",
		MaxPlayers:       8,
		Language:         "en",
		TickRate:         15,
		Shards: []domain.ShardConfig{
			{Name: domain.ShardMaster, Enabled: true, ServerPort: 10999, MasterServerPort: 27016, AuthenticationPort: 8766, WorldGenPreset: "SURVIVAL_TOGETHER", WorldGenOverrides: map[string]string{}},
		},
		CreatedAt: time.Date(2026, 4, 24, 8, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 4, 24, 8, 0, 0, 0, time.UTC),
	}
	repo := &fakeClusterConfigRepository{config: existing}
	service := NewClusterConfigService(repo, nil)

	config, err := service.Initialize(ctx)
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	if config.ClusterName != existing.ClusterName {
		t.Fatalf("ClusterName = %q, want %q", config.ClusterName, existing.ClusterName)
	}
	if repo.saved != nil {
		t.Fatal("expected Initialize() not to overwrite existing config")
	}
}

func TestClusterConfigServiceUpdateNormalizesAndPersists(t *testing.T) {
	ctx := context.Background()
	createdAt := time.Date(2026, 4, 24, 8, 0, 0, 0, time.UTC)
	repo := &fakeClusterConfigRepository{
		config: domain.ClusterConfig{
			ClusterName:      "Old",
			ClusterIntention: "cooperative",
			GameMode:         "survival",
			MaxPlayers:       6,
			Language:         "en",
			TickRate:         15,
			Shards: []domain.ShardConfig{
				{Name: domain.ShardMaster, Enabled: true, ServerPort: 10999, MasterServerPort: 27016, AuthenticationPort: 8766, WorldGenPreset: "SURVIVAL_TOGETHER", WorldGenOverrides: map[string]string{}},
				{Name: domain.ShardCaves, Enabled: true, ServerPort: 11000, MasterServerPort: 27017, AuthenticationPort: 8767, WorldGenPreset: "DST_CAVE", WorldGenOverrides: map[string]string{}},
			},
			CreatedAt: createdAt,
			UpdatedAt: createdAt,
		},
	}
	writer := &fakeClusterFilesWriter{}
	service := NewClusterConfigService(repo, writer)
	updatedAt := createdAt.Add(time.Hour)
	service.now = func() time.Time { return updatedAt }

	config, err := service.Update(ctx, domain.ClusterConfig{
		ClusterName:        "  New Cluster  ",
		ClusterDescription: "  test  ",
		ClusterPassword:    "  secret  ",
		ClusterIntention:   " social ",
		GameMode:           " endless ",
		MaxPlayers:         12,
		Language:           " en ",
		PVP:                true,
		PauseWhenEmpty:     false,
		OfflineCluster:     true,
		LANOnlyCluster:     false,
		TickRate:           30,
		ConsoleEnabled:     true,
		BindIP:             " 0.0.0.0 ",
		MasterPort:         12000,
		ClusterKey:         " cluster-1 ",
		Shards: []domain.ShardConfig{
			{Name: domain.ShardCaves, Enabled: false, ServerPort: 11001, MasterServerPort: 27018, AuthenticationPort: 8768, WorldGenPreset: "DST_CAVE_PLUS", WorldGenOverrides: map[string]string{"wormattacks": "never"}},
			{Name: domain.ShardMaster, Enabled: true, ServerPort: 11000, MasterServerPort: 27017, AuthenticationPort: 8767, WorldGenPreset: "SURVIVAL_TOGETHER_CLASSIC", WorldGenOverrides: map[string]string{"season_start": "autumn"}},
		},
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	if config.ClusterName != "New Cluster" {
		t.Fatalf("ClusterName = %q, want %q", config.ClusterName, "New Cluster")
	}
	if config.ClusterDescription != "test" {
		t.Fatalf("ClusterDescription = %q, want %q", config.ClusterDescription, "test")
	}
	if !config.CreatedAt.Equal(createdAt) {
		t.Fatalf("CreatedAt = %v, want %v", config.CreatedAt, createdAt)
	}
	if !config.UpdatedAt.Equal(updatedAt) {
		t.Fatalf("UpdatedAt = %v, want %v", config.UpdatedAt, updatedAt)
	}
	if len(config.Shards) != 2 || config.Shards[0].Name != domain.ShardMaster {
		t.Fatalf("Shards = %#v, want Master first", config.Shards)
	}
	if writer.written == nil || writer.written.ClusterName != "New Cluster" {
		t.Fatalf("writer config = %#v, want updated cluster config", writer.written)
	}
}

func TestClusterConfigServiceUpdateValidatesConfig(t *testing.T) {
	ctx := context.Background()
	repo := &fakeClusterConfigRepository{
		config: domain.ClusterConfig{
			ClusterName:      "Existing",
			ClusterIntention: "cooperative",
			GameMode:         "survival",
			MaxPlayers:       6,
			Language:         "en",
			TickRate:         15,
			Shards: []domain.ShardConfig{
				{Name: domain.ShardMaster, Enabled: true, ServerPort: 10999, MasterServerPort: 27016, AuthenticationPort: 8766, WorldGenPreset: "SURVIVAL_TOGETHER", WorldGenOverrides: map[string]string{}},
			},
			CreatedAt: time.Date(2026, 4, 24, 8, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2026, 4, 24, 8, 0, 0, 0, time.UTC),
		},
	}
	service := NewClusterConfigService(repo, nil)

	_, err := service.Update(ctx, domain.ClusterConfig{
		ClusterName:      "",
		ClusterIntention: "cooperative",
		GameMode:         "survival",
		MaxPlayers:       6,
		Language:         "en",
		TickRate:         15,
		Shards: []domain.ShardConfig{
			{Name: domain.ShardMaster, Enabled: true, ServerPort: 10999, MasterServerPort: 27016, AuthenticationPort: 8766, WorldGenPreset: "SURVIVAL_TOGETHER", WorldGenOverrides: map[string]string{}},
		},
	})
	if err == nil {
		t.Fatal("Update() error = nil, want validation error")
	}
}

func TestClusterConfigServiceInitializeReturnsRepositoryErrors(t *testing.T) {
	ctx := context.Background()
	wantErr := errors.New("database unavailable")
	repo := &fakeClusterConfigRepository{err: wantErr}
	service := NewClusterConfigService(repo, nil)

	_, err := service.Initialize(ctx)
	if !errors.Is(err, wantErr) {
		t.Fatalf("Initialize() error = %v, want %v", err, wantErr)
	}
}

type fakeClusterConfigRepository struct {
	config domain.ClusterConfig
	err    error
	saved  *domain.ClusterConfig
}

type fakeClusterFilesWriter struct {
	written *domain.ClusterConfig
	err     error
}

func (r *fakeClusterConfigRepository) GetClusterConfig(context.Context) (domain.ClusterConfig, error) {
	if r.err != nil {
		return domain.ClusterConfig{}, r.err
	}
	return r.config, nil
}

func (r *fakeClusterConfigRepository) SaveClusterConfig(_ context.Context, config domain.ClusterConfig) error {
	r.saved = &config
	r.config = config
	r.err = nil
	return nil
}

func (w *fakeClusterFilesWriter) WriteClusterFiles(_ context.Context, config domain.ClusterConfig) error {
	if w.err != nil {
		return w.err
	}

	w.written = &config
	return nil
}
