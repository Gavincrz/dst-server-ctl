package service

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"dst-server-ctl/internal/domain"
)

func TestRuntimeLogServiceGetReturnsRecentLines(t *testing.T) {
	reader := &fakeRuntimeLogReader{lines: []string{"a", "b"}}
	service := NewRuntimeLogService(domain.ManagedLayout{Logs: "/srv/managed/logs"}, reader)

	lines, err := service.Get(context.Background(), domain.ShardMaster, 20)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if !reflect.DeepEqual(lines, []string{"a", "b"}) {
		t.Fatalf("lines = %#v, want %#v", lines, []string{"a", "b"})
	}
	if reader.path != "/srv/managed/logs/master.log" {
		t.Fatalf("path = %q, want /srv/managed/logs/master.log", reader.path)
	}
	if reader.maxLines != 20 {
		t.Fatalf("maxLines = %d, want 20", reader.maxLines)
	}
}

func TestRuntimeLogServiceGetRejectsUnsupportedShard(t *testing.T) {
	service := NewRuntimeLogService(domain.ManagedLayout{}, &fakeRuntimeLogReader{})

	_, err := service.Get(context.Background(), "Ruins", 20)
	if !errors.Is(err, domain.ErrInvalidShard) {
		t.Fatalf("Get() error = %v, want %v", err, domain.ErrInvalidShard)
	}
}

type fakeRuntimeLogReader struct {
	path     string
	maxLines int
	lines    []string
	err      error
}

func (r *fakeRuntimeLogReader) ReadRecent(_ context.Context, path string, maxLines int) ([]string, error) {
	r.path = path
	r.maxLines = maxLines
	if r.err != nil {
		return nil, r.err
	}
	return r.lines, nil
}
