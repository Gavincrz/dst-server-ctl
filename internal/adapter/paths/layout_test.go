package paths

import (
	"os"
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

func TestEnsureManagedLayoutCreatesStableSubdirectories(t *testing.T) {
	root := t.TempDir()
	layout := ManagedLayout(filepath.Join(root, "managed"))

	if err := EnsureManagedLayout(layout); err != nil {
		t.Fatalf("EnsureManagedLayout() error = %v", err)
	}

	for _, dir := range []string{
		layout.Root,
		layout.SteamCMD,
		layout.DST,
		layout.Clusters,
		layout.Logs,
		layout.State,
	} {
		info, err := os.Stat(dir)
		if err != nil {
			t.Fatalf("Stat(%q) error = %v", dir, err)
		}
		if !info.IsDir() {
			t.Fatalf("%q is not a directory", dir)
		}
	}
}

func TestManagedClusterPathsUseStableManagedDirectory(t *testing.T) {
	layout := ManagedLayout("/srv/example")

	if got := ManagedClusterDir(layout); got != "/srv/example/clusters/primary" {
		t.Fatalf("ManagedClusterDir() = %q, want %q", got, "/srv/example/clusters/primary")
	}
	if got := ManagedShardDir(layout, "Master"); got != "/srv/example/clusters/primary/Master" {
		t.Fatalf("ManagedShardDir(Master) = %q, want %q", got, "/srv/example/clusters/primary/Master")
	}
	if got := ManagedShardDir(layout, "Caves"); got != "/srv/example/clusters/primary/Caves" {
		t.Fatalf("ManagedShardDir(Caves) = %q, want %q", got, "/srv/example/clusters/primary/Caves")
	}
}
