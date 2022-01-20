package run

import (
	"io"

	"github.com/google/uuid"
)

type Run interface {
	// Validate the arguments, build the id
	Prepare(args map[string]string) error
	Id() uuid.UUID
	//Cancel this Run
	Cancel()
	// Sync run the compose, return the return code
	Run(stdout io.WriteCloser, stderr io.WriteCloser) (int, error)
}
