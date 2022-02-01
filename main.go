package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/factorysh/microdensity/application"
	"github.com/factorysh/microdensity/conf"
	"github.com/factorysh/microdensity/queue"
	_sessions "github.com/factorysh/microdensity/sessions"
	"github.com/factorysh/microdensity/version"
	"go.etcd.io/bbolt"
)

func main() {
	fmt.Println("Version", version.Version())
	configPath := os.Getenv("CONFIG")
	if configPath == "" {
		configPath = "/etc/microdensity.yml"
	}
	cfg, err := conf.Open(configPath)
	if err != nil {
		log.Fatal(err)
	}
	cfg.Defaults()

	sessions := _sessions.New()
	sessions.Start(15)

	s, err := bbolt.Open(
		path.Join(cfg.Queue, "microdensity.store"),
		0600, &bbolt.Options{})
	if err != nil {
		log.Fatal(err)
	}
	q, err := queue.New(s)
	if err != nil {
		log.Fatal(err)
	}
	a, err := application.New(q, cfg.JWTSecret, cfg.Services)
	if err != nil {
		log.Fatal(err)
	}
	http.ListenAndServe("127.0.0.1:3000", a.Router)
}
