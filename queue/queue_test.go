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

	que := NewQueue(q, run.NewRunner())

	tsk1 := &task.Task{
		Id:      uuid.New(),
		Service: "demo",
		Project: "beuha",
	}

	tsk2 := &task.Task{
		Id:      uuid.New(),
		Service: "demo",
		Project: "alice",
	}

	tsk3 := &task.Task{
		Id:      uuid.New(),
		Project: "another",
		Service: "demo",
	}

	que.Put(tsk1)
	que.Put(tsk2)
	que.Put(tsk3)

	<-que.BatchEnded
}
