package run

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/compose-spec/compose-go/loader"
	"github.com/compose-spec/compose-go/types"
	"github.com/docker/cli/cli/config/configfile"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/docker/compose/v2/pkg/compose"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
)

type ComposeRun struct {
	home    string
	details types.ConfigDetails
	service api.Service
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
	yml, err := loader.ParseYAML(raw)
	if err != nil {
		return nil, err
	}

	details := types.ConfigDetails{
		WorkingDir: home,
		ConfigFiles: []types.ConfigFile{
			{
				Filename: home,
				Config:   yml,
			},
		},
		Environment: map[string]string{},
	}

	return &ComposeRun{
		home:    home,
		details: details,
		service: compose.NewComposeService(docker, dockercfg),
	}, nil

}

func (c *ComposeRun) Run(ctx context.Context, args map[string]interface{}) error {
	project, err := loader.Load(c.details)
	if err != nil {
		return err
	}

	return c.service.Up(ctx, project, api.UpOptions{
		Create: api.CreateOptions{
			RemoveOrphans: true,
		},
		Start: api.StartOptions{
			Wait: false,
		},
	})
}
