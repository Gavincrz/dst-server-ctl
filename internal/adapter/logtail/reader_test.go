package logtail

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestReaderReadRecentReturnsTailLines(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "master.log")
	if err := os.WriteFile(path, []byte("1\n2\n3\n4\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	lines, err := Reader{}.ReadRecent(context.Background(), path, 2)
	if err != nil {
		t.Fatalf("ReadRecent() error = %v", err)
	}

	if !reflect.DeepEqual(lines, []string{"3", "4"}) {
		t.Fatalf("lines = %#v, want %#v", lines, []string{"3", "4"})
	}
}

func TestReaderReadRecentReturnsEmptyWhenMissing(t *testing.T) {
	lines, err := Reader{}.ReadRecent(context.Background(), "/missing.log", 50)
	if err != nil {
		t.Fatalf("ReadRecent() error = %v", err)
	}
	if len(lines) != 0 {
		t.Fatalf("lines = %#v, want empty", lines)
	}
}
