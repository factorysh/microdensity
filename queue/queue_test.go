package queue

import (
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"testing"

	"github.com/docker/go-events"
	"github.com/factorysh/microdensity/run"
	"github.com/factorysh/microdensity/sink"
	"github.com/factorysh/microdensity/storage"
	"github.com/factorysh/microdensity/task"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

var _ events.Sink = (*DummyEventLogger)(nil)

type DummyEventLogger struct {
	Cpt *sync.WaitGroup
}

func (d *DummyEventLogger) Write(evt events.Event) error {
	d.Cpt.Done()
	fmt.Println("evt", evt)
	return nil
}

func (d *DummyEventLogger) Close() error {
	return nil
}

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestDeq(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "data-")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)
	store, err := storage.NewFSStore(dir)
	assert.NoError(t, err)

	r, err := run.NewRunner("../demo/services", "/tmp/microdensity/volumes", []string{})
	assert.NoError(t, err)
	que := NewQueue(store, r, &sink.VoidSink{})
	snk := &DummyEventLogger{
		Cpt: &sync.WaitGroup{},
	}
	que.Sink = snk

	dummySink := &DummyEventLogger{}
	dummySink.Cpt = &sync.WaitGroup{}

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

	snk.Cpt.Add(4)
	// FIXME: asserts on state status
	err = que.Put(tsk1, nil)
	assert.NoError(t, err)
	err = que.Put(tsk2, nil)
	assert.NoError(t, err)
	err = que.Put(tsk3, nil)
	assert.NoError(t, err)

	<-que.BatchEnded
	//snk.Cpt.Wait()

}
