package service

import (
	"context"
	"reflect"
	"testing"
	"time"

	"dst-server-ctl/internal/domain"
)

func TestRuntimeHistoryServiceListReturnsEvents(t *testing.T) {
	now := time.Date(2026, 4, 25, 9, 0, 0, 0, time.UTC)
	events := []domain.RuntimeEvent{
		{ID: 2, Shard: domain.ShardMaster, Kind: domain.RuntimeEventExited, Detail: "exited", CreatedAt: now},
	}
	service := NewRuntimeHistoryService(&fakeRuntimeEventRepository{events: events})

	got, err := service.List(context.Background(), 10)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if !reflect.DeepEqual(got, events) {
		t.Fatalf("events = %#v, want %#v", got, events)
	}
}
