package application

import (
	"net/http"
	"strings"
)

// RefererMiddleware ensure that requests comes from the gitlab domain
func (a *Application) RefererMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.TrimRight(r.Referer(), "/") != strings.TrimRight(a.GitlabURL, "/") {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		next.ServeHTTP(w, r)
	})
}
