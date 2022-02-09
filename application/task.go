package application

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"time"

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
	err = a.storage.Upsert(t)
	if err != nil {
		l.Warn("Queue error", zap.Error(err))
		panic(err)
	}
	err = a.storage.EnsureVolumesDir(t)
	if err != nil {
		l.Warn("Volume creation", zap.Error(err))
		panic(err)
	}
	err = a.queue.Put(t, parsedArgs.Environments)
	if err != nil {
		l.Warn("Task prepare/put", zap.Error(err))
		panic(err)
	}
	err = a.storage.SetLatest(t)
	if err != nil {
		l.Warn("Task set latest", zap.Error(err))
		panic(err)
	}
	render.JSON(w, r, map[string]string{
		"id": id.String(),
	})
}

// TaskHandler show a Task
func (a *Application) TaskHandler(w http.ResponseWriter, r *http.Request) {
	l := a.logger.With(
		zap.String("url", r.URL.String()),
		zap.String("service", chi.URLParam(r, "serviceID")),
		zap.String("project", chi.URLParam(r, "project")),
		zap.String("branch", chi.URLParam(r, "branch")),
		zap.String("commit", chi.URLParam(r, "commit")),
	)
	t, err := a.volumes.Get(
		chi.URLParam(r, "serviceID"),
		chi.URLParam(r, "project"),
		chi.URLParam(r, "branch"),
		chi.URLParam(r, "commit"),
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
