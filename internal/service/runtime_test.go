package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"dst-server-ctl/internal/adapter/command"
	"dst-server-ctl/internal/domain"
)

func TestRuntimeServiceStartLaunchesEnabledShards(t *testing.T) {
	now := time.Date(2026, 4, 24, 14, 0, 0, 0, time.UTC)
	process := &controlledRuntimeProcess{pid: 101, waitCh: make(chan error, 1)}
	starter := &fakeShardProcessStarter{
		processes: map[domain.ShardName]command.Process{
			domain.ShardMaster: process,
			domain.ShardCaves:  &controlledRuntimeProcess{pid: 202, waitCh: make(chan error, 1)},
		},
	}

	service := NewRuntimeService(
		domain.ManagedLayout{Root: "/srv/managed", DST: "/srv/managed/dst"},
		&fakeInstallationStateRepository{
			err: nil,
			state: domain.InstallationState{
				ManagedRoot:    "/srv/managed",
				DSTInstalledAt: &now,
			},
		},
		&fakeRuntimeClusterConfigRepository{
			config: domain.ClusterConfig{
				Shards: []domain.ShardConfig{
					{Name: domain.ShardMaster, Enabled: true},
					{Name: domain.ShardCaves, Enabled: false},
				},
			},
		},
		starter,
	)
	service.dispatch = func(fn func()) {}

	if err := service.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	if len(starter.started) != 1 {
		t.Fatalf("started shard count = %d, want 1", len(starter.started))
	}
	if starter.started[0] != domain.ShardMaster {
		t.Fatalf("first started shard = %q, want Master", starter.started[0])
	}
}

func TestRuntimeServiceStartRejectsWhenDSTNotInstalled(t *testing.T) {
	service := NewRuntimeService(
		domain.ManagedLayout{},
		&fakeInstallationStateRepository{state: domain.InstallationState{ManagedRoot: "/srv/managed"}},
		&fakeRuntimeClusterConfigRepository{},
		&fakeShardProcessStarter{},
	)

	err := service.Start(context.Background())
	if !errors.Is(err, domain.ErrDSTNotInstalled) {
		t.Fatalf("Start() error = %v, want %v", err, domain.ErrDSTNotInstalled)
	}
}

func TestRuntimeServiceStartRejectsWhenAlreadyRunning(t *testing.T) {
	now := time.Date(2026, 4, 24, 14, 0, 0, 0, time.UTC)
	service := NewRuntimeService(
		domain.ManagedLayout{},
		&fakeInstallationStateRepository{
			state: domain.InstallationState{
				ManagedRoot:    "/srv/managed",
				DSTInstalledAt: &now,
			},
		},
		&fakeRuntimeClusterConfigRepository{},
		&fakeShardProcessStarter{},
	)
	service.processes[domain.ShardMaster] = fakeRuntimeProcess{pid: 101}

	err := service.Start(context.Background())
	if !errors.Is(err, domain.ErrServerAlreadyRunning) {
		t.Fatalf("Start() error = %v, want %v", err, domain.ErrServerAlreadyRunning)
	}
}

func TestRuntimeServiceStartStopsStartedShardsWhenLaterShardFails(t *testing.T) {
	now := time.Date(2026, 4, 24, 14, 0, 0, 0, time.UTC)
	master := &trackedRuntimeProcess{pid: 101}
	starter := &fakeShardProcessStarter{
		processes: map[domain.ShardName]command.Process{
			domain.ShardMaster: master,
		},
		errs: map[domain.ShardName]error{
			domain.ShardCaves: errors.New("spawn failed"),
		},
	}

	service := NewRuntimeService(
		domain.ManagedLayout{Root: "/srv/managed", DST: "/srv/managed/dst"},
		&fakeInstallationStateRepository{
			state: domain.InstallationState{
				ManagedRoot:    "/srv/managed",
				DSTInstalledAt: &now,
			},
		},
		&fakeRuntimeClusterConfigRepository{
			config: domain.ClusterConfig{
				Shards: []domain.ShardConfig{
					{Name: domain.ShardMaster, Enabled: true},
					{Name: domain.ShardCaves, Enabled: true},
				},
			},
		},
		starter,
	)
	service.dispatch = func(fn func()) {}

	err := service.Start(context.Background())
	if err == nil {
		t.Fatal("Start() error = nil, want failure")
	}
	if !master.killed {
		t.Fatal("Master process was not killed after later shard failure")
	}
	if len(service.processes) != 0 {
		t.Fatalf("process count = %d, want 0", len(service.processes))
	}
}

