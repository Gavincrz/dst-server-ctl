package dstconfig

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"dst-server-ctl/internal/adapter/paths"
	"dst-server-ctl/internal/domain"
)

func TestWriterWritesClusterAndShardConfigFiles(t *testing.T) {
	layout := paths.ManagedLayout(t.TempDir())
	writer := NewWriter(layout)
	config := domain.ClusterConfig{
		ClusterName:        "Managed DST",
		ClusterDescription: "Test cluster",
		ClusterPassword:    "secret",
		ClusterIntention:   "cooperative",
		GameMode:           "survival",
		MaxPlayers:         8,
		Language:           "en",
		PVP:                true,
		PauseWhenEmpty:     false,
		OfflineCluster:     false,
		LANOnlyCluster:     true,
		TickRate:           30,
		ConsoleEnabled:     true,
		BindIP:             "0.0.0.0",
		MasterPort:         12000,
		ClusterKey:         "cluster-abc",
		Shards: []domain.ShardConfig{
			{Name: domain.ShardMaster, Enabled: true, ServerPort: 11000, MasterServerPort: 27020, AuthenticationPort: 8768, WorldGenPreset: "SURVIVAL_TOGETHER", WorldGenOverrides: map[string]string{"season_start": "autumn"}},
			{Name: domain.ShardCaves, Enabled: true, ServerPort: 11001, MasterServerPort: 27021, AuthenticationPort: 8769, WorldGenPreset: "DST_CAVE", WorldGenOverrides: map[string]string{"wormattacks": "never"}},
		},
		CreatedAt: time.Date(2026, 4, 24, 10, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 4, 24, 10, 0, 0, 0, time.UTC),
	}

	if err := writer.WriteClusterFiles(context.Background(), config); err != nil {
		t.Fatalf("WriteClusterFiles() error = %v", err)
	}

	clusterINI := readFile(t, filepath.Join(paths.ManagedClusterDir(layout), "cluster.ini"))
	if !strings.Contains(clusterINI, "cluster_name = Managed DST") {
		t.Fatalf("cluster.ini = %q, want cluster_name", clusterINI)
	}
	if !strings.Contains(clusterINI, "cluster_language = en") {
		t.Fatalf("cluster.ini = %q, want cluster_language", clusterINI)
	}
	if !strings.Contains(clusterINI, "cluster_intention = cooperative") {
		t.Fatalf("cluster.ini = %q, want cluster_intention", clusterINI)
	}
	if !strings.Contains(clusterINI, "lan_only_cluster = true") {
		t.Fatalf("cluster.ini = %q, want lan_only_cluster = true", clusterINI)
	}
	if !strings.Contains(clusterINI, "master_port = 12000") {
		t.Fatalf("cluster.ini = %q, want master_port = 12000", clusterINI)
	}
	if !strings.Contains(clusterINI, "shard_enabled = true") {
		t.Fatalf("cluster.ini = %q, want shard_enabled = true", clusterINI)
	}

	masterINI := readFile(t, filepath.Join(paths.ManagedShardDir(layout, domain.ShardMaster), "server.ini"))
	if !strings.Contains(masterINI, "server_port = 11000") {
		t.Fatalf("Master/server.ini = %q, want server_port = 11000", masterINI)
	}
	if !strings.Contains(masterINI, "is_master = true") {
		t.Fatalf("Master/server.ini = %q, want is_master = true", masterINI)
	}

	cavesINI := readFile(t, filepath.Join(paths.ManagedShardDir(layout, domain.ShardCaves), "server.ini"))
	if !strings.Contains(cavesINI, "server_port = 11001") {
		t.Fatalf("Caves/server.ini = %q, want server_port = 11001", cavesINI)
	}
	if !strings.Contains(cavesINI, "name = Caves") {
		t.Fatalf("Caves/server.ini = %q, want name = Caves", cavesINI)
	}
	if !strings.Contains(cavesINI, "master_ip = 127.0.0.1") {
		t.Fatalf("Caves/server.ini = %q, want master_ip", cavesINI)
	}
	if !strings.Contains(cavesINI, "master_server_port = 27021") {
		t.Fatalf("Caves/server.ini = %q, want master_server_port = 27021", cavesINI)
	}

	masterWorldGen := readFile(t, filepath.Join(paths.ManagedShardDir(layout, domain.ShardMaster), "worldgenoverride.lua"))
	if !strings.Contains(masterWorldGen, `preset = "SURVIVAL_TOGETHER"`) {
		t.Fatalf("Master/worldgenoverride.lua = %q, want SURVIVAL_TOGETHER preset", masterWorldGen)
	}
	if !strings.Contains(masterWorldGen, `season_start = "autumn"`) {
		t.Fatalf("Master/worldgenoverride.lua = %q, want season_start override", masterWorldGen)
	}

	cavesWorldGen := readFile(t, filepath.Join(paths.ManagedShardDir(layout, domain.ShardCaves), "worldgenoverride.lua"))
	if !strings.Contains(cavesWorldGen, `preset = "DST_CAVE"`) {
		t.Fatalf("Caves/worldgenoverride.lua = %q, want DST_CAVE preset", cavesWorldGen)
	}
	if !strings.Contains(cavesWorldGen, `wormattacks = "never"`) {
		t.Fatalf("Caves/worldgenoverride.lua = %q, want wormattacks override", cavesWorldGen)
	}
}

func TestWriterRemovesDisabledShardServerINI(t *testing.T) {
	layout := paths.ManagedLayout(t.TempDir())
	writer := NewWriter(layout)
	config := domain.ClusterConfig{
		ClusterName:      "Managed DST",
		ClusterIntention: "cooperative",
		GameMode:         "survival",
		MaxPlayers:       6,
		Language:         "en",
		PauseWhenEmpty:   true,
		TickRate:         15,
		Shards: []domain.ShardConfig{
			{Name: domain.ShardMaster, Enabled: true, ServerPort: 10999, MasterServerPort: 27016, AuthenticationPort: 8766, WorldGenPreset: "SURVIVAL_TOGETHER", WorldGenOverrides: map[string]string{}},
			{Name: domain.ShardCaves, Enabled: true, ServerPort: 11000, MasterServerPort: 27017, AuthenticationPort: 8767, WorldGenPreset: "DST_CAVE", WorldGenOverrides: map[string]string{}},
		},
		CreatedAt: time.Date(2026, 4, 24, 10, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 4, 24, 10, 0, 0, 0, time.UTC),
	}

	if err := writer.WriteClusterFiles(context.Background(), config); err != nil {
		t.Fatalf("initial WriteClusterFiles() error = %v", err)
	}

	config.Shards[1].Enabled = false
	if err := writer.WriteClusterFiles(context.Background(), config); err != nil {
		t.Fatalf("second WriteClusterFiles() error = %v", err)
	}

	clusterINI := readFile(t, filepath.Join(paths.ManagedClusterDir(layout), "cluster.ini"))
	if !strings.Contains(clusterINI, "shard_enabled = false") {
		t.Fatalf("cluster.ini = %q, want shard_enabled = false", clusterINI)
	}

	if _, err := os.Stat(filepath.Join(paths.ManagedShardDir(layout, domain.ShardCaves), "server.ini")); !os.IsNotExist(err) {
		t.Fatalf("Caves/server.ini stat error = %v, want not exist", err)
	}
	if _, err := os.Stat(filepath.Join(paths.ManagedShardDir(layout, domain.ShardCaves), "worldgenoverride.lua")); !os.IsNotExist(err) {
		t.Fatalf("Caves/worldgenoverride.lua stat error = %v, want not exist", err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}

	return string(content)
}
