package dstserver

import (
	"context"

	"dst-server-ctl/internal/adapter/command"
	"dst-server-ctl/internal/domain"
)

type Client struct {
	runner command.Runner
}

func NewClient(runner command.Runner) *Client {
	return &Client{runner: runner}
}

func (c *Client) StartShard(ctx context.Context, layout domain.ManagedLayout, shard domain.ShardName) (command.Process, error) {
	plan := StartShardPlan(layout, shard)
	return c.runner.Start(ctx, plan.Name, plan.Args...)
}
