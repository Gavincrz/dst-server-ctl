package steamcmd

import (
	"context"
	"reflect"
	"testing"

	"dst-server-ctl/internal/adapter/command"
	"dst-server-ctl/internal/domain"
)

func TestClientInstallDSTUsesCommandRunner(t *testing.T) {
	runner := &fakeRunner{}
	client := NewClient(runner)

	_, err := client.InstallDST(context.Background(), domain.ManagedLayout{DST: "/srv/dst"})
	if err != nil {
		t.Fatalf("InstallDST() error = %v", err)
	}

	if runner.name != "steamcmd" {
		t.Fatalf("name = %q, want steamcmd", runner.name)
	}

	wantArgs := []string{
		"+force_install_dir", "/srv/dst",
		"+login", "anonymous",
		"+app_update", "343050", "validate",
		"+quit",
	}
	if !reflect.DeepEqual(runner.args, wantArgs) {
		t.Fatalf("args = %#v, want %#v", runner.args, wantArgs)
	}
}

type fakeRunner struct {
	name string
	args []string
}

func (r *fakeRunner) Run(_ context.Context, name string, args ...string) (command.Result, error) {
	r.name = name
	r.args = args
	return command.Result{}, nil
}
