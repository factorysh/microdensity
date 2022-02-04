package application

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

// ServicesHandler show all services
func (a *Application) ServicesHandler(w http.ResponseWriter, r *http.Request) {
	ss := make([]string, len(a.Services))
	i := 0
	for _, service := range a.Services {
		ss[i] = service.Name()
		i++
	}
	render.JSON(w, r, ss)
}

// ServiceHandler show one service
func (a *Application) ServiceHandler(w http.ResponseWriter, r *http.Request) {
	if serviceId := chi.URLParam(r, "serviceID"); serviceId != "" {
		service := a.Services[serviceId]
		if service == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		render.JSON(w, r, map[string]interface{}{
			"name": serviceId,
		})
	}
}

// Http middleware
func (a *Application) ServiceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serviceId := chi.URLParam(r, "serviceID")
		if serviceId == "" {
			http.Error(w, http.StatusText(404), 404)
			return
		}
		service := a.Services[serviceId]
		if service == nil {
			http.Error(w, http.StatusText(404), 404)
			return
		}
		ctx := context.WithValue(r.Context(), "service", service)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
