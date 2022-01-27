package queue

import (
	"testing"

	"github.com/factorysh/microdensity/task"
	"github.com/stretchr/testify/assert"
)

func TestDeq(t *testing.T) {
	db, cleanFunc, err := newTestBbolt()
	defer cleanFunc()
	assert.NoError(t, err)
	q, err := New(db)

	que := NewQueue(q)

	tsk1 := &task.Task{
		Project: "beuha",
	}

	tsk2 := &task.Task{
		Project: "alice",
	}

	tsk3 := &task.Task{
		Project: "another",
	}

	que.Put(tsk1)
	que.Put(tsk2)
	que.Put(tsk3)

	<-que.BatchEnded
}
