package application

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	docker "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	_claims "github.com/factorysh/microdensity/claims"
	"github.com/factorysh/microdensity/html"
	"github.com/factorysh/microdensity/task"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/robert-nix/ansihtml"
	"go.uber.org/zap"
)

// PostTaskHandler create a Task
func (a *Application) PostTaskHandler(w http.ResponseWriter, r *http.Request) {
	serviceID := chi.URLParam(r, "serviceID")
	project := chi.URLParam(r, "project")
	l := a.logger.With(
		zap.String("service", serviceID),
		zap.String("project", project),
	)

	// get the service interface for the requested service
	service, found := a.Services[serviceID]
	if !found {
		l.Warn("Requested service not found", zap.String("serviceID", serviceID))
		w.WriteHeader(http.StatusInternalServerError)
	}

	claims, err := _claims.FromCtx(r.Context())
	if err != nil {
		l.Warn("Claims error", zap.Error(err))
		panic(err)
	}
	if project != url.QueryEscape(claims.ProjectPath) && project != claims.ID {
		l.Warn("Path mismatch with claims", zap.String("claims.Path", claims.ProjectPath))
		w.WriteHeader(403)
		return
	}
	var args map[string]interface{}

	err = render.DecodeJSON(r.Body, &args)
	if err != nil {
		l.Warn("Body JSON decode error", zap.Error(err))
		w.WriteHeader(400)
		render.JSON(w, r, map[string]string{
			"error": err.Error(),
		})
		return
	}

	// validate the arguments
	parsedArgs, err := service.Validate(args)
	if err != nil {
		l.Warn("Validation error",
			zap.Any("args", args),
			zap.Error(err))
		//panic(err)
		w.Header().Set("content-type", "application/json; charset=utf-8")
		w.WriteHeader(400)
		err = json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
		if err != nil {
			panic(err)
		}
		return
	}
	id, err := uuid.NewUUID()
	if err != nil {
		panic(err)
	}
	l = l.With(zap.String("id", id.String()))
	t := &task.Task{
		Id:       id,
		Service:  serviceID,
		Project:  project,
		Branch:   chi.URLParam(r, "branch"),
		Commit:   chi.URLParam(r, "commit"),
		Creation: time.Now(),
		Args:     args,
		State:    task.Ready,
	}

	err = a.addTask(t, parsedArgs.Environments)
	if err != nil {
		l.Error("error when adding task", zap.String("task", t.Id.String()), zap.Error(err))
		return
	}

	url := strings.Join([]string{a.Domain, "service", t.Service, t.Project, t.Branch, t.Commit, "volumes", "data", "result.html"}, "/")

	if html.Accepts(r, "text/plain") {
		w.Header().Add("content-type", "text/plain")
		_, err := w.Write([]byte(fmt.Sprintf("Result URL:\n\n\t\t%s\n\n", url)))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		}
		return
	}

	render.JSON(w, r, map[string]string{
		"id":  id.String(),
		"url": url,
	})
}

// addTask adds a task to a queue
func (a *Application) addTask(t *task.Task, args map[string]string) error {
	err := a.storage.EnsureVolumesDir(t)
	if err != nil {
		return err
	}

	err = a.queue.Put(t, args)
	if err != nil {
		return err
	}

	err = a.storage.Upsert(t)
	if err != nil {
		return err
	}

	err = a.storage.SetLatest(t)
	if err != nil {
		return err
	}

	return err
}

// TaskHandler show a Task
func (a *Application) TaskHandler(latest bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		l := a.logger.With(
			zap.String("url", r.URL.String()),
			zap.String("service", chi.URLParam(r, "serviceID")),
			zap.String("project", chi.URLParam(r, "project")),
			zap.String("branch", chi.URLParam(r, "branch")),
			zap.String("commit", chi.URLParam(r, "commit")),
		)
		t, err := a.storage.GetByCommit(
			chi.URLParam(r, "serviceID"),
			chi.URLParam(r, "project"),
			chi.URLParam(r, "branch"),
			chi.URLParam(r, "commit"),
			latest,
		)
		if err != nil {
			l.Warn("Task get error", zap.Error(err))
			if os.IsNotExist(err) {
				w.WriteHeader(http.StatusNotFound)
			} else {
				w.WriteHeader(http.StatusBadRequest)
			}
			return
		}
		err = json.NewEncoder(w).Encode(t)
		if err != nil {
			l.Error("Json encoding error", zap.Error(err))
			panic(err)
		}
	}
}

// TaskIDHandler get a task using it's ID
func (a *Application) TaskIDHandler(w http.ResponseWriter, r *http.Request) {
	l := a.logger.With(
		zap.String("url", r.URL.String()),
		zap.String("task ID", chi.URLParam(r, "taskID")),
	)

	t, err := a.storage.Get(chi.URLParam(r, "taskID"))
	if err != nil {
		l.Warn("Task get error", zap.Error(err))
		if os.IsNotExist(err) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
		return
	}

	err = json.NewEncoder(w).Encode(t)
	if err != nil {
		l.Error("Json encoding error", zap.Error(err))
		return
	}
}

// TaskLogzHandler get a logs for a task
func (a *Application) TaskLogzHandler(latest bool) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		l := a.logger.With(
			zap.String("url", r.URL.String()),
			zap.String("service", chi.URLParam(r, "serviceID")),
			zap.String("project", chi.URLParam(r, "project")),
			zap.String("branch", chi.URLParam(r, "branch")),
			zap.String("commit", chi.URLParam(r, "commit")),
		)

		t, err := a.storage.GetByCommit(
			chi.URLParam(r, "serviceID"),
			chi.URLParam(r, "project"),
			chi.URLParam(r, "branch"),
			chi.URLParam(r, "commit"),
			latest,
		)

		if err != nil {
			l.Warn("Task get error", zap.Error(err))
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(http.StatusText(http.StatusNotFound)))
			return
		}

		reader, err := t.Logs(r.Context(), false)
		if docker.IsErrNotFound(err) {
			l.Warn("container not found", zap.Error(err))
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(http.StatusText(http.StatusNotFound)))
			return
		}

		if err != nil {
			l.Warn("Task log error", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
			return
		}

		// just stdout for now
		// kudos @ndeloof, @rumpl, @glours
		_, err = stdcopy.StdCopy(w, w, reader)
		if err != nil {
			l.Error("Task log stdcopy write error", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
			return
		}

		w.Header().Set("Content-Type", "text/plain")
	}

}

// TaskLogsHandler get a logs for a task, used to get row logs from curl, not used for now
func (a *Application) TaskLogsHandler(latest bool) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		l := a.logger.With(
			zap.String("url", r.URL.String()),
			zap.String("service", chi.URLParam(r, "serviceID")),
			zap.String("project", chi.URLParam(r, "project")),
			zap.String("branch", chi.URLParam(r, "branch")),
			zap.String("commit", chi.URLParam(r, "commit")),
		)

		t, err := a.storage.GetByCommit(
			chi.URLParam(r, "serviceID"),
			chi.URLParam(r, "project"),
			chi.URLParam(r, "branch"),
			chi.URLParam(r, "commit"),
			latest,
		)
		if err != nil {
			l.Warn("Task get error", zap.Error(err))
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(http.StatusText(http.StatusNotFound)))
			return
		}

		err = a.renderLogsPageForTask(r.Context(), t, w)
		if docker.IsErrNotFound(err) {
			l.Warn("container not found", zap.Error(err))
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(http.StatusText(http.StatusNotFound)))
			return
		}

		if err != nil {
			l.Warn("when rendering a logs page", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
			return
		}
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
