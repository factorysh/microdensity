package queue

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/factorysh/microdensity/run"
	"github.com/factorysh/microdensity/storage"
	"github.com/factorysh/microdensity/task"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestDeq(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "data-")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)
	store, err := storage.NewFSStore(dir)
	assert.NoError(t, err)

	r, err := run.NewRunner("../demo/services", "/tmp/microdensity/volumes", []string{})
	assert.NoError(t, err)
	que := NewQueue(store, r)

	tsk1 := &task.Task{
		Id:      uuid.New(),
		Service: "demo",
		Project: "beuha",
	}
	err = r.Prepare(tsk1, nil)
	assert.NoError(t, err)

	tsk2 := &task.Task{
		Id:      uuid.New(),
		Service: "demo",
		Project: "alice",
	}
	err = r.Prepare(tsk2, nil)
	assert.NoError(t, err)

	tsk3 := &task.Task{
		Id:      uuid.New(),
		Project: "another",
		Service: "demo",
	}
	err = r.Prepare(tsk3, nil)
	assert.NoError(t, err)

	tsk4 := &task.Task{
		Id:      uuid.New(),
		Project: "notprepared",
		Service: "demo",
	}
	assert.NoError(t, err)

	// FIXME: asserts on state status
	que.Put(tsk1, nil)
	que.Put(tsk2, nil)
	que.Put(tsk3, nil)
	que.Put(tsk4, nil)

	<-que.BatchEnded

}
