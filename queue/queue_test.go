package queue

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/factorysh/microdensity/task"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.etcd.io/bbolt"
)

func TestQueue(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "queue-")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	s, err := bbolt.Open(
		fmt.Sprintf("%s/bbolt.store", dir),
		0600, &bbolt.Options{})
	assert.NoError(t, err)
	q, err := New(s)
	assert.NoError(t, err)
	size, err := q.Length()
	assert.NoError(t, err)
	assert.Equal(t, 0, size)
	tsk := &task.Task{
		Project: "beuha",
	}
	assert.Equal(t, uuid.Nil, tsk.Id)
	err = q.Set(tsk)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, tsk.Id)
	size, err = q.Length()
	assert.NoError(t, err)
	assert.Equal(t, 1, size)

}
