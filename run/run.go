package run

import (
	"io"

	"golang.org/x/net/context"
)

type Run interface {
	Run(ctx context.Context, args map[string]string, stdout io.WriteCloser, stderr io.WriteCloser) (int, error)
}
