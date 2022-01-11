package queue

import (
	"bytes"
	"encoding/gob"
	"sync"

	"github.com/factorysh/microdensity/task"
)

type Encoding struct {
	lock    *sync.Mutex
	encoder *gob.Encoder
	decoder *gob.Decoder
	buffer  bytes.Buffer
}

func (e *Encoding) Encode(t *task.Task) ([]byte, error) {
	e.lock.Lock()
	defer e.lock.Unlock()
	err := e.encoder.Encode(t)
	if err != nil {
		return nil, err
	}
	r := make([]byte, e.buffer.Len())
	_, err = e.buffer.Read(r)
	return r, err
}

func (e *Encoding) Decode(raw []byte) (*task.Task, error) {
	var t task.Task
	e.buffer.Write(raw)
	err := e.decoder.Decode(&t)
	if err != nil {
		return nil, err
	}
	return &t, err
}