func TestRuntimeServiceWatchShardClearsStateOnUnexpectedExit(t *testing.T) {
	now := time.Date(2026, 4, 24, 14, 0, 0, 0, time.UTC)
	process := &controlledRuntimeProcess{pid: 101, waitCh: make(chan error, 1)}
	starter := &fakeShardProcessStarter{
		processes: map[domain.ShardName]command.Process{
			domain.ShardMaster: process,
		},
	}

	service := NewRuntimeService(
		domain.ManagedLayout{Root: "/srv/managed", DST: "/srv/managed/dst"},
		&fakeInstallationStateRepository{
			state: domain.InstallationState{
				ManagedRoot:    "/srv/managed",
				DSTInstalledAt: &now,
			},
		},
		&fakeRuntimeClusterConfigRepository{
			config: domain.ClusterConfig{
				Shards: []domain.ShardConfig{{Name: domain.ShardMaster, Enabled: true}},
			},
		},
		starter,
	)
	watchDone := make(chan struct{}, 1)
	service.dispatch = func(fn func()) {
		go func() {
			fn()
			watchDone <- struct{}{}
		}()
	}

	if err := service.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	process.waitCh <- errors.New("exit 1")
	<-watchDone

	status, err := service.Status(context.Background())
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}
	if status.Status != domain.ServerStatusStopped {
		t.Fatalf("status = %q, want stopped", status.Status)
	}
	if status.LastError == "" {
		t.Fatal("LastError = empty, want exit message")
	}
}

func TestRuntimeServiceStatusReturnsRunningShards(t *testing.T) {
	service := NewRuntimeService(
		domain.ManagedLayout{},
		&fakeInstallationStateRepository{},
		&fakeRuntimeClusterConfigRepository{config: domain.ClusterConfig{
			ClusterName:    "DST Server",
			GameMode:       "survival",
			MaxPlayers:     6,
			Language:       "en",
			PauseWhenEmpty: true,
			Shards: []domain.ShardConfig{
				{Name: domain.ShardMaster, Enabled: true},
				{Name: domain.ShardCaves, Enabled: true},
			},
		}},
		&fakeShardProcessStarter{},
	)
	service.processes[domain.ShardCaves] = fakeRuntimeProcess{pid: 202}
	service.processes[domain.ShardMaster] = fakeRuntimeProcess{pid: 101}
	service.startedConfig = &domain.ClusterConfig{
		ClusterName:    "DST Server",
		GameMode:       "survival",
		MaxPlayers:     6,
		Language:       "en",
		PauseWhenEmpty: true,
		Shards: []domain.ShardConfig{
			{Name: domain.ShardMaster, Enabled: true},
			{Name: domain.ShardCaves, Enabled: true},
		},
	}

	status, err := service.Status(context.Background())
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}
	if status.Status != domain.ServerStatusRunning {
		t.Fatalf("status = %q, want running", status.Status)
	}
	if len(status.Shards) != 2 {
		t.Fatalf("shard count = %d, want 2", len(status.Shards))
	}
	if status.Shards[0].Name != domain.ShardMaster || status.Shards[0].PID != 101 {
		t.Fatalf("first shard = %#v, want master pid 101", status.Shards[0])
	}
	if status.RestartRequired {
		t.Fatal("RestartRequired = true, want false")
	}
}

func TestRuntimeServiceStatusMarksRestartRequiredWhenConfigChanged(t *testing.T) {
	service := NewRuntimeService(
		domain.ManagedLayout{},
		&fakeInstallationStateRepository{},
		&fakeRuntimeClusterConfigRepository{config: domain.ClusterConfig{
			ClusterName:    "Changed",
			GameMode:       "survival",
			MaxPlayers:     6,
			Language:       "en",
			PauseWhenEmpty: true,
			Shards:         []domain.ShardConfig{{Name: domain.ShardMaster, Enabled: true}},
		}},
		&fakeShardProcessStarter{},
	)
	service.processes[domain.ShardMaster] = fakeRuntimeProcess{pid: 101}
	service.startedConfig = &domain.ClusterConfig{
		ClusterName:    "Original",
		GameMode:       "survival",
		MaxPlayers:     6,
		Language:       "en",
		PauseWhenEmpty: true,
		Shards:         []domain.ShardConfig{{Name: domain.ShardMaster, Enabled: true}},
	}

	status, err := service.Status(context.Background())
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}
	if !status.RestartRequired {
		t.Fatal("RestartRequired = false, want true")
	}
}

