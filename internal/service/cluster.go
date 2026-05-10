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
		ClusterPassword:    "",
		ClusterIntention:   "cooperative",
		GameMode:           "survival",
		MaxPlayers:         6,
		Language:           "en",
		PVP:                false,
		PauseWhenEmpty:     true,
		OfflineCluster:     false,
		LANOnlyCluster:     false,
		TickRate:           15,
		ConsoleEnabled:     true,
		BindIP:             "127.0.0.1",
		MasterPort:         10888,
		ClusterKey:         "dst-server-ctl",
		Shards: []domain.ShardConfig{
			{Name: domain.ShardMaster, Enabled: true, ServerPort: 10999, MasterServerPort: 27016, AuthenticationPort: 8766, WorldGenPreset: "SURVIVAL_TOGETHER", WorldGenOverrides: map[string]string{}},
			{Name: domain.ShardCaves, Enabled: true, ServerPort: 11000, MasterServerPort: 27017, AuthenticationPort: 8767, WorldGenPreset: "DST_CAVE", WorldGenOverrides: map[string]string{}},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func normalizeClusterConfig(config domain.ClusterConfig) (domain.ClusterConfig, error) {
	config.ClusterName = strings.TrimSpace(config.ClusterName)
	config.ClusterDescription = strings.TrimSpace(config.ClusterDescription)
	config.ClusterPassword = strings.TrimSpace(config.ClusterPassword)
	config.ClusterIntention = strings.TrimSpace(config.ClusterIntention)
	config.GameMode = strings.TrimSpace(config.GameMode)
	config.Language = strings.TrimSpace(config.Language)
	config.BindIP = strings.TrimSpace(config.BindIP)
	config.ClusterKey = strings.TrimSpace(config.ClusterKey)

	if config.ClusterName == "" {
		return domain.ClusterConfig{}, fmt.Errorf("%w: cluster name is required", domain.ErrInvalidClusterConfig)
	}
	if config.ClusterIntention == "" {
		return domain.ClusterConfig{}, fmt.Errorf("%w: cluster intention is required", domain.ErrInvalidClusterConfig)
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
	if config.TickRate < 15 || config.TickRate > 60 {
		return domain.ClusterConfig{}, fmt.Errorf("%w: tick rate must be between 15 and 60", domain.ErrInvalidClusterConfig)
	}
	if len(config.Shards) == 0 {
		return domain.ClusterConfig{}, fmt.Errorf("%w: at least one shard is required", domain.ErrInvalidClusterConfig)
	}

	normalizedShards := make([]domain.ShardConfig, 0, len(config.Shards))
	seen := make(map[domain.ShardName]struct{}, len(config.Shards))
	enabledShardCount := 0
	usedServerPorts := map[int]domain.ShardName{}
	usedMasterServerPorts := map[int]domain.ShardName{}
	usedAuthenticationPorts := map[int]domain.ShardName{}
	for _, shard := range config.Shards {
		if shard.Name != domain.ShardMaster && shard.Name != domain.ShardCaves {
			return domain.ClusterConfig{}, fmt.Errorf("%w: unsupported shard %q", domain.ErrInvalidClusterConfig, shard.Name)
		}
		if _, ok := seen[shard.Name]; ok {
			return domain.ClusterConfig{}, fmt.Errorf("%w: duplicate shard %q", domain.ErrInvalidClusterConfig, shard.Name)
		}
		seen[shard.Name] = struct{}{}
		if shard.ServerPort < 1 || shard.ServerPort > 65535 {
			return domain.ClusterConfig{}, fmt.Errorf("%w: %s server port must be between 1 and 65535", domain.ErrInvalidClusterConfig, shard.Name)
		}
		if shard.MasterServerPort < 1 || shard.MasterServerPort > 65535 {
			return domain.ClusterConfig{}, fmt.Errorf("%w: %s master server port must be between 1 and 65535", domain.ErrInvalidClusterConfig, shard.Name)
		}
		if shard.AuthenticationPort < 1 || shard.AuthenticationPort > 65535 {
			return domain.ClusterConfig{}, fmt.Errorf("%w: %s authentication port must be between 1 and 65535", domain.ErrInvalidClusterConfig, shard.Name)
		}
		shard.WorldGenPreset = strings.TrimSpace(shard.WorldGenPreset)
		if shard.WorldGenPreset == "" {
			return domain.ClusterConfig{}, fmt.Errorf("%w: %s world preset is required", domain.ErrInvalidClusterConfig, shard.Name)
		}
		normalizedOverrides := make(map[string]string, len(shard.WorldGenOverrides))
		for key, value := range shard.WorldGenOverrides {
			trimmedKey := strings.TrimSpace(key)
			trimmedValue := strings.TrimSpace(value)
			if trimmedKey == "" {
				return domain.ClusterConfig{}, fmt.Errorf("%w: %s world override key is required", domain.ErrInvalidClusterConfig, shard.Name)
			}
			if trimmedValue == "" {
				return domain.ClusterConfig{}, fmt.Errorf("%w: %s world override %q value is required", domain.ErrInvalidClusterConfig, shard.Name, trimmedKey)
			}
			normalizedOverrides[trimmedKey] = trimmedValue
		}
		if other, ok := usedServerPorts[shard.ServerPort]; ok {
			return domain.ClusterConfig{}, fmt.Errorf("%w: %s server port conflicts with %s", domain.ErrInvalidClusterConfig, shard.Name, other)
		}
		if other, ok := usedMasterServerPorts[shard.MasterServerPort]; ok {
			return domain.ClusterConfig{}, fmt.Errorf("%w: %s master server port conflicts with %s", domain.ErrInvalidClusterConfig, shard.Name, other)
		}
		if other, ok := usedAuthenticationPorts[shard.AuthenticationPort]; ok {
			return domain.ClusterConfig{}, fmt.Errorf("%w: %s authentication port conflicts with %s", domain.ErrInvalidClusterConfig, shard.Name, other)
		}
		usedServerPorts[shard.ServerPort] = shard.Name
		usedMasterServerPorts[shard.MasterServerPort] = shard.Name
		usedAuthenticationPorts[shard.AuthenticationPort] = shard.Name
		if shard.Enabled {
			enabledShardCount++
		}
		normalizedShards = append(normalizedShards, domain.ShardConfig{
			Name:               shard.Name,
			Enabled:            shard.Enabled,
			ServerPort:         shard.ServerPort,
			MasterServerPort:   shard.MasterServerPort,
			AuthenticationPort: shard.AuthenticationPort,
			WorldGenPreset:     shard.WorldGenPreset,
			WorldGenOverrides:  normalizedOverrides,
		})
	}
	if _, ok := seen[domain.ShardMaster]; !ok {
		return domain.ClusterConfig{}, fmt.Errorf("%w: master shard is required", domain.ErrInvalidClusterConfig)
	}
	if enabledShardCount > 1 {
		if config.BindIP == "" {
			return domain.ClusterConfig{}, fmt.Errorf("%w: bind ip is required when multiple shards are enabled", domain.ErrInvalidClusterConfig)
		}
		if config.MasterPort < 1 || config.MasterPort > 65535 {
			return domain.ClusterConfig{}, fmt.Errorf("%w: master port must be between 1 and 65535", domain.ErrInvalidClusterConfig)
		}
		if config.ClusterKey == "" {
			return domain.ClusterConfig{}, fmt.Errorf("%w: cluster key is required when multiple shards are enabled", domain.ErrInvalidClusterConfig)
		}
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
