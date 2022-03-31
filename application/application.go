package application

/*
application.Application is the hub of µdensity.
All routes are declared here.
*/

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"syscall"
	"time"

	"github.com/docker/go-events"
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
	"github.com/factorysh/microdensity/task"
	"github.com/factorysh/microdensity/volumes"
	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tchap/zapext/v2/zapsentry"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/semaphore"
)

type Application struct {
	Services      map[string]service.Service
	serviceFolder string
	Domain        string
	GitlabURL     string
	Router        http.Handler
	AdminRouter   http.Handler
	storage       storage.Storage
	volumes       *volumes.Volumes
	logger        *zap.Logger
	queue         *queue.Queue
	Sink          *events.Broadcaster
	Server        *http.Server
	AdminServer   *http.Server
	PruneLock     semaphore.Weighted
	Stopper       chan (os.Signal)
}

func New(cfg *conf.Conf) (*Application, error) {
	err := service.ValidateServicesDefinitions(cfg.Services)
	if err != nil {
		return nil, err
	}

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

	ar := chi.NewRouter()
	ar.Use(middleware.Logger)
	ar.Use(middleware.Recoverer)
	ar.Get("/", AdminHomeHandler)
	ar.Get("/metrics", promhttp.Handler().ServeHTTP)
	ar.Get("/robots.txt", RobotsHandler)
	ar.Get("/favicon.png", FaviconHandler)

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

	_sink := events.NewBroadcaster()
	q := queue.NewQueue(s, runner, _sink)

	r := chi.NewRouter()

	a := &Application{
		PruneLock: *semaphore.NewWeighted(int64(1)),
		Services:  svcs,
		Domain:    cfg.OAuth.AppURL,
		// FIXME: use dedicated variable
		GitlabURL:     cfg.OAuth.ProviderURL,
		serviceFolder: cfg.Services,
		storage:       s,
		Router:        MagicPathHandler(r), // Convert service/demo/path/to/project- to service/demo/path%2fto%2fproject/
		AdminRouter:   ar,
		volumes:       v,
		logger:        logger,
		queue:         &q,
		Sink:          _sink,
		Stopper:       make(chan os.Signal, 1),
	}
	ar.Get("/status", a.StatusHandler)
	ar.Get("/sink", a.SinkAllHandler)
	ar.Post("/prune", a.PruneHandler)
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
	r.Get("/oauth/callback", oauth.CallbackHandler(&cfg.OAuth, &sessions))

	r.Get("/services", a.ServicesHandler)
	r.Get("/service/{serviceID}", a.ReadmeHandler)
	r.Route("/service/{serviceID}/{project}", func(r chi.Router) {
		r.Use(a.ServiceMiddleware)
		r.Route("/", func(r chi.Router) {
			r.Route("/{branch}", func(r chi.Router) {
				r.Route("/{commit}", func(r chi.Router) {
					r.Group(func(r chi.Router) {
						r.Use(authMiddleware.Middleware())
						r.Post("/", a.PostTaskHandler)
						r.Post("/_image", a.PostImageHandler)
						r.Get("/", a.TaskHandler(false))
						r.Get("/volumes/*", a.VolumesHandler(6, false))
						r.Get("/logs", a.TaskLogsHandler(false))
					})
					r.Group(func(r chi.Router) {
						r.Use(a.RefererMiddleware)
						r.Get("/status", badge.StatusBadge(a.storage, false)) // status of this task
						r.Get("/badge/{badge}", a.BadgeMyTaskHandler(false))  // badge wrote by docker run
					})
				})
				r.Route("/latest", func(r chi.Router) { // alias to latest run
					r.Group(func(r chi.Router) {
						r.Use(authMiddleware.Middleware())
						r.Get("/", a.TaskHandler(true))
						r.Get("/volumes/*", a.VolumesHandler(6, true))
						r.Get("/logs", a.TaskLogsHandler(true))
					})
					r.Group(func(r chi.Router) {
						r.Use(a.RefererMiddleware)
						r.Get("/status", badge.StatusBadge(a.storage, true))
						r.Get("/badge/{badge}", a.BadgeMyTaskHandler(true))
					})
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

// Run make the app listen and serve requests
func (a *Application) Run(listen string) error {
	// listen for stop/restart signals and sends them to the stopper channel
	signal.Notify(a.Stopper, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// setup the router
	a.Server = &http.Server{
		Addr:    listen,
		Handler: a.Router,
	}

	// interrupted tasks becomes ready task and are added to queue
	tasks, err := a.storage.All()
	if err != nil {
		return err
	}

	for _, t := range tasks {
		if t.State == task.Ready || t.State == task.Interrupted {
			t.State = task.Ready
			parsedArgs, err := a.Services[t.Service].Validate(t.Args)
			// non blocking error
			if err != nil {
				t.State = task.Failed
				err := a.storage.Upsert(t)
				if err != nil {
					a.logger.Fatal("unable to save task", zap.Error(err))
				}

				a.logger.Error("error when validating task args", zap.String("task", t.Id.String()))
				continue
			}

			err = a.addTask(t, parsedArgs.Environments)
			// non blocking error
			if err != nil {
				a.logger.Error("error when adding task", zap.Error(err))
			}
		}
	}

	// start and serve
	go func() {
		if err := a.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Fatal("fatal error when running server", zap.Error(err))
		}
	}()

	return nil
}

func (a *Application) AdminRun(listen string) error {
	a.AdminServer = &http.Server{
		Addr:    listen,
		Handler: a.AdminRouter,
	}

	// start and serve
	go func() {
		if err := a.AdminServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Fatal("fatal error when running server", zap.Error(err))
		}
	}()
	return nil
}

// Shutdown the server and put running tasks into interrupted state
func (a *Application) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// handle the last pending requests
	// we don't want this error stop the shutdown flow
	// just log it
	err := a.Server.Shutdown(ctx)
	if err != nil {
		a.logger.Error("error on server shutdown", zap.Error(err))
	}

	tasks, err := a.storage.All()
	if err != nil {
		return err
	}

	for _, t := range tasks {
		// running tasks becomes interrupted tasks
		if t.State == task.Running {
			// TODO: send a cancel request to docker ?
			t.State = task.Interrupted
			err := a.storage.Upsert(t)
			// same here, non blocking error
			if err != nil {
				a.logger.Error("error when updating task status", zap.Error(err), zap.String("task id", t.Id.String()))
			}
		}
	}

	a.logger.Info("server shutdown")

	return nil
}
