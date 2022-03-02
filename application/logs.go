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

type doubleLogger struct {
	Stdout io.Writer
	Stderr io.Writer
}

type simpleLogger struct {
	ouptut io.Writer
	class  string
}

var space = []byte(" ")
var cr = []byte("\n")

func (s *simpleLogger) Write(p []byte) (int, error) {
	lines := bytes.Split(p, cr)
	n := 0

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		parts := bytes.SplitN(line, space, 2)
		parts[0] = []byte(fmt.Sprintf("<span class=\"%s\">%s</span>", s.class, parts[0]))
		i, err := s.ouptut.Write(bytes.Join(parts, space))
		if err != nil {
			return 0, err
		}
		j, err := s.ouptut.Write(cr)
		if err != nil {
			return 0, err
		}
		n += i + j
	}

	return n, nil
}

func newDoubleLogger(output io.Writer) *doubleLogger {
	return &doubleLogger{
		Stdout: &simpleLogger{
			ouptut: output,
			class:  "stdout-prefix",
		},
		Stderr: &simpleLogger{
			ouptut: output,
			class:  "stderr-prefix",
		},
	}
}

func (a *Application) renderLogsPageForTask(ctx context.Context, t *task.Task, w http.ResponseWriter) error {

	reader, err := t.Logs(ctx, false)
	if err != nil {
		return err
	}

	var buffer bytes.Buffer
	logger := newDoubleLogger(&buffer)
	_, err = stdcopy.StdCopy(logger.Stdout, logger.Stderr, reader)
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
