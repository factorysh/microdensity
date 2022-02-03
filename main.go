package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/factorysh/microdensity/application"
	"github.com/factorysh/microdensity/conf"
	"github.com/factorysh/microdensity/version"
)

func main() {
	fmt.Println("Version", version.Version())
	configPath := os.Getenv("CONFIG")
	if configPath == "" {
		configPath = "/etc/microdensity.yml"
	}
	fmt.Println("Config path", configPath)
	cfg, err := conf.Open(configPath)
	if err != nil {
		log.Fatal(err)
	}
	cfg.Defaults()

	a, err := application.NewFromConfig(cfg)
	if err != nil {
		log.Fatal("Application crash", err)
	}
	fmt.Println("Listen", cfg.Listen)
	http.ListenAndServe(cfg.Listen, a.Router)
}
