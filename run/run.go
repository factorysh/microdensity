package run

import (
	"io"

	"github.com/google/uuid"
)

type Run interface {
	Prepare(args map[string]string) error
	Id() uuid.UUID
	Run(stdout io.WriteCloser, stderr io.WriteCloser) (int, error)
}
