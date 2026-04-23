package taskid

import (
	"crypto/rand"
	"encoding/hex"

	"dst-server-ctl/internal/domain"
)

type Generator struct{}

func (Generator) NewTaskID() (domain.TaskID, error) {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", err
	}

	return domain.TaskID(hex.EncodeToString(bytes[:])), nil
}
