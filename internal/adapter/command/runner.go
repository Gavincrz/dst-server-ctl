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

type Process interface {
	PID() int
	Wait() error
	Kill() error
}

type Runner interface {
	Run(ctx context.Context, name string, args ...string) (Result, error)
	Start(ctx context.Context, name string, args ...string) (Process, error)
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

func (ExecRunner) Start(ctx context.Context, name string, args ...string) (Process, error) {
	if name == "" {
		return nil, ErrEmptyCommand
	}

	cmd := exec.CommandContext(ctx, name, args...)
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return execProcess{cmd: cmd}, nil
}

type execProcess struct {
	cmd *exec.Cmd
}

func (p execProcess) PID() int {
	if p.cmd == nil || p.cmd.Process == nil {
		return 0
	}
	return p.cmd.Process.Pid
}

func (p execProcess) Wait() error {
	if p.cmd == nil {
		return nil
	}
	return p.cmd.Wait()
}

func (p execProcess) Kill() error {
	if p.cmd == nil || p.cmd.Process == nil {
		return nil
	}
	return p.cmd.Process.Kill()
}
