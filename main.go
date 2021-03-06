package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/factorysh/microdensity/application"
	"github.com/factorysh/microdensity/conf"
	"github.com/factorysh/microdensity/version"
	"go.uber.org/zap"
)

func main() {
	fmt.Println("Version", version.Version())
	configPath := os.Getenv("CONFIG")
	if configPath == "" {
		configPath = "/etc/microdensity.yml"
	}
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	defer logger.Sync()
	l := logger.With(zap.String("config path", configPath))
	cfg, err := conf.Open(configPath)
	if err != nil {
		l.Error("Opening conf", zap.Error(err))
		os.Exit(1)
	}
	cfg.Defaults()
	raw, err := json.Marshal(cfg)
	if err != nil {
		l.Error("Marshalling conf", zap.Error(err))
		os.Exit(1)
	}
	var cfgPub conf.Conf
	err = json.Unmarshal(raw, &cfgPub)
	if err != nil {
		l.Error("Marshalling conf", zap.Error(err))
		os.Exit(1)
	}
	cfgPub.OAuth.AppSecret = "•••"

	l = l.With(zap.Any("config", cfgPub))

	a, err := application.New(cfg)
	if err != nil {
		l.Error("Application", zap.Error(err))
		os.Exit(1)
	}

	logger.Info("starting")
	err = a.Run(cfg.Listen)
	if err != nil {
		l.Error("Run", zap.Error(err))
		os.Exit(1)
	}
	if cfg.AdminListen == "" {
		cfg.AdminListen = "127.0.0.1:3615"
	}
	err = a.AdminRun(cfg.AdminListen)
	if err != nil {
		l.Error("AdminRun", zap.Error(err))
		os.Exit(1)
	}

	<-a.Stopper

	logger.Info("shutdown signal received")

	err = a.Shutdown()
	if err != nil {
		l.Error("error on shutdown", zap.Error(err))
	}
}
