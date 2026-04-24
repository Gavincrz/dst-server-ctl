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
	loopbackAddress   = "127.0.0.1"
	defaultMasterPort = 10888
	defaultClusterKey = "dst-server-ctl"
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
	builder.WriteString(fmt.Sprintf("cluster_language = %s\n\n", config.Language))

	builder.WriteString("[SHARD]\n")
	builder.WriteString(fmt.Sprintf("shard_enabled = %s\n", formatBool(enabledShards > 1)))
	if enabledShards > 1 {
		builder.WriteString(fmt.Sprintf("bind_ip = %s\n", loopbackAddress))
		builder.WriteString(fmt.Sprintf("master_port = %d\n", defaultMasterPort))
		builder.WriteString(fmt.Sprintf("cluster_key = %s\n", defaultClusterKey))
	}

	return builder.String()
}

func renderServerINI(config domain.ClusterConfig, shardName domain.ShardName) string {
	var builder strings.Builder

	if countEnabledShards(config.Shards) > 1 {
		builder.WriteString("[SHARD]\n")
		if shardName == domain.ShardMaster {
			builder.WriteString("is_master = true\n")
		} else {
			builder.WriteString("is_master = false\n")
			builder.WriteString(fmt.Sprintf("name = %s\n", shardName))
			builder.WriteString("\n[NETWORK]\n")
			builder.WriteString(fmt.Sprintf("master_ip = %s\n", loopbackAddress))
			builder.WriteString(fmt.Sprintf("master_port = %d\n", defaultMasterPort))
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
