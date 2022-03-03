package application

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/http"

	"github.com/docker/docker/pkg/stdcopy"
	"github.com/factorysh/microdensity/html"
	"github.com/factorysh/microdensity/task"
	"github.com/robert-nix/ansihtml"
)

func (a *Application) renderLogsPageForTask(ctx context.Context, t *task.Task, w http.ResponseWriter) error {

	reader, err := t.Logs(ctx, false)
	if err != nil {
		return err
	}

	var buffer bytes.Buffer
	_, err = stdcopy.StdCopy(&buffer, &buffer, reader)
	if err != nil {
		return err
	}

	data, err := NewTaskPage(t, template.HTML(fmt.Sprintf("<pre>%s</pre>", ansihtml.ConvertToHTML(buffer.Bytes()))), a.GitlabURL, "Task Logs", "terminal")
	if err != nil {
		return err
	}

	p := html.Page{
		Domain: a.Domain,
		Detail: fmt.Sprintf("%s / %s - logs", t.Service, t.Commit),
		Partial: html.Partial{
			Template: taskTemplate,
			Data:     data,
		},
	}

	w.WriteHeader(http.StatusOK)
	return p.Render(w)
}
