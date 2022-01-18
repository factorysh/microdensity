package run

import (
	"fmt"
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
	docker := &client.Client{}
	err = client.FromEnv(docker)
	if err != nil {
		return nil, err
	}

	service := compose.NewComposeService(docker, dockercfg)
	fmt.Println(service)

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
	fmt.Println(yml)

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

	cr := &ComposeRun{
		home:    home,
		details: details,
		service: service,
	}

	return cr, nil
}

func (c *ComposeRun) Run(ctx context.Context, args map[string]interface{}) error {
	project, err := loader.Load(c.details)
	if err != nil {
		return err
	}

	return c.service.Up(ctx, project, api.UpOptions{})
}
