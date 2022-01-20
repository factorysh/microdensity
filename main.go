package main

import (
	"log"
	"net/http"

	"github.com/factorysh/microdensity/middlewares"
	"github.com/factorysh/microdensity/oauth"
	_sessions "github.com/factorysh/microdensity/sessions"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	// get OAuth config from env
	oauthConfig, err := oauth.NewConfigFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	sessions := _sessions.New()

	// routing and handlers
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// protected routes
	r.Group(func(r chi.Router) {
		r.Use(middlewares.OAuth(oauthConfig))
		r.Get("/service/{service}/{project}/latest", func(w http.ResponseWriter, _ *http.Request) {
			w.Write([]byte("i am protected"))
			w.WriteHeader(http.StatusOK)
		})
	})
	// oauth callback hander on /oauth/callback
	r.Get(oauth.CallbackEndpoint, oauth.CallbackHandler(oauthConfig, &sessions))

	r.Get("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("welcome"))
	})

	http.ListenAndServe(":3000", r)
}
