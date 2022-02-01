package application

import (
	"fmt"
	"os"

	"github.com/factorysh/microdensity/conf"
	"github.com/factorysh/microdensity/middlewares/jwt"
	jwtoroauth2 "github.com/factorysh/microdensity/middlewares/jwt_or_oauth2"
	"github.com/factorysh/microdensity/queue"
	"github.com/factorysh/microdensity/service"
	"github.com/factorysh/microdensity/sessions"
	"github.com/factorysh/microdensity/volumes"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

type Application struct {
	Services []service.Service
	Router   chi.Router
	queue    *queue.Storage
	router   chi.Router
	volumes  *volumes.Volumes
	logger   *zap.Logger
}

func New(q *queue.Storage, secret string, volumePath string) (*Application, error) {
	sessions := sessions.New()
	err := sessions.Start(15)
	if err != nil {
		return nil, err
	}

	jwtProvider := os.Getenv("JWT_PROVIDER_URL")
	if jwtProvider == "" {
		return nil, fmt.Errorf("missing JWT_PROVIDER_URL environment variable")
	}

	jwtAuth, err := jwt.NewJWTAuthenticator(jwtProvider)
	if err != nil {
		return nil, err
	}

	oAuthConfig, err := conf.NewOAuthConfigFromEnv()
	if err != nil {
		return nil, err
	}

	v, err := volumes.New(volumePath)
	if err != nil {
		return nil, err
	}

	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}

	r := chi.NewRouter()

	a := &Application{
		Services: make([]service.Service, 0),
		queue:    q,
		Router:   r,
		volumes:  v,
		logger:   logger,
	}
	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	authMiddleware, err := jwtoroauth2.NewJWTOrOauth2(jwtAuth, oAuthConfig, &sessions)
	if err != nil {
		return nil, err
	}
	r.Use(authMiddleware.Middleware())

	r.Get("/services", a.services)
	r.Route("/service/{serviceID}", func(r chi.Router) {
		r.Use(a.serviceMiddleware)
		r.Get("/", a.service)
		r.Post("/{project}/{branch}/{commit}", a.newTask)
		r.Get("/-{taskID}", a.task)
		r.Route("/{project}", func(r chi.Router) {
			r.Route("/{branch}", func(r chi.Router) {
				r.Route("/{commit}", func(r chi.Router) {
					r.Get("/", a.task)
				})
				r.Get("/latest", nil)
			})
		})
	})

	return a, nil
}
