package service

import (
	"testing"
	"time"

	"dst-server-ctl/internal/domain"
)

func TestStatusServiceReturnsRunningStatusAndStartupTime(t *testing.T) {
	service := NewStatusService("test-version")
	startedAt := time.Date(2026, 4, 24, 13, 0, 0, 0, time.UTC)
	service.startedAt = startedAt

	status := service.Status()

	if status.Version != "test-version" {
		t.Fatalf("Version = %q, want test-version", status.Version)
	}
	if status.Status != domain.ServerStatusRunning {
		t.Fatalf("Status = %q, want %q", status.Status, domain.ServerStatusRunning)
	}
	if status.StartedAt == nil {
		t.Fatal("StartedAt = nil, want populated")
	}
	if !status.StartedAt.Equal(startedAt) {
		t.Fatalf("StartedAt = %v, want %v", status.StartedAt, startedAt)
	}
}
