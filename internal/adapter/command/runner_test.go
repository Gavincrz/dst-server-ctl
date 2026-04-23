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
