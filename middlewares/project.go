package middlewares

import (
	"context"
	"net/http"

	"github.com/factorysh/microdensity/httpcontext"
	"github.com/go-chi/chi/v5"
)

// Project adds requested project to context
func Project() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			project := chi.URLParam(r, "project")
			if project == "" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			ctx := context.WithValue(r.Context(), httpcontext.RequestedProject, project)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
