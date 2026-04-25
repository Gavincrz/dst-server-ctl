package service

import (
	"context"
	"errors"
	"fmt"
	"slices"
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
	stopping  map[domain.ShardName]bool
	dispatch  func(func())
	lastError string
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
		stopping:  make(map[domain.ShardName]bool),
		dispatch: func(fn func()) {
			go fn()
		},
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
		s.lastError = err.Error()
		return err
	}
	if state.DSTInstalledAt == nil {
		s.lastError = domain.ErrDSTNotInstalled.Error()
		return domain.ErrDSTNotInstalled
	}

	config, err := s.clusters.GetClusterConfig(ctx)
	if err != nil {
		s.lastError = err.Error()
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
			startErr := fmt.Errorf("start shard %s: %w", shard.Name, err)
			s.lastError = startErr.Error()
			return startErr
		}

		s.processes[shard.Name] = process
		s.cancels[shard.Name] = cancel
		s.stopping[shard.Name] = false
		started = append(started, shard.Name)
		s.watchShard(shard.Name, process)
	}

	s.lastError = ""
	return nil
}

func (s *RuntimeService) Stop(context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.processes) == 0 {
		return domain.ErrServerNotRunning
	}

	s.stopStartedLocked(s.runningShardsLocked())
	s.lastError = ""
	return nil
}

func (s *RuntimeService) Status(context.Context) (domain.RuntimeStatus, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	shards := make([]domain.ShardState, 0, len(s.processes))
	for _, shard := range s.runningShardsLocked() {
		process := s.processes[shard]
		shards = append(shards, domain.ShardState{
			Name:    shard,
			Running: true,
			PID:     process.PID(),
		})
	}

	status := domain.ServerStatusStopped
	if len(shards) > 0 {
		status = domain.ServerStatusRunning
	}

	return domain.RuntimeStatus{
		Status:    status,
		Shards:    shards,
		LastError: s.lastError,
	}, nil
}

func (s *RuntimeService) stopStartedLocked(shards []domain.ShardName) {
	for i := len(shards) - 1; i >= 0; i-- {
		shard := shards[i]
		s.stopping[shard] = true
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

func (s *RuntimeService) runningShardsLocked() []domain.ShardName {
	shards := make([]domain.ShardName, 0, len(s.processes))
	for shard := range s.processes {
		shards = append(shards, shard)
	}
	slices.SortFunc(shards, compareShardName)
	return shards
}

func (s *RuntimeService) watchShard(shard domain.ShardName, process command.Process) {
	s.dispatch(func() {
		err := process.Wait()

		s.mu.Lock()
		defer s.mu.Unlock()

		current, ok := s.processes[shard]
		if !ok || current != process {
			return
		}

		intentional := s.stopping[shard]
		delete(s.processes, shard)
		delete(s.cancels, shard)
		delete(s.stopping, shard)

		if intentional || err == nil || errors.Is(err, context.Canceled) {
			return
		}

		s.lastError = fmt.Sprintf("shard %s exited: %v", shard, err)
	})
}
