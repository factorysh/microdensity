package run

import (
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
	"github.com/docker/cli/cli/config/configfile"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/docker/compose/v2/pkg/compose"
	dtypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
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

func dockerConfig() (*configfile.ConfigFile, error) {
	dockercfg := &configfile.ConfigFile{}
	var home string
	me, err := user.Current()
	if err != nil {
		// In docker container, you can -u an unknown user
		home = os.TempDir()
	} else {
		home = me.HomeDir
	}

	pth := path.Join(home, "/.docker/config.json")
	_, err = os.Stat(pth)
	if err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(path.Join(home, "/.docker"), 0700)
			if err != nil {
				return nil, err
			}
			dockercfg = configfile.New(pth)
		} else {
			return nil, err
		}
	} else {
		f, err := os.Open(pth)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		err = dockercfg.LoadFromReader(f)
		if err != nil {
			return nil, err
		}
	}
	return dockercfg, nil
}

func (c *ComposeRun) Id() uuid.UUID {
	return c.id
}

func (c *ComposeRun) Cancel() {
	_, cancelFunc := context.WithCancel(c.runCtx)
	cancelFunc()
}

func NewComposeRun(home string) (*ComposeRun, error) {
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

	_, err = docker.Ping(context.TODO())
	if err != nil {
		l.Error("Docker doesn't ping", zap.Error(err))
		return nil, err
	}

	project, details, err := LoadCompose(home)
	if err != nil {
		l.Error("Load compose error", zap.Error(err))
		return nil, err
	}
	l = l.With(zap.String("project", project.Name))

	srv := compose.NewComposeService(docker, dockercfg)
	grph := compose.NewGraph(project.Services, compose.ServiceStopped)
	roots := grph.Roots()
	if len(roots) == 0 {
		panic("There is no roots")
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
	ensureNetwork(docker, networkName)
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
func (c *ComposeRun) Prepare(envs map[string]string, volumesRoot string, id uuid.UUID) error {
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

	/*
		You can watch normalized YAML with
		b := &bytes.Buffer{}
		err = yaml.NewEncoder(b).Encode(project)
		b.String()
	*/
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
	c.logger.Info("Run service",
		zap.String("name", c.project.Name),
		zap.String("service", c.run),
		zap.String("working dir", c.project.WorkingDir),
		zap.String("id", c.id.String()))
	chrono := time.Now()
	n, err := c.service.RunOneOffContainer(c.runCtx, c.project, api.RunOptions{
		Name:    fmt.Sprintf("%s_%s_%v", c.project.Name, c.run, c.id),
		Service: c.run,
		Detach:  false,
		// FIXME: true
		AutoRemove: false,
		Privileged: false,
		QuietPull:  true,
		Tty:        false,
		Stdin:      os.Stdin,
		Stdout:     stdout,
		Stderr:     stderr,
		NoDeps:     false,
		Labels: types.Labels{
			"sh.factory.density.id": c.id.String(),
		},
		Index: 0,
	})
	l := c.logger.With(
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

// Lazy network creation
func ensureNetwork(cli *client.Client, networkName string) error {
	networks, err := cli.NetworkList(context.Background(), dtypes.NetworkListOptions{
		Filters: filters.NewArgs(filters.KeyValuePair{
			Key:   "name",
			Value: networkName,
		},
		)})

	if err != nil {
		return err
	}

	if len(networks) == 0 {
		_, err = cli.NetworkCreate(context.Background(), networkName, dtypes.NetworkCreate{})
		if err != nil {
			return err

		}
	}

	return err
}

// LoadCompose loads a docker-compose.yml file
func LoadCompose(home string) (*types.Project, *types.ConfigDetails, error) {
	path := filepath.Join(home, "docker-compose.yml")
	cfg, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer cfg.Close()

	raw, err := ioutil.ReadAll(cfg)
	if err != nil {
		return nil, nil, err
	}

	details := types.ConfigDetails{
		WorkingDir: home,
		ConfigFiles: []types.ConfigFile{
			{
				Filename: path,
				Content:  raw,
			},
		},
		Environment: map[string]string{},
	}

	p, err := loader.Load(details)

	return p, &details, err
}
