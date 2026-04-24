package dstserver

import (
	"path/filepath"

	"dst-server-ctl/internal/domain"
)

const managedClusterName = "primary"

type CommandPlan struct {
	Name string
	Args []string
}

func StartShardPlan(layout domain.ManagedLayout, shard domain.ShardName) CommandPlan {
	return CommandPlan{
		Name: filepath.Join(layout.DST, "bin64", "dontstarve_dedicated_server_nullrenderer"),
		Args: []string{
			"-persistent_storage_root", layout.Root,
			"-conf_dir", "clusters",
			"-cluster", managedClusterName,
			"-console",
			"-shard", string(shard),
		},
	}
}
