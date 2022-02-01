package queue

import (
	"testing"

	"github.com/factorysh/microdensity/run"
	"github.com/factorysh/microdensity/task"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestDeq(t *testing.T) {
	db, cleanFunc, err := newTestBbolt()
	defer cleanFunc()
	assert.NoError(t, err)
	q, err := New(db)

	r, err := run.NewRunner("../", "/tmp/microdensity/volumes")
	assert.NoError(t, err)
	que := NewQueue(q, r)

	tsk1 := &task.Task{
		Id:      uuid.New(),
		Service: "demo",
		Project: "beuha",
	}
	err = r.Prepare(tsk1)
	assert.NoError(t, err)

	tsk2 := &task.Task{
		Id:      uuid.New(),
		Service: "demo",
		Project: "alice",
	}
	err = r.Prepare(tsk2)
	assert.NoError(t, err)

	tsk3 := &task.Task{
		Id:      uuid.New(),
		Project: "another",
		Service: "demo",
	}
	err = r.Prepare(tsk3)
	assert.NoError(t, err)

	tsk4 := &task.Task{
		Id:      uuid.New(),
		Project: "notprepared",
		Service: "demo",
	}
	assert.NoError(t, err)

	// FIXME: asserts on state status
	que.Put(tsk1)
	que.Put(tsk2)
	que.Put(tsk3)
	que.Put(tsk4)

	<-que.BatchEnded

}
