package steamcmd

import (
	"reflect"
	"testing"

	"dst-server-ctl/internal/domain"
)

func TestInstallDSTPlanUsesSteamCMDArgumentArray(t *testing.T) {
	plan := InstallDSTPlan(domain.ManagedLayout{
		SteamCMD: "/srv/managed/steamcmd",
		DST:      "/srv/dst with spaces",
	})

	if plan.Name != "/srv/managed/steamcmd/steamcmd.sh" {
		t.Fatalf("Name = %q, want /srv/managed/steamcmd/steamcmd.sh", plan.Name)
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

func TestRemoteVersionPlanUsesSteamCMDArgumentArray(t *testing.T) {
	plan := RemoteVersionPlan(domain.ManagedLayout{
		SteamCMD: "/srv/managed/steamcmd",
	})

	if plan.Name != "/srv/managed/steamcmd/steamcmd.sh" {
		t.Fatalf("Name = %q, want /srv/managed/steamcmd/steamcmd.sh", plan.Name)
	}

	wantArgs := []string{
		"+login", "anonymous",
		"+app_info_update", "1",
		"+app_info_print", "343050",
		"+quit",
	}
	if !reflect.DeepEqual(plan.Args, wantArgs) {
		t.Fatalf("Args = %#v, want %#v", plan.Args, wantArgs)
	}
}

func TestParseVersions(t *testing.T) {
	localVersion, err := ParseLocalVersion(`"AppState" { "buildid" "123456" }`)
	if err != nil {
		t.Fatalf("ParseLocalVersion() error = %v", err)
	}
	if localVersion != "123456" {
		t.Fatalf("localVersion = %q, want 123456", localVersion)
	}

	remoteVersion, err := ParseRemoteVersion(`"branches" { "public" { "buildid" "654321" } }`)
	if err != nil {
		t.Fatalf("ParseRemoteVersion() error = %v", err)
	}
	if remoteVersion != "654321" {
		t.Fatalf("remoteVersion = %q, want 654321", remoteVersion)
	}
}
