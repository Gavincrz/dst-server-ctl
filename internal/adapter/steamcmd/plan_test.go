package steamcmd

import (
	"reflect"
	"testing"

	"dst-server-ctl/internal/domain"
)

func TestInstallDSTPlanUsesSteamCMDArgumentArray(t *testing.T) {
	plan := InstallDSTPlan(domain.ManagedLayout{DST: "/srv/dst with spaces"})

	if plan.Name != "steamcmd" {
		t.Fatalf("Name = %q, want steamcmd", plan.Name)
	}

	wantArgs := []string{
		"+force_install_dir", "/srv/dst with spaces",
		"+login", "anonymous",
		"+app_update", "343050", "validate",
		"+quit",
	}
	if !reflect.DeepEqual(plan.Args, wantArgs) {
		t.Fatalf("Args = %#v, want %#v", plan.Args, wantArgs)
	}
}
