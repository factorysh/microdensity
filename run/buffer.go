package run

import "bytes"

// ClosingBuffer it's a bytes.Buffer, but closable
type ClosingBuffer struct {
	*bytes.Buffer
}

func (c *ClosingBuffer) Close() error {
	return nil
}
