package application

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/factorysh/microdensity/service"
	"github.com/go-chi/chi/v5"
)

func (a *Application) services(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	ss := make([]string, len(a.Services))
	for n, service := range a.Services {
		ss[n] = service.Name()
	}
	err := json.NewEncoder(w).Encode(ss)
	if err != nil {
		panic(err)
	}
}

func (a *Application) findServiceByName(name string) service.Service {
	for _, s := range a.Services {
		if s.Name() == name {
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

func (a *Application) serviceCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serviceId := chi.URLParam(r, "serviceID")
		if serviceId == "" {
			http.Error(w, http.StatusText(404), 404)
			return
		}
		service := a.findServiceByName(serviceId)
		if service == nil {
			http.Error(w, http.StatusText(404), 404)
			return
		}
		ctx := context.WithValue(r.Context(), "service", service)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
