package service

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"

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
	events   RuntimeEventRepository
	starter  ShardProcessStarter

	mu            sync.Mutex
	processes     map[domain.ShardName]command.Process
	cancels       map[domain.ShardName]context.CancelFunc
	stopping      map[domain.ShardName]bool
	retries       map[domain.ShardName]int
	dispatch      func(func())
	startedConfig *domain.ClusterConfig
	now           func() time.Time
	lastError     string
}

func NewRuntimeService(
	layout domain.ManagedLayout,
	installs InstallationStateRepository,
	clusters ClusterConfigRepository,
	events RuntimeEventRepository,
	starter ShardProcessStarter,
) *RuntimeService {
	return &RuntimeService{
		layout:    layout,
		installs:  installs,
		clusters:  clusters,
		events:    events,
		starter:   starter,
		processes: make(map[domain.ShardName]command.Process),
		cancels:   make(map[domain.ShardName]context.CancelFunc),
		stopping:  make(map[domain.ShardName]bool),
		retries:   make(map[domain.ShardName]int),
		dispatch: func(fn func()) {
			go fn()
		},
		now: time.Now,
	}
}

func (s *RuntimeService) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.startLocked(ctx)
}

func (s *RuntimeService) Restart(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.processes) > 0 {
		s.stopStartedLocked(s.runningShardsLocked())
	}

	return s.startLocked(ctx)
}

func (s *RuntimeService) startLocked(ctx context.Context) error {

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

	startedConfig := config
	started := make([]domain.ShardName, 0, len(config.Shards))
	for _, shard := range config.Shards {
		if !shard.Enabled {
			continue
		}

		process, cancel, err := s.startShardLocked(shard.Name)
		if err != nil {
			s.stopStartedLocked(started)
			startErr := fmt.Errorf("start shard %s: %w", shard.Name, err)
			s.lastError = startErr.Error()
			return startErr
		}

		s.processes[shard.Name] = process
		s.cancels[shard.Name] = cancel
		s.stopping[shard.Name] = false
		s.retries[shard.Name] = 0
		started = append(started, shard.Name)
		s.watchShard(shard.Name, process)
		s.recordEventLocked(shard.Name, domain.RuntimeEventStarted, fmt.Sprintf("%s shard started with PID %d", shard.Name, process.PID()))
	}

	s.startedConfig = cloneClusterConfig(startedConfig)
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
	s.startedConfig = nil
	s.lastError = ""
	return nil
}

func (s *RuntimeService) Status(ctx context.Context) (domain.RuntimeStatus, error) {
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

	restartRequired := false
	if len(shards) > 0 {
		config, err := s.clusters.GetClusterConfig(ctx)
		if err != nil {
			return domain.RuntimeStatus{}, err
		}
		restartRequired = !clusterConfigsEqual(s.startedConfig, &config)
	}

	return domain.RuntimeStatus{
		Status:          status,
		Shards:          shards,
		RestartRequired: restartRequired,
		LastError:       s.lastError,
	}, nil
}

func (s *RuntimeService) stopStartedLocked(shards []domain.ShardName) {
	for i := len(shards) - 1; i >= 0; i-- {
		shard := shards[i]
		s.stopping[shard] = true
		s.recordEventLocked(shard, domain.RuntimeEventStopped, fmt.Sprintf("%s shard stop requested", shard))
		if cancel, ok := s.cancels[shard]; ok {
			cancel()
			delete(s.cancels, shard)
		}
		if process, ok := s.processes[shard]; ok {
			_ = process.Kill()
			delete(s.processes, shard)
		}
		delete(s.retries, shard)
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

		s.recordEventLocked(shard, domain.RuntimeEventExited, fmt.Sprintf("%s shard exited unexpectedly: %v", shard, err))
		if s.retries[shard] < 1 {
			s.retries[shard]++
			process, cancel, restartErr := s.startShardLocked(shard)
			if restartErr == nil {
				s.processes[shard] = process
				s.cancels[shard] = cancel
				s.stopping[shard] = false
				s.recordEventLocked(shard, domain.RuntimeEventRetried, fmt.Sprintf("%s shard auto-restarted with PID %d", shard, process.PID()))
				s.lastError = fmt.Sprintf("shard %s exited and was restarted automatically", shard)
				s.watchShard(shard, process)
				return
			}
			s.recordEventLocked(shard, domain.RuntimeEventRetried, fmt.Sprintf("%s shard auto-restart failed: %v", shard, restartErr))
			s.lastError = fmt.Sprintf("shard %s exited and auto-restart failed: %v", shard, restartErr)
			if len(s.processes) == 0 {
				s.startedConfig = nil
			}
			return
		}

		if len(s.processes) == 0 {
			s.startedConfig = nil
		}
		s.lastError = fmt.Sprintf("shard %s exited: %v", shard, err)
	})
}

func (s *RuntimeService) startShardLocked(shard domain.ShardName) (command.Process, context.CancelFunc, error) {
	shardCtx, cancel := context.WithCancel(context.Background())
	process, err := s.starter.StartShard(shardCtx, s.layout, shard)
	if err != nil {
		cancel()
		return nil, nil, err
	}
	return process, cancel, nil
}

func clusterConfigsEqual(a, b *domain.ClusterConfig) bool {
	if a == nil || b == nil {
		return a == b
	}
	if a.ClusterName != b.ClusterName ||
		a.ClusterDescription != b.ClusterDescription ||
		a.GameMode != b.GameMode ||
		a.MaxPlayers != b.MaxPlayers ||
		a.Language != b.Language ||
		a.PVP != b.PVP ||
		a.PauseWhenEmpty != b.PauseWhenEmpty ||
		len(a.Shards) != len(b.Shards) {
		return false
	}
	for i := range a.Shards {
		if a.Shards[i] != b.Shards[i] {
			return false
		}
	}
	return true
}

func cloneClusterConfig(config domain.ClusterConfig) *domain.ClusterConfig {
	cloned := config
	cloned.Shards = append([]domain.ShardConfig(nil), config.Shards...)
	return &cloned
}

func (s *RuntimeService) recordEventLocked(shard domain.ShardName, kind domain.RuntimeEventKind, detail string) {
	if s.events == nil {
		return
	}
	_ = s.events.CreateRuntimeEvent(context.Background(), domain.RuntimeEvent{
		Shard:     shard,
		Kind:      kind,
		Detail:    detail,
		CreatedAt: s.now().UTC(),
	})
}
