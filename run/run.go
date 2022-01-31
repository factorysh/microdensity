package run

import (
	"bytes"
	"errors"
	"fmt"
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

func NewRunner() *Runner {
	return &Runner{
		tasks: make(map[uuid.UUID]*Context),
	}
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

	// FIXME: dir for service
	runnable, err := NewComposeRun(fmt.Sprintf("../%s", t.Service))
	if err != nil {
		return err
	}

	// FIXME: Args and volume root
	err = runnable.Prepare(nil, "/tmp")
	if err != nil {
		return err
	}

	r.tasks[t.Id] = &Context{
		task:   t,
		Stdout: &ClosingBuffer{&bytes.Buffer{}},
		Stderr: &ClosingBuffer{&bytes.Buffer{}},
		run:    runnable,
	}

	ctx := r.tasks[t.Id]
	ret, err := ctx.run.Run(ctx.Stdout, ctx.Stderr)
	fmt.Println(ret)

	return err
}
