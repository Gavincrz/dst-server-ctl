package service

import (
	"context"
	"fmt"

	"dst-server-ctl/internal/adapter/paths"
	"dst-server-ctl/internal/domain"
)

type RuntimeLogReader interface {
	ReadRecent(ctx context.Context, path string, maxLines int) ([]string, error)
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
	if maxLines <= 0 {
		maxLines = 200
	}
	if maxLines > 500 {
		maxLines = 500
	}

	return s.reader.ReadRecent(ctx, paths.ManagedShardLogPath(s.layout, shard), maxLines)
}
