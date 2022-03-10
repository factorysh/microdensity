package run

/*
User defines a task.Task, after validation, it goes to a queue.Queue
and a run.Runner pick it and run.Runner#Prepare then run.Runner#Run it.
*/
import (
	"bytes"
	"fmt"
	"io"

	"github.com/factorysh/microdensity/task"
	"github.com/factorysh/microdensity/volumes"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	serviceRun = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "run_total",
		Help: "Total tasks run",
	}, []string{"service", "project"})
)

// Context is a run context, with a STDOUT and a STDERR
type Context struct {
	Stdout io.WriteCloser
	Stderr io.WriteCloser
	task   *task.Task
	run    Runnable
}

type Runnable interface {
	Prepare(map[string]string, string, uuid.UUID, []string) error
	Run(stdout io.WriteCloser, stderr io.WriteCloser) (int, error)
	Cancel()
}

type Runner struct {
	tasks       map[uuid.UUID]*Context
	servicesDir string
	volumes     *volumes.Volumes
	hosts       []string
}

func NewRunner(servicesDir string, volumesRoot string, hosts []string) (*Runner, error) {
	v, err := volumes.New(volumesRoot)
	if err != nil {
		return nil, err
	}

	return &Runner{
		tasks:       make(map[uuid.UUID]*Context),
		servicesDir: servicesDir,
		volumes:     v,
		hosts:       hosts,
	}, nil
}

// Prepare the run
// Prepare is synchronous, in order to raise an error in the REST endpoint.
// Prepare checks volumes stuff.
func (r *Runner) Prepare(t *task.Task, env map[string]string) (string, error) {
	if t.Id == uuid.Nil {
		return "", fmt.Errorf("task requires an ID to be prepared")
	}

	if _, found := r.tasks[t.Id]; found {
		return "", fmt.Errorf("task with id `%s` already prepared", t.Id)
	}

	runnable, err := NewComposeRun(fmt.Sprintf("%s/%s", r.servicesDir, t.Service))
	if err != nil {
		return "", err
	}

	err = runnable.Prepare(env,
		r.volumes.Path(t.Service, t.Project, t.Branch, t.Id.String()),
		t.Id,
		r.hosts)
	if err != nil {
		return "", err
	}

	r.tasks[t.Id] = &Context{
		task:   t,
		Stdout: &ClosingBuffer{&bytes.Buffer{}},
		Stderr: &ClosingBuffer{&bytes.Buffer{}},
		run:    runnable,
	}

	return runnable.run, nil
}

func (r *Runner) Run(t *task.Task) (int, error) {

	ctx, found := r.tasks[t.Id]
	if !found {
		return 0, fmt.Errorf("task with id `%s` not found in runner", t.Id)
	}
	defer serviceRun.With(prometheus.Labels{
		"service": t.Service,
		"project": t.Project}).Inc()
	return ctx.run.Run(ctx.Stdout, ctx.Stderr)
}
