package service

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"dst-server-ctl/internal/domain"
)

func TestUpdateTaskLogServiceGetReturnsRecentLines(t *testing.T) {
	reader := &fakeRuntimeLogReader{lines: []string{"line 1", "line 2"}}
	service := NewUpdateTaskLogService(domain.ManagedLayout{Logs: "/srv/managed/logs"}, reader)

	lines, err := service.Get(context.Background(), "task-1", 20)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if !reflect.DeepEqual(lines, []string{"line 1", "line 2"}) {
		t.Fatalf("lines = %#v, want %#v", lines, []string{"line 1", "line 2"})
	}
	if reader.path != "/srv/managed/logs/update-task-1.log" {
		t.Fatalf("path = %q, want /srv/managed/logs/update-task-1.log", reader.path)
	}
	if reader.maxLines != 20 {
		t.Fatalf("maxLines = %d, want 20", reader.maxLines)
	}
}

func TestUpdateTaskLogServiceGetRejectsEmptyTaskID(t *testing.T) {
	service := NewUpdateTaskLogService(domain.ManagedLayout{}, &fakeRuntimeLogReader{})

	_, err := service.Get(context.Background(), "", 20)
	if !errors.Is(err, domain.ErrTaskNotFound) {
		t.Fatalf("Get() error = %v, want %v", err, domain.ErrTaskNotFound)
	}
}
