package logtail

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"

	"dst-server-ctl/internal/domain"
)

func (Reader) OpenStream(path string, maxLines int) (domain.LogStream, error) {
	stream := &Stream{
		path:     path,
		maxLines: maxLines,
	}

	if err := stream.reset(); err != nil {
		return nil, err
	}

	return stream, nil
}

type Stream struct {
	path     string
	maxLines int

	file     *os.File
	fileInfo os.FileInfo
	offset   int64
	pending  []byte
	lines    []string
}

func (s *Stream) Snapshot() []string {
	return append([]string(nil), s.lines...)
}

func (s *Stream) Poll(_ context.Context) (domain.LogStreamUpdate, error) {
	info, err := os.Stat(s.path)
	if errors.Is(err, os.ErrNotExist) {
		if s.file == nil && s.offset == 0 && len(s.lines) == 0 && len(s.pending) == 0 {
			return domain.LogStreamUpdate{}, nil
		}

		s.closeFile()
		s.fileInfo = nil
		s.offset = 0
		s.pending = nil
		s.lines = nil
		return domain.LogStreamUpdate{
			Reset:   true,
			Changed: true,
			Lines:   []string{},
		}, nil
	}
	if err != nil {
		return domain.LogStreamUpdate{}, err
	}

	if s.file == nil || s.fileInfo == nil || !os.SameFile(s.fileInfo, info) || info.Size() < s.offset || (info.Size() == s.offset && !info.ModTime().Equal(s.fileInfo.ModTime())) {
		previous := s.Snapshot()
		if err := s.reset(); err != nil {
			return domain.LogStreamUpdate{}, err
		}

		if slicesEqual(previous, s.lines) {
			return domain.LogStreamUpdate{}, nil
		}
		return domain.LogStreamUpdate{
			Reset:   true,
			Changed: true,
			Lines:   s.Snapshot(),
		}, nil
	}

	if info.Size() == s.offset {
		return domain.LogStreamUpdate{}, nil
	}

	if _, err := s.file.Seek(s.offset, io.SeekStart); err != nil {
		return domain.LogStreamUpdate{}, err
	}

	chunk, err := io.ReadAll(s.file)
	if err != nil {
		return domain.LogStreamUpdate{}, err
	}

	s.offset += int64(len(chunk))
	s.fileInfo = info

	lines := splitCompleteLines(&s.pending, chunk)
	if len(lines) == 0 {
		return domain.LogStreamUpdate{}, nil
	}

	return domain.LogStreamUpdate{
		Lines:   lines,
		Changed: true,
	}, nil
}

func (s *Stream) Close() error {
	return s.closeFile()
}

func (s *Stream) reset() error {
	s.closeFile()
	s.offset = 0
	s.pending = nil
	s.lines = nil

	file, err := os.Open(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}

	lines, offset, err := readRecentFromFile(file, s.maxLines)
	if err != nil {
		file.Close()
		return err
	}

	info, err := file.Stat()
	if err != nil {
		file.Close()
		return err
	}
	if _, err := file.Seek(offset, io.SeekStart); err != nil {
		file.Close()
		return err
	}

	s.file = file
	s.fileInfo = info
	s.offset = offset
	s.lines = lines
	return nil
}

func (s *Stream) closeFile() error {
	if s.file == nil {
		return nil
	}

	err := s.file.Close()
	s.file = nil
	return err
}

func splitCompleteLines(pending *[]byte, chunk []byte) []string {
	if len(chunk) == 0 && len(*pending) == 0 {
		return nil
	}

	data := append(append([]byte(nil), *pending...), chunk...)
	segments := bytes.Split(data, []byte{'\n'})
	if len(segments) == 0 {
		return nil
	}

	last := segments[len(segments)-1]
	*pending = append((*pending)[:0], last...)
	segments = segments[:len(segments)-1]

	lines := make([]string, 0, len(segments))
	for _, segment := range segments {
		lines = append(lines, string(bytes.TrimSuffix(segment, []byte{'\r'})))
	}
	return lines
}

func slicesEqual(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
