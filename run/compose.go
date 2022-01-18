package run

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/docker/cli/cli/compose/loader"
)

type ComposeRun struct {
	home string
}

func NewComposeRun(home string) (*ComposeRun, error) {
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
	fmt.Println(yml)

	if err != nil {
		return nil, err
	}
	return &ComposeRun{
		home: home,
	}, nil
}

func (c *ComposeRun) Run(args map[string]interface{}) {

}
