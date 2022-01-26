package run

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

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
	"golang.org/x/net/context"
)

var _ Runnable = (*ComposeRun)(nil)

type ComposeRun struct {
	home    string
	details types.ConfigDetails
	service api.Service
	run     string
	name    string
	id      uuid.UUID
	runCtx  context.Context
	project *types.Project
}

func dockerConfig() (*configfile.ConfigFile, error) {
	pth := path.Join(os.Getenv("HOME"), "/.docker/config.json")
	_, err := os.Stat(pth)
	dockercfg := &configfile.ConfigFile{}
	if err != nil {
		if os.IsNotExist(err) {
			f, err := os.Open(pth)
			if err != nil {
				return nil, err
			}
			defer f.Close()
			err = dockercfg.LoadFromReader(f)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		dockercfg = configfile.New(pth)
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
	dockercfg, err := dockerConfig()
	if err != nil {
		return nil, err
	}

	docker, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}

	_, err = docker.Ping(context.TODO())
	if err != nil {
		return nil, err
	}

	cfg, err := os.Open(path.Join(home, "docker-compose.yml"))
	if err != nil {
		return nil, err
	}
	defer cfg.Close()
	raw, err := ioutil.ReadAll(cfg)
	if err != nil {
		return nil, err
	}

	details := types.ConfigDetails{
		WorkingDir: home,
		ConfigFiles: []types.ConfigFile{
			{
				Filename: home,
				Content:  raw,
			},
		},
		Environment: map[string]string{},
	}

	srv := compose.NewComposeService(docker, dockercfg)
	project, err := loader.Load(details)
	if err != nil {
		return nil, err
	}
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
		return nil, fmt.Errorf("i need only one root not %v", rr)
	}

	_, name := path.Split(strings.TrimSuffix(home, "/"))

	// use default compose network name
	ensureNetwork(docker, fmt.Sprintf("%s_default", name))

	return &ComposeRun{
		home:    home,
		details: details,
		service: srv,
		run:     roots[0].Service,
		name:    name,
	}, nil

}

func (c *ComposeRun) Prepare(args map[string]string, volumesRoot string) error {
	var err error
	c.id, err = uuid.NewUUID()
	if err != nil {
		return err
	}
	c.runCtx = context.TODO()
	details := types.ConfigDetails{
		WorkingDir: c.details.WorkingDir,
		ConfigFiles: []types.ConfigFile{
			{
				Filename: c.details.ConfigFiles[0].Filename,
				Content:  c.details.ConfigFiles[0].Content,
			},
		},
		Environment: args,
	}
	c.project, err = loader.Load(details, func(opt *loader.Options) {
		opt.Name = c.name
		opt.SkipInterpolation = false
	})

	c.PrepareVolumes(volumesRoot)

	if err != nil {
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

func (c *ComposeRun) Run(stdout io.WriteCloser, stderr io.WriteCloser) (int, error) {
	return c.service.RunOneOffContainer(c.runCtx, c.project, api.RunOptions{
		Name:       fmt.Sprintf("%s_%s_%v", c.project.Name, c.run, c.id),
		Service:    c.run,
		Detach:     false,
		AutoRemove: true,
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
}

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
