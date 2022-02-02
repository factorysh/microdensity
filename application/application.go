package application

import (
	"os"

	"github.com/factorysh/microdensity/conf"
	"github.com/factorysh/microdensity/middlewares/jwt"
	jwtoroauth2 "github.com/factorysh/microdensity/middlewares/jwt_or_oauth2"
	"github.com/factorysh/microdensity/queue"
	"github.com/factorysh/microdensity/service"
	"github.com/factorysh/microdensity/sessions"
	"github.com/factorysh/microdensity/volumes"
	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/tchap/zapext/v2/zapsentry"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Application struct {
	Services map[string]service.Service
	Router   chi.Router
	queue    *queue.Storage
	router   chi.Router
	volumes  *volumes.Volumes
	logger   *zap.Logger
}

func New(q *queue.Storage, oAuthConfig *conf.OAuthConf, jwtAuth *jwt.JWTAuthenticator, volumePath string) (*Application, error) {
	var logger *zap.Logger
	var err error
	dsn := os.Getenv("SENTRY_DSN")
	if dsn != "" {
		client, err := sentry.NewClient(sentry.ClientOptions{Dsn: dsn})
		if err != nil {
			return nil, err
		}
		logger = zap.New(zapsentry.NewCore(zapcore.ErrorLevel, client))
		logger.Info("Sentry is set")
	} else {
		logger, err = zap.NewProduction()
		if err != nil {
			return nil, err
		}
		logger.Info("There is no Sentry set")
	}
	sessions := sessions.New()
	err = sessions.Start(15)
	if err != nil {
		logger.Error("Session crash", zap.Error(err))
		return nil, err
	}

	v, err := volumes.New(volumePath)
	if err != nil {
		logger.Error("Volumes crash", zap.Error(err))
		return nil, err
	}

	r := chi.NewRouter()

	a := &Application{
		Services: make(map[string]service.Service),
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
		logger.Error("JWT or OAth2 middleware crash", zap.Error(err))
		return nil, err
	}
	r.Get("/", HomeHandler)

	r.Get("/services", a.services)
	r.Route("/service/{serviceID}", func(r chi.Router) {
		r.Use(authMiddleware.Middleware())
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
