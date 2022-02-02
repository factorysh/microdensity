package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/factorysh/microdensity/application"
	"github.com/factorysh/microdensity/conf"
	"github.com/factorysh/microdensity/middlewares/jwt"
	"github.com/factorysh/microdensity/queue"
	"github.com/factorysh/microdensity/version"
	"go.etcd.io/bbolt"
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

	jwtAuth, err := jwt.NewJWTAuthenticator(cfg.JWKProvider)
	if err != nil {
		log.Fatal(err)
	}

	storePath := path.Join(cfg.Queue, "microdensity.store")
	fmt.Println("bbolt path", storePath)
	s, err := bbolt.Open(
		storePath,
		0600, &bbolt.Options{})
	if err != nil {
		log.Fatal("bbolt error", err)
	}
	q, err := queue.New(s)
	if err != nil {
		log.Fatal("Queue error", err)
	}
	// FIXME: path
	a, err := application.New(q, &cfg.OAuth, jwtAuth, "/tmp/microdensity")
	if err != nil {
		log.Fatal("Application crash", err)
	}
	fmt.Println("Listen", cfg.Listen)
	http.ListenAndServe(cfg.Listen, a.Router)
}
