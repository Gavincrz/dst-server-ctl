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
	starter := &fakeShardProcessStarter{
		processes: map[domain.ShardName]command.Process{
			domain.ShardMaster: fakeRuntimeProcess{pid: 101},
			domain.ShardCaves:  fakeRuntimeProcess{pid: 202},
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

func TestRuntimeServiceStatusReturnsRunningShards(t *testing.T) {
	service := NewRuntimeService(
		domain.ManagedLayout{},
		&fakeInstallationStateRepository{},
		&fakeRuntimeClusterConfigRepository{},
		&fakeShardProcessStarter{},
	)
	service.processes[domain.ShardCaves] = fakeRuntimeProcess{pid: 202}
	service.processes[domain.ShardMaster] = fakeRuntimeProcess{pid: 101}

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
