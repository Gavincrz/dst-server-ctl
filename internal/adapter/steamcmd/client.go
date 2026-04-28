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

func (c *Client) InstallSteamCMD(ctx context.Context, layout domain.ManagedLayout, logPath string) (command.Result, error) {
	archivePath := filepath.Join(layout.SteamCMD, "steamcmd_linux.tar.gz")
	options := command.StartOptions{StdoutPath: logPath, StderrPath: logPath}

	downloadResult, err := c.runner.RunWithOptions(
		ctx,
		options,
		"curl",
		"-fsSL",
		steamCMDDownloadURL,
		"-o",
		archivePath,
	)
	if err != nil {
		return downloadResult, fmt.Errorf("download steamcmd: %w", err)
	}

	extractResult, err := c.runner.RunWithOptions(
		ctx,
		options,
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

func (c *Client) InstallDST(ctx context.Context, layout domain.ManagedLayout, logPath string) (command.Result, error) {
	plan := InstallDSTPlan(layout)
	return c.runner.RunWithOptions(ctx, command.StartOptions{StdoutPath: logPath, StderrPath: logPath}, plan.Name, plan.Args...)
}

func (c *Client) UpdateDST(ctx context.Context, layout domain.ManagedLayout, logPath string) (command.Result, error) {
	return c.InstallDST(ctx, layout, logPath)
}

func (c *Client) LocalVersion(_ context.Context, layout domain.ManagedLayout) (string, error) {
	version, err := ReadLocalVersion(layout)
	if err != nil {
		return "", fmt.Errorf("read local version: %w", err)
	}
	return version, nil
}

func (c *Client) RemoteVersion(ctx context.Context, layout domain.ManagedLayout) (string, command.Result, error) {
	plan := RemoteVersionPlan(layout)
	result, err := c.runner.Run(ctx, plan.Name, plan.Args...)
	if err != nil {
		return "", result, err
	}

	version, err := ParseRemoteVersion(result.Stdout)
	if err != nil {
		return "", result, err
	}
	return version, result, nil
}
