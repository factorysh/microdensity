package application

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/factorysh/microdensity/httpcontext"
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
	}
	var args map[string]interface{}
	err = render.DecodeJSON(r.Body, args)
	if err != nil {
		panic(err)
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
