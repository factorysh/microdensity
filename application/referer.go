package application

import "net/http"

// RefererMiddleware ensure that requests comes from the gitlab domain
func (a *Application) RefererMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		referer := r.Referer()
		if referer != a.GitlabURL {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		next.ServeHTTP(w, r)
	})
}
