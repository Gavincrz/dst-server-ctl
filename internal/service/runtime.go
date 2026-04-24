package service

import (
	"context"
	"fmt"
	"sync"

	"dst-server-ctl/internal/adapter/command"
	"dst-server-ctl/internal/domain"
)

type ShardProcessStarter interface {
	StartShard(ctx context.Context, layout domain.ManagedLayout, shard domain.ShardName) (command.Process, error)
}

type RuntimeService struct {
	layout   domain.ManagedLayout
	installs InstallationStateRepository
	clusters ClusterConfigRepository
	starter  ShardProcessStarter

	mu        sync.Mutex
	processes map[domain.ShardName]command.Process
	cancels   map[domain.ShardName]context.CancelFunc
}

func NewRuntimeService(
	layout domain.ManagedLayout,
	installs InstallationStateRepository,
	clusters ClusterConfigRepository,
	starter ShardProcessStarter,
) *RuntimeService {
	return &RuntimeService{
		layout:    layout,
		installs:  installs,
		clusters:  clusters,
		starter:   starter,
		processes: make(map[domain.ShardName]command.Process),
		cancels:   make(map[domain.ShardName]context.CancelFunc),
	}
}

func (s *RuntimeService) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.processes) > 0 {
		return domain.ErrServerAlreadyRunning
	}

	state, err := s.installs.GetInstallationState(ctx)
	if err != nil {
		return err
	}
	if state.DSTInstalledAt == nil {
		return domain.ErrDSTNotInstalled
	}

	config, err := s.clusters.GetClusterConfig(ctx)
	if err != nil {
		return err
	}

	started := make([]domain.ShardName, 0, len(config.Shards))
	for _, shard := range config.Shards {
		if !shard.Enabled {
			continue
		}

		shardCtx, cancel := context.WithCancel(context.Background())
		process, err := s.starter.StartShard(shardCtx, s.layout, shard.Name)
		if err != nil {
			cancel()
			s.stopStartedLocked(started)
			return fmt.Errorf("start shard %s: %w", shard.Name, err)
		}

		s.processes[shard.Name] = process
		s.cancels[shard.Name] = cancel
		started = append(started, shard.Name)
	}

	return nil
}

func (s *RuntimeService) stopStartedLocked(shards []domain.ShardName) {
	for i := len(shards) - 1; i >= 0; i-- {
		shard := shards[i]
		if cancel, ok := s.cancels[shard]; ok {
			cancel()
			delete(s.cancels, shard)
		}
		if process, ok := s.processes[shard]; ok {
			_ = process.Kill()
			delete(s.processes, shard)
		}
	}
}
