package application

import (
	"html/template"
	"net/url"

	"github.com/factorysh/microdensity/task"
)

// Result of a task for html rendering
type Result struct {
	GitlabDomain string
	Domain       string
	Project      string
	Commit       string
	Service      string
	ID           string
	Inner        template.HTML
}

// NewResultFromTask inits a result for a task
func NewResultFromTask(t *task.Task, inner template.HTML, gitlabDomain string) (*Result, error) {
	prettyPath, err := url.PathUnescape(t.Project)
	if err != nil {
		return nil, err
	}

	return &Result{
		Project:      prettyPath,
		GitlabDomain: gitlabDomain,
		Commit:       t.Commit,
		Service:      t.Service,
		ID:           t.Id.String(),
		Inner:        inner,
	}, nil
}
