package run

import (
	"bytes"
	"errors"
	"io"

	"github.com/factorysh/microdensity/task"
	"github.com/google/uuid"
)

type Context struct {
	Stdout io.WriteCloser
	Stderr io.WriteCloser
	task   *task.Task
	run    Runnable
}

type Runnable interface {
	Prepare(map[string]string, string) error
	Run(stdout io.WriteCloser, stderr io.WriteCloser) (int, error)
	Cancel()
}

type Runner struct {
	tasks map[uuid.UUID]*Context
}

type ClosingBuffer struct {
	*bytes.Buffer
}

func (c *ClosingBuffer) Close() error {
	return nil
}

func (r *Runner) Run(t *task.Task) error {
	if t.Id == uuid.Nil {
		return errors.New("the task has no id")
	}
	r.tasks[t.Id] = &Context{
		task:   t,
		Stdout: &ClosingBuffer{},
		Stderr: &ClosingBuffer{},
	}
	return nil
}
