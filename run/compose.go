package run

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/compose-spec/compose-go/loader"
	"github.com/compose-spec/compose-go/types"
	"github.com/docker/cli/cli/config/configfile"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/docker/compose/v2/pkg/compose"
	"github.com/docker/docker/client"
	"github.com/google/uuid"
	"golang.org/x/net/context"
)

type ComposeRun struct {
	home    string
	details types.ConfigDetails
	service api.Service
	run     string
	name    string
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

	return &ComposeRun{
		home:    home,
		details: details,
		service: srv,
		run:     roots[0].Service,
		name:    name,
	}, nil

}

func (c *ComposeRun) Run(ctx context.Context, args map[string]string, stdout io.WriteCloser, stderr io.WriteCloser) (int, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return 0, err
	}
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
	project, err := loader.Load(details, func(opt *loader.Options) {
		opt.Name = c.name
		opt.SkipInterpolation = false
	})
	if err != nil {
		return 0, err
	}
	/*
		You can watch normalized YAML with
		b := &bytes.Buffer{}
		err = yaml.NewEncoder(b).Encode(project)
		b.String()
	*/
	return c.service.RunOneOffContainer(ctx, project, api.RunOptions{
		Name:       fmt.Sprintf("%s_%s_%v", project.Name, c.run, id),
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
			"sh.factory.density.id": id.String(),
		},
		Index: 0,
	})
}
