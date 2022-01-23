package run

import (
	"bytes"
	"context"
	"errors"
	"io"

	"github.com/factorysh/microdensity/queue"
	"github.com/factorysh/microdensity/task"
	"github.com/google/uuid"
)

type Run struct {
	Stdout io.WriteCloser
	Stderr io.WriteCloser
	cancel context.CancelFunc
	task   *task.Task
}

type Runner struct {
	queue *queue.Queue
	tasks map[uuid.UUID]*Run
}

type ClosingBuffer struct {
	*bytes.Buffer
}

func (c *ClosingBuffer) Close() error {
	return nil
}

func (r *Runner) Run(t *task.Task) error {
	if t.Id == uuid.Nil {
		return errors.New("the has no id")
	}
	r.tasks[t.Id] = &Run{
		task:   t,
		Stdout: &ClosingBuffer{},
		Stderr: &ClosingBuffer{},
	}
	return nil
}
