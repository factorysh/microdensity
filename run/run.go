package run

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/factorysh/microdensity/task"
	"github.com/factorysh/microdensity/volumes"
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
	tasks       map[uuid.UUID]*Context
	servicesDir string
	volumes     *volumes.Volumes
}

func NewRunner(servicesDir string, volumesRoot string) (*Runner, error) {

	v, err := volumes.New(volumesRoot)
	if err != nil {
		return nil, err
	}

	return &Runner{
		tasks:       make(map[uuid.UUID]*Context),
		servicesDir: servicesDir,
		volumes:     v,
	}, nil
}

type ClosingBuffer struct {
	*bytes.Buffer
}

func (c *ClosingBuffer) Close() error {
	return nil
}

func (r *Runner) Run(t *task.Task) (int, error) {
	if t.Id == uuid.Nil {
		return 0, errors.New("the task has no id")
	}

	runnable, err := NewComposeRun(fmt.Sprintf("%s/%s", r.servicesDir, t.Service))
	if err != nil {
		return 0, err
	}

	err = runnable.Prepare(nil, r.volumes.Path(t.Project, t.Branch, t.Id.String()))
	if err != nil {
		return 0, err
	}

	r.tasks[t.Id] = &Context{
		task:   t,
		Stdout: &ClosingBuffer{&bytes.Buffer{}},
		Stderr: &ClosingBuffer{&bytes.Buffer{}},
		run:    runnable,
	}

	ctx := r.tasks[t.Id]

	return ctx.run.Run(ctx.Stdout, ctx.Stderr)
}
