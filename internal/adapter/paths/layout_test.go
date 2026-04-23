package paths

import (
	"path/filepath"
	"testing"
)

func TestDefaultManagedRootUsesXDGDataHome(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", "/tmp/example-data")

	got := DefaultManagedRoot()
	want := filepath.Join("/tmp/example-data", "dst-server-ctl")

	if got != want {
		t.Fatalf("DefaultManagedRoot() = %q, want %q", got, want)
	}
}

func TestManagedLayoutUsesStableSubdirectories(t *testing.T) {
	layout := ManagedLayout("/srv/example")

	if layout.SteamCMD != "/srv/example/steamcmd" {
		t.Fatalf("SteamCMD path = %q", layout.SteamCMD)
	}
	if layout.DST != "/srv/example/dst" {
		t.Fatalf("DST path = %q", layout.DST)
	}
	if layout.Clusters != "/srv/example/clusters" {
		t.Fatalf("Clusters path = %q", layout.Clusters)
	}
	if layout.Logs != "/srv/example/logs" {
		t.Fatalf("Logs path = %q", layout.Logs)
	}
	if layout.State != "/srv/example/state" {
		t.Fatalf("State path = %q", layout.State)
	}
}
