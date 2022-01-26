package application

import (
	"github.com/factorysh/microdensity/middlewares"
	"github.com/factorysh/microdensity/queue"
	"github.com/factorysh/microdensity/service"
	"github.com/factorysh/microdensity/volumes"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Application struct {
	Services []service.Service
	queue    *queue.Queue
	router   chi.Router
	volumes  *volumes.Volumes
}

func New(q *queue.Queue, secret string) (*Application, error) {
	r := chi.NewRouter()
	a := &Application{
		Services: make([]service.Service, 0),
		queue:    q,
		router:   r,
	}
	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(middlewares.Tokens())
	r.Use(middlewares.Auth(secret))

	r.Get("/services", a.services)
	r.Route("/service/{serviceID}", func(r chi.Router) {
		r.Use(a.serviceMiddleware)
		r.Get("/", a.service)
		r.Post("/{project}/{branch}/{commit}", a.newTask)
		r.Get("/-{taskID}", a.task)
		r.Route("/{project}", func(r chi.Router) {
			r.Route("/{branch}", func(r chi.Router) {
				r.Route("/{commit}", func(r chi.Router) {
					r.Get("/", nil)
				})
				r.Get("/latest", nil)
			})
		})
	})

	return a, nil
}
