package conf

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Conf struct {
	Issuer      string
	OAuth       OAuthConf `yaml:"OAuth"`
	Services    string    `yaml:"service"` // Service folder
	Queue       string    `yaml:"queue"`   // Queue path
	JWKProvider string    `yaml:"jwk_provider"`
}

func (c *Conf) Defaults() {
	if c.Services == "" {
		c.Services = "/var/lib/microdensity/services"
	}
	if c.Queue == "" {
		c.Queue = "/var/lib/microdensity/queue"
	}
}

func Open(path string) (*Conf, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var cfg Conf
	err = yaml.NewDecoder(f).Decode(&cfg)
	return &cfg, err
}
