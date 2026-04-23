package paths

import (
	"os"
	"path/filepath"

	"dst-server-ctl/internal/domain"
)

const appDirName = "dst-server-ctl"

func DefaultManagedRoot() string {
	if dataHome := os.Getenv("XDG_DATA_HOME"); dataHome != "" {
		return filepath.Join(dataHome, appDirName)
	}

	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return filepath.Join(".", "data")
	}

	return filepath.Join(home, ".local", "share", appDirName)
}

func ManagedLayout(root string) domain.ManagedLayout {
	return domain.ManagedLayout{
		Root:     root,
		SteamCMD: filepath.Join(root, "steamcmd"),
		DST:      filepath.Join(root, "dst"),
		Clusters: filepath.Join(root, "clusters"),
		Logs:     filepath.Join(root, "logs"),
		State:    filepath.Join(root, "state"),
	}
}

func EnsureManagedLayout(layout domain.ManagedLayout) error {
	dirs := []string{
		layout.Root,
		layout.SteamCMD,
		layout.DST,
		layout.Clusters,
		layout.Logs,
		layout.State,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return err
		}
	}

	return nil
}
