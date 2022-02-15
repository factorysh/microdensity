package application

import (
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/factorysh/microdensity/badge"
	"github.com/factorysh/microdensity/conf"
	"github.com/factorysh/microdensity/middlewares/jwt"
	jwtoroauth2 "github.com/factorysh/microdensity/middlewares/jwt_or_oauth2"
	"github.com/factorysh/microdensity/oauth"
	"github.com/factorysh/microdensity/queue"
	"github.com/factorysh/microdensity/run"
	"github.com/factorysh/microdensity/service"
	"github.com/factorysh/microdensity/sessions"
	"github.com/factorysh/microdensity/storage"
	"github.com/factorysh/microdensity/volumes"
	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tchap/zapext/v2/zapsentry"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Application struct {
	Services      map[string]service.Service
	serviceFolder string
	Domain        string
	GitlabDomain  string
	Router        http.Handler
	storage       storage.Storage
	volumes       *volumes.Volumes
	logger        *zap.Logger
	queue         *queue.Queue
}

func New(cfg *conf.Conf) (*Application, error) {
	jwtAuth, err := jwt.NewJWTAuthenticator(cfg.JWKProvider)
	if err != nil {
		return nil, err
	}

	storePath := path.Join(cfg.DataPath)
	s, err := storage.NewFSStore(storePath)
	if err != nil {
		return nil, err
	}

	var logger *zap.Logger
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

	v, err := volumes.New(cfg.DataPath)
	if err != nil {
		logger.Error("Volumes crash", zap.Error(err))
		return nil, err
	}

	svcs, err := loadServices(cfg.Services)
	if err != nil {
		logger.Error("Services crash", zap.Error(err))
		return nil, err
	}
	logger.Info("Load services",
		zap.String("service path", cfg.Services),
		zap.Any("services", svcs))

	runner, err := run.NewRunner(cfg.Services, cfg.DataPath, cfg.Hosts)
	if err != nil {
		logger.Error("Runner crash", zap.Error(err))
		return nil, err
	}

	q := queue.NewQueue(s, runner)

	r := chi.NewRouter()

	a := &Application{
		Services: svcs,
		Domain:   cfg.OAuth.AppURL,
		// FIXME: use dedicated variable
		GitlabDomain:  cfg.OAuth.ProviderURL,
		serviceFolder: cfg.Services,
		storage:       s,
		Router:        MagicPathHandler(r),
		volumes:       v,
		logger:        logger,
		queue:         &q,
	}
	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	authMiddleware, err := jwtoroauth2.NewJWTOrOauth2(jwtAuth, &cfg.OAuth, &sessions)
	if err != nil {
		logger.Error("JWT or OAth2 middleware crash", zap.Error(err))
		return nil, err
	}
	r.Get("/", a.HomeHandler())
	r.Get("/robots.txt", RobotsHandler)
	r.Get("/favicon.png", FaviconHandler)
	r.Get("/metrics", promhttp.Handler().ServeHTTP)
	r.Get("/oauth/callback", oauth.CallbackHandler(&cfg.OAuth, &sessions))

	r.Get("/services", a.ServicesHandler)
	r.Get("/service/{serviceID}", a.ReadmeHandler)
	r.Route("/service/{serviceID}/{project}", func(r chi.Router) {
		r.Use(authMiddleware.Middleware())
		r.Use(a.ServiceMiddleware)
		r.Route("/", func(r chi.Router) {
			r.Route("/{branch}", func(r chi.Router) {
				r.Route("/{commit}", func(r chi.Router) {
					r.Post("/", a.PostTaskHandler)
					r.Get("/", a.TaskHandler(false))
					r.Get("/status", badge.StatusBadge(a.storage, false))
					r.Get("/badge/{badge}", a.BadgeMyTaskHandler(false))
					r.Get("/volumes/*", a.VolumesHandler(6, false))
				})
				r.Route("/latest", func(r chi.Router) {
					r.Get("/", a.TaskHandler(true))
					r.Get("/status", badge.StatusBadge(a.storage, true))
					r.Get("/badge/{badge}", a.BadgeMyTaskHandler(true))
					r.Get("/volumes/*", a.VolumesHandler(6, true))
				})
			})
		})
	})

	return a, nil
}

// loop into all services sub dirs and create a service using files found
func loadServices(path string) (map[string]service.Service, error) {
	svcs := make(map[string]service.Service)

	subs, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, sub := range subs {
		svc, err := service.NewFolder(filepath.Join(path, sub.Name()))
		if err != nil {
			return nil, err
		}
		svcs[sub.Name()] = svc
	}

	return svcs, nil
}

// ListServices returns a list of all services as string array
func (a *Application) ListServices() []string {
	list := make([]string, len(a.Services))

	i := 0
	for key := range a.Services {
		list[i] = key
		i++
	}

	return list
}
