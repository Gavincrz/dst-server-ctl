package dstconfig

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"dst-server-ctl/internal/adapter/paths"
	"dst-server-ctl/internal/domain"
)

const (
	loopbackAddress = "127.0.0.1"
)

type Writer struct {
	layout domain.ManagedLayout
}

func NewWriter(layout domain.ManagedLayout) *Writer {
	return &Writer{layout: layout}
}

func (w *Writer) WriteClusterFiles(_ context.Context, config domain.ClusterConfig) error {
	clusterDir := paths.ManagedClusterDir(w.layout)
	if err := os.MkdirAll(clusterDir, 0o700); err != nil {
		return fmt.Errorf("create cluster directory: %w", err)
	}

	if err := os.WriteFile(filepath.Join(clusterDir, "cluster.ini"), []byte(renderClusterINI(config)), 0o600); err != nil {
		return fmt.Errorf("write cluster.ini: %w", err)
	}

	for _, shardName := range []domain.ShardName{domain.ShardMaster, domain.ShardCaves} {
		enabled := shardEnabled(config.Shards, shardName)
		shardDir := paths.ManagedShardDir(w.layout, shardName)
		serverPath := filepath.Join(shardDir, "server.ini")

		if !enabled {
			if err := os.Remove(serverPath); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("remove %s server.ini: %w", shardName, err)
			}
			continue
		}

		if err := os.MkdirAll(shardDir, 0o700); err != nil {
			return fmt.Errorf("create %s shard directory: %w", shardName, err)
		}
		if err := os.WriteFile(serverPath, []byte(renderServerINI(config, shardName)), 0o600); err != nil {
			return fmt.Errorf("write %s server.ini: %w", shardName, err)
		}
	}

	return nil
}

func renderClusterINI(config domain.ClusterConfig) string {
	enabledShards := countEnabledShards(config.Shards)

	var builder strings.Builder
	builder.WriteString("[GAMEPLAY]\n")
	builder.WriteString(fmt.Sprintf("game_mode = %s\n", config.GameMode))
	builder.WriteString(fmt.Sprintf("max_players = %d\n", config.MaxPlayers))
	builder.WriteString(fmt.Sprintf("pvp = %s\n", formatBool(config.PVP)))
	builder.WriteString(fmt.Sprintf("pause_when_empty = %s\n\n", formatBool(config.PauseWhenEmpty)))

	builder.WriteString("[NETWORK]\n")
	builder.WriteString(fmt.Sprintf("cluster_name = %s\n", config.ClusterName))
	builder.WriteString(fmt.Sprintf("cluster_description = %s\n", config.ClusterDescription))
	builder.WriteString(fmt.Sprintf("cluster_password = %s\n", config.ClusterPassword))
	builder.WriteString(fmt.Sprintf("cluster_intention = %s\n", config.ClusterIntention))
	builder.WriteString(fmt.Sprintf("cluster_language = %s\n", config.Language))
	builder.WriteString(fmt.Sprintf("offline_cluster = %s\n", formatBool(config.OfflineCluster)))
	builder.WriteString(fmt.Sprintf("lan_only_cluster = %s\n", formatBool(config.LANOnlyCluster)))
	builder.WriteString(fmt.Sprintf("tick_rate = %d\n\n", config.TickRate))

	builder.WriteString("[MISC]\n")
	builder.WriteString(fmt.Sprintf("console_enabled = %s\n\n", formatBool(config.ConsoleEnabled)))

	builder.WriteString("[SHARD]\n")
	builder.WriteString(fmt.Sprintf("shard_enabled = %s\n", formatBool(enabledShards > 1)))
	if enabledShards > 1 {
		builder.WriteString(fmt.Sprintf("bind_ip = %s\n", config.BindIP))
		builder.WriteString(fmt.Sprintf("master_port = %d\n", config.MasterPort))
		builder.WriteString(fmt.Sprintf("cluster_key = %s\n", config.ClusterKey))
	}

	return builder.String()
}

func renderServerINI(config domain.ClusterConfig, shardName domain.ShardName) string {
	var builder strings.Builder
	shard, ok := findShard(config.Shards, shardName)
	if !ok {
		return ""
	}

	builder.WriteString("[NETWORK]\n")
	builder.WriteString(fmt.Sprintf("server_port = %d\n", shard.ServerPort))
	if countEnabledShards(config.Shards) > 1 && shardName != domain.ShardMaster {
		builder.WriteString(fmt.Sprintf("master_ip = %s\n", loopbackAddress))
		builder.WriteString(fmt.Sprintf("master_port = %d\n", config.MasterPort))
	}
	builder.WriteString("\n[STEAM]\n")
	builder.WriteString(fmt.Sprintf("master_server_port = %d\n", shard.MasterServerPort))
	builder.WriteString(fmt.Sprintf("authentication_port = %d\n", shard.AuthenticationPort))

	if countEnabledShards(config.Shards) > 1 {
		builder.WriteString("\n[SHARD]\n")
		if shardName == domain.ShardMaster {
			builder.WriteString("is_master = true\n")
		} else {
			builder.WriteString("is_master = false\n")
			builder.WriteString(fmt.Sprintf("name = %s\n", shardName))
		}
	}

	return builder.String()
}

func formatBool(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func countEnabledShards(shards []domain.ShardConfig) int {
	count := 0
	for _, shard := range shards {
		if shard.Enabled {
			count++
		}
	}
	return count
}

func shardEnabled(shards []domain.ShardConfig, name domain.ShardName) bool {
	for _, shard := range shards {
		if shard.Name == name {
			return shard.Enabled
		}
	}

	return false
}

func findShard(shards []domain.ShardConfig, name domain.ShardName) (domain.ShardConfig, bool) {
	for _, shard := range shards {
		if shard.Name == name {
			return shard, true
		}
	}

	return domain.ShardConfig{}, false
}
