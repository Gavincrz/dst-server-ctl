package steamcmd

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

func (c *Client) InstallDST(ctx context.Context, layout domain.ManagedLayout) (command.Result, error) {
	plan := InstallDSTPlan(layout)
	return c.runner.Run(ctx, plan.Name, plan.Args...)
}
