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
		GameMode:           "survival",
		MaxPlayers:         8,
		Language:           "en",
		PVP:                true,
		PauseWhenEmpty:     false,
		Shards: []domain.ShardConfig{
			{Name: domain.ShardMaster, Enabled: true},
			{Name: domain.ShardCaves, Enabled: true},
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
	if !strings.Contains(clusterINI, "shard_enabled = true") {
		t.Fatalf("cluster.ini = %q, want shard_enabled = true", clusterINI)
	}

	masterINI := readFile(t, filepath.Join(paths.ManagedShardDir(layout, domain.ShardMaster), "server.ini"))
	if !strings.Contains(masterINI, "is_master = true") {
		t.Fatalf("Master/server.ini = %q, want is_master = true", masterINI)
	}

	cavesINI := readFile(t, filepath.Join(paths.ManagedShardDir(layout, domain.ShardCaves), "server.ini"))
	if !strings.Contains(cavesINI, "name = Caves") {
		t.Fatalf("Caves/server.ini = %q, want name = Caves", cavesINI)
	}
	if !strings.Contains(cavesINI, "master_ip = 127.0.0.1") {
		t.Fatalf("Caves/server.ini = %q, want master_ip", cavesINI)
	}
}

func TestWriterRemovesDisabledShardServerINI(t *testing.T) {
	layout := paths.ManagedLayout(t.TempDir())
	writer := NewWriter(layout)
	config := domain.ClusterConfig{
		ClusterName:    "Managed DST",
		GameMode:       "survival",
		MaxPlayers:     6,
		Language:       "en",
		PauseWhenEmpty: true,
		Shards: []domain.ShardConfig{
			{Name: domain.ShardMaster, Enabled: true},
			{Name: domain.ShardCaves, Enabled: true},
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
}

func readFile(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}

	return string(content)
}
