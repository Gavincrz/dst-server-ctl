package paths

import (
	"path/filepath"
	"strings"

	"dst-server-ctl/internal/domain"
)

func ManagedShardLogPath(layout domain.ManagedLayout, shard domain.ShardName) string {
	return filepath.Join(layout.Logs, strings.ToLower(string(shard))+".log")
}
