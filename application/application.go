package application

import (
	"encoding/json"
	"net/http"

	"github.com/factorysh/microdensity/queue"
	"github.com/factorysh/microdensity/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Application struct {
	Services []*service.Service
	queue    *queue.Queue
	router   chi.Router
}

func New(q *queue.Queue) (*Application, error) {

	r := chi.NewRouter()
	a := &Application{
		Services: make([]*service.Service, 0),
		queue:    q,
		router:   r,
	}
	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/services", a.services)
	r.Get("/service/{serviceID}", a.service)

	return a, nil
}

func (a *Application) services(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	ss := make([]string, len(a.Services))
	for n, service := range a.Services {
		ss[n] = service.Name
	}
	err := json.NewEncoder(w).Encode(ss)
	if err != nil {
		panic(err)
	}
}

func (a *Application) findServiceByName(name string) *service.Service {
	for _, s := range a.Services {
		if s.Name == name {
			return s
		}
	}
	return nil
}

func (a *Application) service(w http.ResponseWriter, r *http.Request) {
	if serviceId := chi.URLParam(r, "serviceID"); serviceId != "" {
		service := a.findServiceByName(serviceId)
		if service == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}
}
