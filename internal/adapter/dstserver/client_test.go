package dstserver

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"dst-server-ctl/internal/adapter/command"
	"dst-server-ctl/internal/domain"
)

func TestClientStartShardUsesProcessRunner(t *testing.T) {
	runner := &fakeRunner{process: fakeProcess{pid: 42}}
	client := NewClient(runner)

	process, err := client.StartShard(context.Background(), domain.ManagedLayout{
		Root: "/srv/managed",
		Logs: "/srv/managed/logs",
		DST:  "/srv/managed/dst",
	}, domain.ShardMaster)
	if err != nil {
		t.Fatalf("StartShard() error = %v", err)
	}

	if process.PID() != 42 {
		t.Fatalf("PID() = %d, want 42", process.PID())
	}
	if len(runner.calls) != 1 {
		t.Fatalf("call count = %d, want 1", len(runner.calls))
	}
	if runner.calls[0].name != "/srv/managed/dst/bin64/dontstarve_dedicated_server_nullrenderer" {
		t.Fatalf("name = %q", runner.calls[0].name)
	}

	wantArgs := []string{
		"-persistent_storage_root", "/srv/managed",
		"-conf_dir", "clusters",
		"-cluster", "primary",
		"-console",
		"-shard", "Master",
	}
	if !reflect.DeepEqual(runner.calls[0].args, wantArgs) {
		t.Fatalf("args = %#v, want %#v", runner.calls[0].args, wantArgs)
	}
	if runner.calls[0].options.StdoutPath != "/srv/managed/logs/master.log" {
		t.Fatalf("stdout path = %q, want /srv/managed/logs/master.log", runner.calls[0].options.StdoutPath)
	}
	if runner.calls[0].options.StderrPath != "/srv/managed/logs/master.log" {
		t.Fatalf("stderr path = %q, want /srv/managed/logs/master.log", runner.calls[0].options.StderrPath)
	}
}

func TestClientStartShardReturnsRunnerError(t *testing.T) {
	wantErr := errors.New("start failed")
	runner := &fakeRunner{err: wantErr}
	client := NewClient(runner)

	_, err := client.StartShard(context.Background(), domain.ManagedLayout{}, domain.ShardCaves)
	if !errors.Is(err, wantErr) {
		t.Fatalf("StartShard() error = %v, want %v", err, wantErr)
	}
}

type fakeRunnerCall struct {
	name    string
	args    []string
	options command.StartOptions
}

type fakeRunner struct {
	calls   []fakeRunnerCall
	process command.Process
	err     error
}

func (r *fakeRunner) Run(context.Context, string, ...string) (command.Result, error) {
	return command.Result{}, nil
}

func (r *fakeRunner) Start(_ context.Context, name string, args ...string) (command.Process, error) {
	return r.StartWithOptions(context.Background(), command.StartOptions{}, name, args...)
}

func (r *fakeRunner) StartWithOptions(_ context.Context, options command.StartOptions, name string, args ...string) (command.Process, error) {
	r.calls = append(r.calls, fakeRunnerCall{name: name, args: args, options: options})
	if r.err != nil {
		return nil, r.err
	}
	return r.process, nil
}

type fakeProcess struct {
	pid int
}

func (p fakeProcess) PID() int    { return p.pid }
func (p fakeProcess) Wait() error { return nil }
func (p fakeProcess) Kill() error { return nil }
