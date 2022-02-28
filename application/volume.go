package application

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/factorysh/microdensity/html"
	"github.com/factorysh/microdensity/task"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

var (
	// used to ensure embed import
	_ embed.FS
	//go:embed templates/task.html
	taskTemplate string
)

// VolumesHandler expose volumes of a task
func (a *Application) VolumesHandler(basePathLen int, latest bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := a.logger.With(
			zap.String("url", r.URL.String()),
			zap.String("service", chi.URLParam(r, "serviceID")),
			zap.String("project", chi.URLParam(r, "project")),
			zap.String("branch", chi.URLParam(r, "branch")),
			zap.String("commit", chi.URLParam(r, "commit")),
			zap.String("branch", chi.URLParam(r, "branch")),
		)
		service := chi.URLParam(r, "serviceID")
		project := chi.URLParam(r, "project")
		branch := chi.URLParam(r, "branch")
		commit := chi.URLParam(r, "commit")

		t, err := a.storage.GetByCommit(service, project, branch, commit, latest)
		if err != nil {
			l.Error("Get task", zap.Error(err))
			w.WriteHeader(http.StatusNotFound)
			return
		}

		filePath, err := extractPathFromURL(r.URL.Path, basePathLen)
		if err != nil {
			l.Error("Extrat path", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if strings.Contains(filePath, "..") {
			l.Warn("Dangerous path trying to access parent directory")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		fullPath := filepath.Join(a.storage.GetVolumePath(t), filePath)

		// if we just want a regular file/directory, expose it
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			l.Warn("Path not found", zap.Error(err))

			data, err := NewTaskPage(t, "No result for this task", a.GitlabURL, "Task Result", "container task-output")
			if err != nil {
				l.Error("when creating result from a task", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			p := html.Page{
				Domain: a.Domain,
				Detail: fmt.Sprintf("%s / %s", t.Service, t.Commit),
				Partial: html.Partial{
					Template: taskTemplate,
					Data:     data,
				}}

			w.WriteHeader(http.StatusNotFound)
			err = p.Render(w)
			if err != nil {
				l.Error("when trying to render 404 result page", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			return
		}

		// if we want the html result of a task
		// return early in case of success
		if strings.HasSuffix(fullPath, "result.html") {
			err := a.renderResultPageForTask(t, fullPath, w)
			if err != nil {
				l.Warn("when trying to access a result page", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			return
		}

		http.ServeFile(w, r, fullPath)
	}
}

func extractPathFromURL(p string, basePathLen int) (string, error) {
	urlPath, file := path.Split(strings.TrimLeft(p, "/"))

	parts := strings.Split(urlPath, "/")
	if len(parts) < basePathLen {
		return "", fmt.Errorf("can not split path `%s` in more than %d elements", p, basePathLen)
	}

	folderPath := parts[basePathLen:]
	folderPath = append(folderPath, file)

	return path.Join(folderPath...), nil
}

func (a *Application) renderResultPageForTask(t *task.Task, filePath string, w http.ResponseWriter) error {
	// try to fetch the result page from fs
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	data, err := NewTaskPage(t, template.HTML(content), a.GitlabURL, "Task Result", "container task-output")
	// create the page
	p := html.Page{
		Domain: a.Domain,
		Detail: fmt.Sprintf("%s / %s", t.Service, t.Commit),
		Partial: html.Partial{
			Template: taskTemplate,
			Data:     data,
		},
	}

	w.WriteHeader(http.StatusOK)
	return p.Render(w)
}
