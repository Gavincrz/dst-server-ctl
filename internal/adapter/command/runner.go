package command

import (
	"context"
	"errors"
	"os/exec"
)

var ErrEmptyCommand = errors.New("empty command")

type Result struct {
	Stdout string
	Stderr string
}

type Runner interface {
	Run(ctx context.Context, name string, args ...string) (Result, error)
}

type ExecRunner struct{}

func (ExecRunner) Run(ctx context.Context, name string, args ...string) (Result, error) {
	if name == "" {
		return Result{}, ErrEmptyCommand
	}

	cmd := exec.CommandContext(ctx, name, args...)

	stdout, err := cmd.Output()
	if err == nil {
		return Result{Stdout: string(stdout)}, nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return Result{
			Stdout: string(stdout),
			Stderr: string(exitErr.Stderr),
		}, err
	}

	return Result{Stdout: string(stdout)}, err
}
