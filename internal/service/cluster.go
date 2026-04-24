package service

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"dst-server-ctl/internal/domain"
)

type ClusterConfigService struct {
	repo   ClusterConfigRepository
	writer ClusterFilesWriter
	now    func() time.Time
}

type ClusterFilesWriter interface {
	WriteClusterFiles(ctx context.Context, config domain.ClusterConfig) error
}

func NewClusterConfigService(repo ClusterConfigRepository, writer ClusterFilesWriter) *ClusterConfigService {
	return &ClusterConfigService{
		repo:   repo,
		writer: writer,
		now:    time.Now,
	}
}

func (s *ClusterConfigService) Initialize(ctx context.Context) (domain.ClusterConfig, error) {
	config, err := s.repo.GetClusterConfig(ctx)
	if err == nil {
		return config, nil
	}
	if !errors.Is(err, domain.ErrClusterConfigNotFound) {
		return domain.ClusterConfig{}, err
	}

	now := s.now().UTC()
	config = defaultClusterConfig(now)
	if err := s.repo.SaveClusterConfig(ctx, config); err != nil {
		return domain.ClusterConfig{}, err
	}
	if err := s.writeClusterFiles(ctx, config); err != nil {
		return domain.ClusterConfig{}, err
	}

	return config, nil
}

func (s *ClusterConfigService) Get(ctx context.Context) (domain.ClusterConfig, error) {
	return s.repo.GetClusterConfig(ctx)
}

func (s *ClusterConfigService) Update(ctx context.Context, config domain.ClusterConfig) (domain.ClusterConfig, error) {
	normalized, err := normalizeClusterConfig(config)
	if err != nil {
		return domain.ClusterConfig{}, err
	}

	existing, err := s.repo.GetClusterConfig(ctx)
	if err != nil {
		return domain.ClusterConfig{}, err
	}

	normalized.CreatedAt = existing.CreatedAt
	normalized.UpdatedAt = s.now().UTC()
	if err := s.repo.SaveClusterConfig(ctx, normalized); err != nil {
		return domain.ClusterConfig{}, err
	}
	if err := s.writeClusterFiles(ctx, normalized); err != nil {
		return domain.ClusterConfig{}, err
	}

	return normalized, nil
}

func (s *ClusterConfigService) writeClusterFiles(ctx context.Context, config domain.ClusterConfig) error {
	if s.writer == nil {
		return nil
	}

	return s.writer.WriteClusterFiles(ctx, config)
}

func defaultClusterConfig(now time.Time) domain.ClusterConfig {
	return domain.ClusterConfig{
		ClusterName:        "DST Server",
		ClusterDescription: "",
		GameMode:           "survival",
		MaxPlayers:         6,
		Language:           "en",
		PVP:                false,
		PauseWhenEmpty:     true,
		Shards: []domain.ShardConfig{
			{Name: domain.ShardMaster, Enabled: true},
			{Name: domain.ShardCaves, Enabled: true},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func normalizeClusterConfig(config domain.ClusterConfig) (domain.ClusterConfig, error) {
	config.ClusterName = strings.TrimSpace(config.ClusterName)
	config.ClusterDescription = strings.TrimSpace(config.ClusterDescription)
	config.GameMode = strings.TrimSpace(config.GameMode)
	config.Language = strings.TrimSpace(config.Language)

	if config.ClusterName == "" {
		return domain.ClusterConfig{}, fmt.Errorf("%w: cluster name is required", domain.ErrInvalidClusterConfig)
	}
	if config.GameMode == "" {
		return domain.ClusterConfig{}, fmt.Errorf("%w: game mode is required", domain.ErrInvalidClusterConfig)
	}
	if config.Language == "" {
		return domain.ClusterConfig{}, fmt.Errorf("%w: language is required", domain.ErrInvalidClusterConfig)
	}
	if config.MaxPlayers < 1 {
		return domain.ClusterConfig{}, fmt.Errorf("%w: max players must be at least 1", domain.ErrInvalidClusterConfig)
	}
	if len(config.Shards) == 0 {
		return domain.ClusterConfig{}, fmt.Errorf("%w: at least one shard is required", domain.ErrInvalidClusterConfig)
	}

	normalizedShards := make([]domain.ShardConfig, 0, len(config.Shards))
	seen := make(map[domain.ShardName]struct{}, len(config.Shards))
	for _, shard := range config.Shards {
		if shard.Name != domain.ShardMaster && shard.Name != domain.ShardCaves {
			return domain.ClusterConfig{}, fmt.Errorf("%w: unsupported shard %q", domain.ErrInvalidClusterConfig, shard.Name)
		}
		if _, ok := seen[shard.Name]; ok {
			return domain.ClusterConfig{}, fmt.Errorf("%w: duplicate shard %q", domain.ErrInvalidClusterConfig, shard.Name)
		}
		seen[shard.Name] = struct{}{}
		normalizedShards = append(normalizedShards, domain.ShardConfig{
			Name:    shard.Name,
			Enabled: shard.Enabled,
		})
	}
	if _, ok := seen[domain.ShardMaster]; !ok {
		return domain.ClusterConfig{}, fmt.Errorf("%w: master shard is required", domain.ErrInvalidClusterConfig)
	}

	slices.SortFunc(normalizedShards, func(a, b domain.ShardConfig) int {
		return compareShardName(a.Name, b.Name)
	})
	config.Shards = normalizedShards

	return config, nil
}

func compareShardName(a, b domain.ShardName) int {
	order := map[domain.ShardName]int{
		domain.ShardMaster: 0,
		domain.ShardCaves:  1,
	}
	return order[a] - order[b]
}
