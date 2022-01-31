package project

import (
	"net/http"
	"net/url"

	"github.com/factorysh/microdensity/httpcontext"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// Project adds requested project to context
func AssertProject(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger, _ := zap.NewProduction()
		l := logger.With(zap.String("url", r.URL.String()))
		projectRaw := r.Context().Value(httpcontext.RequestedProject)
		if projectRaw == nil {
			l.Warn("No project in context")
			w.WriteHeader(http.StatusForbidden)
			return
		}
		pp := projectRaw.(string)
		l = l.With(zap.String("project", pp))
		if pp == "" {
			l.Warn("empty project")
			w.WriteHeader(http.StatusForbidden)
			return
		}
		p, err := url.QueryUnescape(chi.URLParam(r, "project"))
		if err != nil || p == "" {
			l.Error("can't find project in url", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		l = l.With(zap.String("project url", p))
		if pp != p {
			l.Warn("projects are not equal")
			w.WriteHeader(http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