func TestRuntimeServiceRestartRestartsShards(t *testing.T) {
	now := time.Date(2026, 4, 24, 14, 0, 0, 0, time.UTC)
	oldProcess := &trackedRuntimeProcess{pid: 10}
	newProcess := &controlledRuntimeProcess{pid: 20, waitCh: make(chan error, 1)}
	starter := &fakeShardProcessStarter{
		processes: map[domain.ShardName]command.Process{
			domain.ShardMaster: newProcess,
		},
	}
	service := NewRuntimeService(
		domain.ManagedLayout{},
		&fakeInstallationStateRepository{state: domain.InstallationState{ManagedRoot: "/srv/managed", DSTInstalledAt: &now}},
		&fakeRuntimeClusterConfigRepository{config: domain.ClusterConfig{
			ClusterName:    "DST Server",
			GameMode:       "survival",
			MaxPlayers:     6,
			Language:       "en",
			PauseWhenEmpty: true,
			Shards:         []domain.ShardConfig{{Name: domain.ShardMaster, Enabled: true}},
		}},
		starter,
	)
	service.processes[domain.ShardMaster] = oldProcess
	service.cancels[domain.ShardMaster] = func() {}
	service.dispatch = func(fn func()) {}

	if err := service.Restart(context.Background()); err != nil {
		t.Fatalf("Restart() error = %v", err)
	}
	if !oldProcess.killed {
		t.Fatal("old process not killed during restart")
	}
	if service.processes[domain.ShardMaster] != newProcess {
		t.Fatalf("new process = %#v, want restarted process", service.processes[domain.ShardMaster])
	}
}

func TestRuntimeServiceStopKillsRunningShards(t *testing.T) {
	master := &trackedRuntimeProcess{pid: 101}
	caves := &trackedRuntimeProcess{pid: 202}
	service := NewRuntimeService(
		domain.ManagedLayout{},
		&fakeInstallationStateRepository{},
		&fakeRuntimeClusterConfigRepository{},
		&fakeShardProcessStarter{},
	)
	service.processes[domain.ShardMaster] = master
	service.processes[domain.ShardCaves] = caves
	service.cancels[domain.ShardMaster] = func() {}
	service.cancels[domain.ShardCaves] = func() {}

	if err := service.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	if !master.killed || !caves.killed {
		t.Fatalf("killed master=%v caves=%v, want both true", master.killed, caves.killed)
	}
	if len(service.processes) != 0 {
		t.Fatalf("process count = %d, want 0", len(service.processes))
	}
}

func TestRuntimeServiceStopRejectsWhenNotRunning(t *testing.T) {
	service := NewRuntimeService(
		domain.ManagedLayout{},
		&fakeInstallationStateRepository{},
		&fakeRuntimeClusterConfigRepository{},
		&fakeShardProcessStarter{},
	)

	err := service.Stop(context.Background())
	if !errors.Is(err, domain.ErrServerNotRunning) {
		t.Fatalf("Stop() error = %v, want %v", err, domain.ErrServerNotRunning)
	}
}

type fakeRuntimeClusterConfigRepository struct {
	config domain.ClusterConfig
	err    error
}

func (r *fakeRuntimeClusterConfigRepository) GetClusterConfig(context.Context) (domain.ClusterConfig, error) {
	if r.err != nil {
		return domain.ClusterConfig{}, r.err
	}
	return r.config, nil
}

func (r *fakeRuntimeClusterConfigRepository) SaveClusterConfig(context.Context, domain.ClusterConfig) error {
	return nil
}

type fakeShardProcessStarter struct {
	started   []domain.ShardName
	processes map[domain.ShardName]command.Process
	errs      map[domain.ShardName]error
}

func (s *fakeShardProcessStarter) StartShard(_ context.Context, _ domain.ManagedLayout, shard domain.ShardName) (command.Process, error) {
	s.started = append(s.started, shard)
	if err := s.errs[shard]; err != nil {
		return nil, err
	}
	if process := s.processes[shard]; process != nil {
		return process, nil
	}
	return fakeRuntimeProcess{pid: 1}, nil
}

type fakeRuntimeProcess struct {
	pid int
}

func (p fakeRuntimeProcess) PID() int    { return p.pid }
func (p fakeRuntimeProcess) Wait() error { return nil }
func (p fakeRuntimeProcess) Kill() error { return nil }

type trackedRuntimeProcess struct {
	pid    int
	killed bool
}

func (p *trackedRuntimeProcess) PID() int    { return p.pid }
func (p *trackedRuntimeProcess) Wait() error { return nil }
func (p *trackedRuntimeProcess) Kill() error {
	p.killed = true
	return nil
}

type controlledRuntimeProcess struct {
	pid    int
	waitCh chan error
	killed bool
}

func (p *controlledRuntimeProcess) PID() int { return p.pid }

func (p *controlledRuntimeProcess) Wait() error {
	return <-p.waitCh
}

func (p *controlledRuntimeProcess) Kill() error {
	p.killed = true
	return nil
}
