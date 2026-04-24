package paths

import (
	"path/filepath"

	"dst-server-ctl/internal/domain"
)

const managedClusterDirName = "primary"

func ManagedClusterDir(layout domain.ManagedLayout) string {
	return filepath.Join(layout.Clusters, managedClusterDirName)
}

func ManagedShardDir(layout domain.ManagedLayout, shard domain.ShardName) string {
	return filepath.Join(ManagedClusterDir(layout), string(shard))
}
