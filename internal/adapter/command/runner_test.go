package command

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestExecRunnerRejectsEmptyCommand(t *testing.T) {
	_, err := ExecRunner{}.Run(context.Background(), "")
	if !errors.Is(err, ErrEmptyCommand) {
		t.Fatalf("Run() error = %v, want ErrEmptyCommand", err)
	}
}

func TestExecRunnerPassesArgumentsWithoutShellExpansion(t *testing.T) {
	result, err := ExecRunner{}.Run(context.Background(), "printf", "%s", "$HOME && echo unsafe")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if result.Stdout != "$HOME && echo unsafe" {
		t.Fatalf("stdout = %q", result.Stdout)
	}
}

func TestExecRunnerRunWithOptionsCapturesOutputAndWritesLogsToFiles(t *testing.T) {
	root := t.TempDir()
	stdoutPath := filepath.Join(root, "stdout.log")
	stderrPath := filepath.Join(root, "stderr.log")

	result, err := ExecRunner{}.RunWithOptions(
		context.Background(),
		StartOptions{StdoutPath: stdoutPath, StderrPath: stderrPath},
		"sh",
		"-c",
		"printf 'out'; printf 'err' >&2",
	)
	if err != nil {
		t.Fatalf("RunWithOptions() error = %v", err)
	}

	if result.Stdout != "out" {
		t.Fatalf("stdout = %q, want out", result.Stdout)
	}
	if result.Stderr != "err" {
		t.Fatalf("stderr = %q, want err", result.Stderr)
	}

	stdout, err := os.ReadFile(stdoutPath)
	if err != nil {
		t.Fatalf("ReadFile(stdout) error = %v", err)
	}
	stderr, err := os.ReadFile(stderrPath)
	if err != nil {
		t.Fatalf("ReadFile(stderr) error = %v", err)
	}
	if !bytes.Equal(stdout, []byte("out")) {
		t.Fatalf("stdout log = %q, want out", string(stdout))
	}
	if !bytes.Equal(stderr, []byte("err")) {
		t.Fatalf("stderr log = %q, want err", string(stderr))
	}
}

func TestExecRunnerStartRejectsEmptyCommand(t *testing.T) {
	_, err := ExecRunner{}.Start(context.Background(), "")
	if !errors.Is(err, ErrEmptyCommand) {
		t.Fatalf("Start() error = %v, want ErrEmptyCommand", err)
	}
}

func TestExecRunnerStartPassesArgumentsWithoutShellExpansion(t *testing.T) {
	process, err := ExecRunner{}.Start(context.Background(), "sleep", "0")
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	if process.PID() == 0 {
		t.Fatalf("PID() = %d, want non-zero", process.PID())
	}
	if err := process.Wait(); err != nil {
		t.Fatalf("Wait() error = %v", err)
	}
}

func TestExecRunnerStartWithOptionsWritesLogsToFiles(t *testing.T) {
	root := t.TempDir()
	stdoutPath := filepath.Join(root, "stdout.log")
	stderrPath := filepath.Join(root, "stderr.log")

	process, err := ExecRunner{}.StartWithOptions(
		context.Background(),
		StartOptions{StdoutPath: stdoutPath, StderrPath: stderrPath},
		"sh",
		"-c",
		"printf 'out'; printf 'err' >&2",
	)
	if err != nil {
		t.Fatalf("StartWithOptions() error = %v", err)
	}
	if err := process.Wait(); err != nil {
		t.Fatalf("Wait() error = %v", err)
	}

	stdout, err := os.ReadFile(stdoutPath)
	if err != nil {
		t.Fatalf("ReadFile(stdout) error = %v", err)
	}
	stderr, err := os.ReadFile(stderrPath)
	if err != nil {
		t.Fatalf("ReadFile(stderr) error = %v", err)
	}
	if !bytes.Equal(stdout, []byte("out")) {
		t.Fatalf("stdout = %q, want out", string(stdout))
	}
	if !bytes.Equal(stderr, []byte("err")) {
		t.Fatalf("stderr = %q, want err", string(stderr))
	}
}
