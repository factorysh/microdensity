package application

import (
	"html/template"
	"net/url"

	"github.com/factorysh/microdensity/task"
)

// TaskPage present a page and an associated inner content
type TaskPage struct {
	GitlabDomain    string
	Domain          string
	Project         string
	Commit          string
	Service         string
	ID              string
	CreatedAt       string
	InnerDivClasses string
	InnerTitle      string
	Inner           template.HTML
}

// NewTaskPage inits a result for a task
func NewTaskPage(t *task.Task, inner template.HTML, gitlabDomain, InnerTitle, InnerDivClasses string) (*TaskPage, error) {
	prettyPath, err := url.PathUnescape(t.Project)
	if err != nil {
		return nil, err
	}

	return &TaskPage{
		Project:         prettyPath,
		GitlabDomain:    gitlabDomain,
		Commit:          t.Commit,
		Service:         t.Service,
		ID:              t.Id.String(),
		CreatedAt:       t.Creation.Format("2006-01-02 15:04:05"),
		InnerTitle:      InnerTitle,
		InnerDivClasses: InnerDivClasses,
		Inner:           inner,
	}, nil
}
