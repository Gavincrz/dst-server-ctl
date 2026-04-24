package command

import (
	"context"
	"errors"
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
