package logtail

import (
	"bufio"
	"context"
	"errors"
	"io"
	"os"
)

type Reader struct{}

func (Reader) ReadRecent(_ context.Context, path string, maxLines int) ([]string, error) {
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return []string{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()

	lines, _, err := readRecentFromFile(file, maxLines)
	return lines, err
}

func readRecentFromFile(file *os.File, maxLines int) ([]string, int64, error) {
	lines := make([]string, 0, maxLines)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if maxLines > 0 && len(lines) > maxLines {
			lines = lines[1:]
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, 0, err
	}

	offset, err := file.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, 0, err
	}

	return lines, offset, nil
}
