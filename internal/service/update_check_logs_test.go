package service

import (
	"context"
	"reflect"
	"testing"

	"dst-server-ctl/internal/domain"
)

func TestUpdateCheckLogServiceGetReturnsRecentLines(t *testing.T) {
	reader := &fakeRuntimeLogReader{lines: []string{"line 1", "line 2"}}
	service := NewUpdateCheckLogService(domain.ManagedLayout{Logs: "/srv/managed/logs"}, reader)

	lines, err := service.Get(context.Background(), 20)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if !reflect.DeepEqual(lines, []string{"line 1", "line 2"}) {
		t.Fatalf("lines = %#v, want %#v", lines, []string{"line 1", "line 2"})
	}
	if reader.path != "/srv/managed/logs/update-check.log" {
		t.Fatalf("path = %q, want /srv/managed/logs/update-check.log", reader.path)
	}
	if reader.maxLines != 20 {
		t.Fatalf("maxLines = %d, want 20", reader.maxLines)
	}
}
