package steamcmd

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"dst-server-ctl/internal/adapter/command"
	"dst-server-ctl/internal/domain"
)

func TestClientInstallSteamCMDUsesCommandRunner(t *testing.T) {
	runner := &fakeRunner{}
	client := NewClient(runner)

	_, err := client.InstallSteamCMD(context.Background(), domain.ManagedLayout{SteamCMD: "/srv/managed/steamcmd"})
	if err != nil {
		t.Fatalf("InstallSteamCMD() error = %v", err)
	}

	if len(runner.calls) != 2 {
		t.Fatalf("call count = %d, want 2", len(runner.calls))
	}

	if runner.calls[0].name != "curl" {
		t.Fatalf("first command = %q, want curl", runner.calls[0].name)
	}

	wantDownloadArgs := []string{
		"-fsSL",
		"https://steamcdn-a.akamaihd.net/client/installer/steamcmd_linux.tar.gz",
		"-o",
		"/srv/managed/steamcmd/steamcmd_linux.tar.gz",
	}
	if !reflect.DeepEqual(runner.calls[0].args, wantDownloadArgs) {
		t.Fatalf("download args = %#v, want %#v", runner.calls[0].args, wantDownloadArgs)
	}

	if runner.calls[1].name != "tar" {
		t.Fatalf("second command = %q, want tar", runner.calls[1].name)
	}

	wantExtractArgs := []string{
		"-xzf",
		"/srv/managed/steamcmd/steamcmd_linux.tar.gz",
		"-C",
		"/srv/managed/steamcmd",
	}
	if !reflect.DeepEqual(runner.calls[1].args, wantExtractArgs) {
		t.Fatalf("extract args = %#v, want %#v", runner.calls[1].args, wantExtractArgs)
	}
}

func TestClientInstallSteamCMDReturnsDownloadError(t *testing.T) {
	wantErr := errors.New("curl failed")
	runner := &fakeRunner{errs: []error{wantErr}}
	client := NewClient(runner)

	_, err := client.InstallSteamCMD(context.Background(), domain.ManagedLayout{SteamCMD: "/srv/managed/steamcmd"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("InstallSteamCMD() error = %v, want %v", err, wantErr)
	}
}

func TestClientInstallDSTUsesCommandRunner(t *testing.T) {
	runner := &fakeRunner{}
	client := NewClient(runner)

	_, err := client.InstallDST(context.Background(), domain.ManagedLayout{
		SteamCMD: "/srv/managed/steamcmd",
		DST:      "/srv/dst",
	})
	if err != nil {
		t.Fatalf("InstallDST() error = %v", err)
	}

	if len(runner.calls) != 1 {
		t.Fatalf("call count = %d, want 1", len(runner.calls))
	}
	if runner.calls[0].name != "/srv/managed/steamcmd/steamcmd.sh" {
		t.Fatalf("name = %q, want /srv/managed/steamcmd/steamcmd.sh", runner.calls[0].name)
	}
	wantArgs := []string{
		"+force_install_dir", "/srv/dst",
		"+login", "anonymous",
		"+app_update", "343050", "validate",
		"+quit",
	}
	if !reflect.DeepEqual(runner.calls[0].args, wantArgs) {
		t.Fatalf("args = %#v, want %#v", runner.calls[0].args, wantArgs)
	}
}

type fakeRunnerCall struct {
	name string
	args []string
}

type fakeRunner struct {
	calls []fakeRunnerCall
	errs  []error
}

func (r *fakeRunner) Run(_ context.Context, name string, args ...string) (command.Result, error) {
	r.calls = append(r.calls, fakeRunnerCall{name: name, args: args})
	if len(r.errs) == 0 {
		return command.Result{}, nil
	}
	err := r.errs[0]
	r.errs = r.errs[1:]
	return command.Result{}, err
}

func (r *fakeRunner) Start(context.Context, string, ...string) (command.Process, error) {
	return nil, nil
}

func (r *fakeRunner) StartWithOptions(context.Context, command.StartOptions, string, ...string) (command.Process, error) {
	return nil, nil
}
