package run

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"os/user"

	"github.com/compose-spec/compose-go/loader"
	"github.com/compose-spec/compose-go/types"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/docker/compose/v2/pkg/compose"
	"github.com/docker/docker/client"
	"github.com/factorysh/microdensity/volumes"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

var _ Runnable = (*ComposeRun)(nil)

type ComposeRun struct {
	home    string
	details *types.ConfigDetails
	service api.Service
	run     string
	name    string
	id      uuid.UUID
	runCtx  context.Context
	project *types.Project
	logger  *zap.Logger
}

func (c *ComposeRun) Id() uuid.UUID {
	return c.id
}

func (c *ComposeRun) Cancel() {
	_, cancelFunc := context.WithCancel(c.runCtx)
	cancelFunc()
}

func NewComposeRun(home string, env map[string]string) (*ComposeRun, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	l := logger.With(zap.String("home", home))

	dockercfg, err := dockerConfig()
	if err != nil {
		l.Error("Docker config drama", zap.Error(err))
		return nil, err
	}

	docker, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		l.Error("Docker client drama", zap.Error(err))
		return nil, err
	}

	ctx := context.TODO()
	_, err = docker.Ping(ctx)
	if err != nil {
		l.Error("Docker doesn't ping", zap.Error(err))
		return nil, err
	}

	project, details, err := LoadCompose(home, env)
	if err != nil {
		l.Error("Load compose error", zap.Error(err))
		return nil, err
	}
	l = l.With(zap.String("project", project.Name))

	srv := compose.NewComposeService(docker, dockercfg)
	grph := compose.NewGraph(project.Services, compose.ServiceStopped)
	roots := grph.Roots()
	if len(roots) == 0 {
		return nil, errors.New("There is no roots")
	}
	if len(roots) > 1 {
		rr := make([]string, len(roots))
		for i, r := range roots {
			rr[i] = r.Service
		}
		err = fmt.Errorf("i need only one root not %v", rr)
		l.Error("Compose graph error", zap.Error(err))
		return nil, err
	}

	// FIXME name is project.Name ?
	_, name := path.Split(strings.TrimSuffix(home, "/"))

	// use default compose network name
	networkName := fmt.Sprintf("%s_default", name)
	err = ensureNetwork(docker, networkName)
	if err != nil {
		l.Error("Ensure network", zap.Error(err))
		return nil, err
	}
	l.Info("Ensure Network", zap.String("name", networkName))

	return &ComposeRun{
		home:    home,
		details: details,
		service: srv,
		run:     roots[0].Service,
		name:    name,
		logger:  logger,
	}, nil

}

// Prepare set a quiet compose environment, waiting for its wake
func (c *ComposeRun) Prepare(envs map[string]string, volumesRoot string, id uuid.UUID, hosts []string) error {
	var err error
	c.id = id
	c.runCtx = context.TODO()
	details := types.ConfigDetails{
		WorkingDir: c.details.WorkingDir,
		ConfigFiles: []types.ConfigFile{
			{
				Filename: c.details.ConfigFiles[0].Filename,
				Content:  c.details.ConfigFiles[0].Content,
			},
		},
		Environment: envs,
	}
	c.project, err = loader.Load(details, func(opt *loader.Options) {
		opt.Name = c.name
		opt.SkipInterpolation = false
	})

	if err != nil {
		c.logger.Error("compose load", zap.Error(err))
		return err
	}

	err = c.PrepareVolumes(volumesRoot)
	if err != nil {
		c.logger.Error("Volumes preparation error", zap.Error(err))
		return err
	}

	err = c.PrepareServices(hosts)
	if err != nil {
		c.logger.Error("Adding hosts to all services", zap.Error(err))
		return err
	}
	/*
		You can watch normalized YAML with
		b := &bytes.Buffer{}
		err = yaml.NewEncoder(b).Encode(project)
		b.String()
	*/
	return nil
}

func (c *ComposeRun) PrepareServices(hosts []string) error {
	services := make(types.Services, len(c.project.Services))
	for i, service := range c.project.Services {
		service.ExtraHosts = append(service.ExtraHosts, hosts...)
		services[i] = service
	}
	c.project.Services = services
	return nil
}

// PrepareVolumes by prepending a custom full path and creating the path on the host
func (c *ComposeRun) PrepareVolumes(prependPath string) error {
	for _, svc := range c.project.Services {
		for i, vol := range svc.Volumes {
			if vol.Type != "bind" {
				continue
			}

			vol.Source = filepath.Join(prependPath, "volumes", vol.Source)
			svc.Volumes[i] = vol

			err := os.MkdirAll(vol.Source, volumes.DirMode)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Run a compose service, writing the STDOUT and STDERR outputs, returns the UNIX return code
func (c *ComposeRun) Run(stdout io.WriteCloser, stderr io.WriteCloser) (int, error) {
	return c.runCommand(stdout, stderr, []string{})
}

// Does this function need to be public?
func (c *ComposeRun) runCommand(stdout io.WriteCloser, stderr io.WriteCloser, commands []string) (int, error) {
	l := c.logger.With(
		zap.String("name", c.project.Name),
		zap.String("service", c.run),
		zap.String("working dir", c.project.WorkingDir),
		zap.String("id", c.id.String()),
	)
	chrono := time.Now()
	u, err := user.Current()
	if err != nil {
		l.Error("Current user")
		return -1, err
	}
	l.With(zap.String("uid", u.Uid))

	err = c.service.Remove(context.TODO(), c.project, api.RemoveOptions{
		Force: true,
	})
	if err != nil {
		l.Error("Remove service", zap.Error(err))
		return -1, err
	}

	defer c.Cancel()
	n, err := c.service.RunOneOffContainer(c.runCtx, c.project, api.RunOptions{
		Name:       fmt.Sprintf("%s_%s_%v", c.project.Name, c.run, c.id),
		Service:    c.run,
		Command:    commands,
		Detach:     false,
		AutoRemove: false, // FIXME: true
		Privileged: false,
		QuietPull:  true,
		Tty:        false,
		Stdin:      os.Stdin,
		Stdout:     stdout,
		Stderr:     stderr,
		User:       u.Uid,
		NoDeps:     false,
		Labels: types.Labels{
			"sh.factory.density.id": c.id.String(),
		},
		Index: 0,
	})
	l = c.logger.With(
		zap.String("id", c.id.String()),
		zap.Int("return code", n),
		zap.Float64("timing µs", float64(time.Since(chrono))/1000),
	)
	if err == nil {
		l.Info("End run")
	} else {
		l.Error("Run error", zap.Error(err))
	}
	return n, err
}

// LoadCompose loads a docker-compose.yml file
func LoadCompose(home string, env map[string]string) (*types.Project, *types.ConfigDetails, error) {
	path := filepath.Clean(filepath.Join(home, "docker-compose.yml"))
	if !strings.HasPrefix(path, home) {
		panic("no path escape: " + path)
	}
	cfg, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}

	raw, err := ioutil.ReadAll(cfg)
	err2 := cfg.Close()
	if err != nil {
		return nil, nil, err
	}
	if err2 != nil {
		return nil, nil, err2
	}

	details := types.ConfigDetails{
		WorkingDir: home,
		ConfigFiles: []types.ConfigFile{
			{
				Filename: path,
				Content:  raw,
			},
		},
		Environment: env,
	}

	p, err := loader.Load(details)

	return p, &details, err
}
