package queue

import (
	"bytes"
	"encoding/gob"
	"sync"
	"testing"

	"github.com/factorysh/microdensity/task"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestEncoding(t *testing.T) {
	var b bytes.Buffer
	e := &Encoding{
		lock:    &sync.Mutex{},
		encoder: gob.NewEncoder(&b),
		decoder: gob.NewDecoder(&b),
		buffer:  b,
	}

	ida, err := uuid.NewUUID()
	assert.NoError(t, err)
	alice := &task.Task{
		Id:      ida,
		Project: "Alice",
	}
	idb, err := uuid.NewUUID()
	assert.NoError(t, err)
	bob := &task.Task{
		Id:      idb,
		Project: "Bob",
	}

	rawa, err := e.Encode(alice)
	assert.NoError(t, err)

	rawb, err := e.Encode(bob)
	assert.NoError(t, err)

	aa, err := e.Decode(rawa)
	assert.NoError(t, err)
	assert.Equal(t, ida, aa.Id)

	bb, err := e.Decode(rawb)
	assert.NoError(t, err)
	assert.Equal(t, idb, bb.Id)
}
