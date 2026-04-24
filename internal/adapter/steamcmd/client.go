package steamcmd

import (
	"context"
	"fmt"
	"path/filepath"

	"dst-server-ctl/internal/adapter/command"
	"dst-server-ctl/internal/domain"
)

const steamCMDDownloadURL = "https://steamcdn-a.akamaihd.net/client/installer/steamcmd_linux.tar.gz"

type Client struct {
	runner command.Runner
}

func NewClient(runner command.Runner) *Client {
	return &Client{runner: runner}
}

func (c *Client) InstallSteamCMD(ctx context.Context, layout domain.ManagedLayout) (command.Result, error) {
	archivePath := filepath.Join(layout.SteamCMD, "steamcmd_linux.tar.gz")

	downloadResult, err := c.runner.Run(
		ctx,
		"curl",
		"-fsSL",
		steamCMDDownloadURL,
		"-o",
		archivePath,
	)
	if err != nil {
		return downloadResult, fmt.Errorf("download steamcmd: %w", err)
	}

	extractResult, err := c.runner.Run(
		ctx,
		"tar",
		"-xzf",
		archivePath,
		"-C",
		layout.SteamCMD,
	)
	if err != nil {
		return extractResult, fmt.Errorf("extract steamcmd: %w", err)
	}

	return extractResult, nil
}

func (c *Client) InstallDST(ctx context.Context, layout domain.ManagedLayout) (command.Result, error) {
	plan := InstallDSTPlan(layout)
	return c.runner.Run(ctx, plan.Name, plan.Args...)
}
