package conf

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Conf struct {
	Issuer      string    // FIXME what the hell is an issuer?
	OAuth       OAuthConf `yaml:"OAuth"`
	Services    string    `yaml:"services"` // Service folder
	JWKProvider string    `yaml:"jwk_provider"`
	Listen      string    `yaml:"listen"` // http listen address
	AdminListen string    `yaml:"admin_listen"`
	DataPath    string    `yaml:"data_path"`
	Hosts       []string  `yaml:"hosts"` // private hostnames for exposing private services, like browserless
}

func (c *Conf) Defaults() {
	if c.Services == "" {
		c.Services = "/var/lib/microdensity/services"
	}
	if c.DataPath == "" {
		c.DataPath = "/var/lib/microdensity/data"
	}
	if c.Listen == "" {
		c.Listen = "127.0.0.1:3000"
	}
}

func Open(path string) (*Conf, error) {
	path = filepath.Clean(path)
	f, err := os.Open(path) //#nosec
	if err != nil {
		return nil, err
	}
	var cfg Conf
	err = yaml.NewDecoder(f).Decode(&cfg)
	err2 := f.Close()
	if err != nil {
		return nil, err
	}
	if err2 != nil {
		return nil, err2
	}
	return &cfg, nil
}
