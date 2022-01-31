package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/factorysh/microdensity/httpcontext"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestProject(t *testing.T) {
	route := chi.NewRouter()
	yes := false
	route.Use(
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if yes {
					fmt.Println("yes", yes)
					ctx := context.WithValue(r.Context(), httpcontext.RequestedProject, "bob")
					r = r.WithContext(ctx)
				}
				next.ServeHTTP(w, r)
			})
		})

	route.Route("/service/{project}", func(route chi.Router) {
		route.Use(AssertProject)
		route.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			fmt.Println("handler for", r.URL)
			fmt.Fprintf(w, "Hello %s", chi.URLParam(r, "project"))
		})
	})

	srv := httptest.NewServer(route)
	defer srv.Close()

	cli := &http.Client{}
	res, err := cli.Get(fmt.Sprintf("%s/service/bob", srv.URL))
	assert.NoError(t, err)
	assert.Equal(t, 403, res.StatusCode)

	yes = true
	res, err = cli.Get(fmt.Sprintf("%s/service/bob", srv.URL))
	assert.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)
}
