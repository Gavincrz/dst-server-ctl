package service

import (
	"testing"
	"time"

	"dst-server-ctl/internal/domain"
)

func TestInstallPlannerPlansMissingSteamCMDAndDST(t *testing.T) {
	plan := NewInstallPlanner().Plan(domain.InstallationState{})

	if len(plan.Steps) != 2 {
		t.Fatalf("steps = %d, want 2", len(plan.Steps))
	}
	if plan.Steps[0].Type != domain.TaskTypeInstallSteamCMD {
		t.Fatalf("first step = %q, want %q", plan.Steps[0].Type, domain.TaskTypeInstallSteamCMD)
	}
	if plan.Steps[1].Type != domain.TaskTypeInstallDST {
		t.Fatalf("second step = %q, want %q", plan.Steps[1].Type, domain.TaskTypeInstallDST)
	}
}

func TestInstallPlannerSkipsCompletedSteps(t *testing.T) {
	now := time.Now()
	plan := NewInstallPlanner().Plan(domain.InstallationState{
		SteamCMDInstalledAt: &now,
		DSTInstalledAt:      &now,
	})

	if len(plan.Steps) != 0 {
		t.Fatalf("steps = %d, want 0", len(plan.Steps))
	}
}
