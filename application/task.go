package application

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/factorysh/microdensity/httpcontext"
	_service "github.com/factorysh/microdensity/service"
	"github.com/factorysh/microdensity/task"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

func (a *Application) newTask(w http.ResponseWriter, r *http.Request) {
	service := chi.URLParam(r, "serviceID")
	project := chi.URLParam(r, "projet")
	jwtProject, err := httpcontext.GetRequestedProject(r)
	if err != nil {
		panic(err)
	}
	if project != jwtProject {
		w.WriteHeader(403)
		return
	}
	var args map[string]interface{}

	err = render.DecodeJSON(r.Body, args)
	if err != nil {
		panic(err)
	}
	s := r.Context().Value("service").(_service.Service)
	err = s.Validate(args)
	if err != nil {
		w.WriteHeader(400)
		fmt.Fprintln(w, err)
		return
	}
	id, err := uuid.NewUUID()
	if err != nil {
		panic(err)
	}
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
	a.queue.Put(t)
	json.NewEncoder(w).Encode(map[string]string{
		"id": id.String(),
	})
}

func (a *Application) task(w http.ResponseWriter, r *http.Request) {
}
