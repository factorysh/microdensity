package application

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io"
	"net/http"

	"github.com/docker/docker/pkg/stdcopy"
	"github.com/factorysh/microdensity/html"
	"github.com/factorysh/microdensity/task"
	"github.com/robert-nix/ansihtml"
)

type DoubleLogger struct {
	Stdout io.Writer
	Stderr io.Writer
}

type simpleLogger struct {
	ouptut io.Writer
	color  string
}

func (s *simpleLogger) Write(p []byte) (int, error) {
	lines := bytes.Split(p, []byte{10})
	n := 0
	for _, line := range lines {
		// FIXME use s.color here
		i, _ := s.ouptut.Write(line)
		n += i
	}
	return n, nil
}

func NewDoubleLogger(output io.Writer) *DoubleLogger {
	return &DoubleLogger{
		Stdout: &simpleLogger{
			ouptut: output,
			color:  "blue",
		},
		Stderr: &simpleLogger{
			ouptut: output,
			color:  "red",
		},
	}
}

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
