package command

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
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

type StartOptions struct {
	StdoutPath string
	StderrPath string
}

type Runner interface {
	Run(ctx context.Context, name string, args ...string) (Result, error)
	RunWithOptions(ctx context.Context, options StartOptions, name string, args ...string) (Result, error)
	Start(ctx context.Context, name string, args ...string) (Process, error)
	StartWithOptions(ctx context.Context, options StartOptions, name string, args ...string) (Process, error)
}

type ExecRunner struct{}

func (ExecRunner) Run(ctx context.Context, name string, args ...string) (Result, error) {
	return ExecRunner{}.RunWithOptions(ctx, StartOptions{}, name, args...)
}

func (ExecRunner) RunWithOptions(ctx context.Context, options StartOptions, name string, args ...string) (Result, error) {
	if name == "" {
		return Result{}, ErrEmptyCommand
	}

	cmd := exec.CommandContext(ctx, name, args...)
	stdoutWriter, stdoutBuffer, stdoutCloser, err := prepareOutputWriter(options.StdoutPath)
	if err != nil {
		return Result{}, err
	}
	if stdoutCloser != nil {
		defer stdoutCloser.Close()
	}
	stderrWriter, stderrBuffer, stderrCloser, err := prepareOutputWriter(options.StderrPath)
	if err != nil {
		return Result{}, err
	}
	if stderrCloser != nil {
		defer stderrCloser.Close()
	}
	cmd.Stdout = stdoutWriter
	cmd.Stderr = stderrWriter

	err = cmd.Run()
	result := Result{
		Stdout: stdoutBuffer.String(),
		Stderr: stderrBuffer.String(),
	}
	if err == nil {
		return result, nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return result, err
	}

	return result, err
}

func (ExecRunner) Start(ctx context.Context, name string, args ...string) (Process, error) {
	return ExecRunner{}.StartWithOptions(ctx, StartOptions{}, name, args...)
}

func (ExecRunner) StartWithOptions(ctx context.Context, options StartOptions, name string, args ...string) (Process, error) {
	if name == "" {
		return nil, ErrEmptyCommand
	}

	cmd := exec.CommandContext(ctx, name, args...)
	if err := applyStartOptions(cmd, options); err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return execProcess{cmd: cmd}, nil
}

func applyStartOptions(cmd *exec.Cmd, options StartOptions) error {
	if options.StdoutPath != "" {
		file, err := openLogFile(options.StdoutPath)
		if err != nil {
			return err
		}
		cmd.Stdout = file
	}
	if options.StderrPath != "" {
		file, err := openLogFile(options.StderrPath)
		if err != nil {
			return err
		}
		cmd.Stderr = file
	}
	return nil
}

func prepareOutputWriter(path string) (io.Writer, *bytes.Buffer, io.Closer, error) {
	buffer := &bytes.Buffer{}
	if path == "" {
		return buffer, buffer, nil, nil
	}

	file, err := openLogFile(path)
	if err != nil {
		return nil, nil, nil, err
	}

	return io.MultiWriter(buffer, file), buffer, file, nil
}

func openLogFile(path string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}
	return os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
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
