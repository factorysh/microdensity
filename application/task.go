package application

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/docker/docker/pkg/stdcopy"
	_claims "github.com/factorysh/microdensity/claims"
	"github.com/factorysh/microdensity/task"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
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
	render.JSON(w, r, map[string]string{
		"id": id.String(),
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

// TaskLogsHandler get a logs for a task
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

		reader, err := t.Logs(r.Context(), false)
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

		err = a.renderLogsPageForTask(r.Context(), t, w)
		if err != nil {
			l.Warn("when rendering a logs page", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
			return
		}
	}
}
