package logtail

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"dst-server-ctl/internal/domain"
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

func TestReaderOpenStreamReadsOnlyAppendedLines(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "master.log")
	if err := os.WriteFile(path, []byte("1\n2\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	stream, err := Reader{}.OpenStream(path, 10)
	if err != nil {
		t.Fatalf("OpenStream() error = %v", err)
	}
	defer stream.Close()

	if got, want := stream.Snapshot(), []string{"1", "2"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Snapshot() = %#v, want %#v", got, want)
	}

	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatalf("OpenFile() error = %v", err)
	}
	if _, err := file.WriteString("3\n4\n"); err != nil {
		t.Fatalf("WriteString() error = %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	update, err := stream.Poll(context.Background())
	if err != nil {
		t.Fatalf("Poll() error = %v", err)
	}

	if got, want := update, (domain.LogStreamUpdate{Lines: []string{"3", "4"}, Changed: true}); !reflect.DeepEqual(got, want) {
		t.Fatalf("Poll() = %#v, want %#v", got, want)
	}
}

func TestReaderOpenStreamReturnsSnapshotAfterReplace(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "master.log")
	if err := os.WriteFile(path, []byte("1\n2\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	stream, err := Reader{}.OpenStream(path, 10)
	if err != nil {
		t.Fatalf("OpenStream() error = %v", err)
	}
	defer stream.Close()

	if err := os.Rename(path, filepath.Join(root, "master.log.1")); err != nil {
		t.Fatalf("Rename() error = %v", err)
	}
	if err := os.WriteFile(path, []byte("7\n8\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	update, err := stream.Poll(context.Background())
	if err != nil {
		t.Fatalf("Poll() error = %v", err)
	}

	if got, want := update, (domain.LogStreamUpdate{Lines: []string{"7", "8"}, Reset: true, Changed: true}); !reflect.DeepEqual(got, want) {
		t.Fatalf("Poll() = %#v, want %#v", got, want)
	}
}
