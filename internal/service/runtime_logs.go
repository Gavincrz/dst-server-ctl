package service

import (
	"context"
	"fmt"

	"dst-server-ctl/internal/adapter/paths"
	"dst-server-ctl/internal/domain"
)

type RuntimeLogReader interface {
	ReadRecent(ctx context.Context, path string, maxLines int) ([]string, error)
	OpenStream(path string, maxLines int) (domain.LogStream, error)
}

type RuntimeLogService struct {
	layout domain.ManagedLayout
	reader RuntimeLogReader
}

func NewRuntimeLogService(layout domain.ManagedLayout, reader RuntimeLogReader) *RuntimeLogService {
	return &RuntimeLogService{
		layout: layout,
		reader: reader,
	}
}

func (s *RuntimeLogService) Get(ctx context.Context, shard domain.ShardName, maxLines int) ([]string, error) {
	if shard != domain.ShardMaster && shard != domain.ShardCaves {
		return nil, fmt.Errorf("%w: unsupported shard %q", domain.ErrInvalidShard, shard)
	}
	maxLines = normalizeLogLineLimit(maxLines)

	return s.reader.ReadRecent(ctx, paths.ManagedShardLogPath(s.layout, shard), maxLines)
}

func (s *RuntimeLogService) Stream(shard domain.ShardName, maxLines int) (domain.LogStream, error) {
	if shard != domain.ShardMaster && shard != domain.ShardCaves {
		return nil, fmt.Errorf("%w: unsupported shard %q", domain.ErrInvalidShard, shard)
	}

	return s.reader.OpenStream(paths.ManagedShardLogPath(s.layout, shard), normalizeLogLineLimit(maxLines))
}

func normalizeLogLineLimit(maxLines int) int {
	if maxLines <= 0 {
		return 200
	}
	if maxLines > 500 {
		return 500
	}
	return maxLines
}
