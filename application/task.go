package application

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	_claims "github.com/factorysh/microdensity/claims"
	_service "github.com/factorysh/microdensity/service"
	"github.com/factorysh/microdensity/task"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (a *Application) newTask(w http.ResponseWriter, r *http.Request) {
	service := chi.URLParam(r, "serviceID")
	project := chi.URLParam(r, "project")
	l := a.logger.With(
		zap.String("service", service),
		zap.String("project", project),
	)
	claims, err := _claims.FromCtx(r.Context())
	if err != nil {
		l.Warn("Claims error", zap.Error(err))
		panic(err)
	}
	if project != url.QueryEscape(claims.Path) {
		l.Warn("Path mismatch with claims", zap.String("claims.Path", claims.Path))
		w.WriteHeader(403)
		return
	}
	var args map[string]interface{}

	err = render.DecodeJSON(r.Body, &args)
	if err != nil {
		l.Warn("Body JSON decode error", zap.Error(err))
		panic(err)
	}
	s := r.Context().Value("service").(_service.Service)
	err = s.Validate(args)
	if err != nil {
		l.Warn("Validation error", zap.Error(err))
		w.WriteHeader(400)

		fmt.Fprintln(w, err)
		return
	}
	id, err := uuid.NewUUID()
	if err != nil {
		panic(err)
	}
	l = l.With(zap.String("id", id.String()))
	t := &task.Task{
		Id:       id,
		Service:  service,
		Project:  project,
		Branch:   chi.URLParam(r, "branch"),
		Commit:   chi.URLParam(r, "commit"),
		Creation: time.Now(),
		Args:     args,
		State:    task.Ready,
	}
	err = a.volumes.Create(t)
	if err != nil {
		l.Warn("Volume creation", zap.Error(err))
		panic(err)
	}
	err = a.queue.Put(t)
	if err != nil {
		l.Warn("Queue error", zap.Error(err))
		panic(err)
	}
	json.NewEncoder(w).Encode(map[string]string{
		"id": id.String(),
	})
}

func (a *Application) task(w http.ResponseWriter, r *http.Request) {
	t, err := a.volumes.Get(
		chi.URLParam(r, "serviceID"),
		chi.URLParam(r, "project"),
		chi.URLParam(r, "branch"),
		chi.URLParam(r, "commit"),
	)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println(err)
		return
	}
	err = json.NewEncoder(w).Encode(t)
	if err != nil {
		panic(err)
	}
}
