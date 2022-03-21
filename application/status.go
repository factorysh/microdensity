package application

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"go.uber.org/zap"
)

type gitlabStatus struct {
	Url    string `json:"url"`
	Status int
}

type Status struct {
	Ping   types.Ping `json:"docker"`
	Gitlab *gitlabStatus
}

func (a *Application) StatusHandler(w http.ResponseWriter, r *http.Request) {
	docker, err := client.NewEnvClient()
	if err != nil {
		a.logger.With(zap.Error(err)).Error("Status: docker error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	ctx, cancel := context.WithTimeout(context.TODO(), 3*time.Second)
	defer cancel()
	ping, err := docker.Ping(ctx)
	if err != nil {
		a.logger.With(zap.Error(err)).Error("Status: docker ping error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	status := &Status{
		Ping: ping,
	}

	resp, err := http.Get(a.GitlabURL)
	if err != nil {
		a.logger.With(zap.Error(err), zap.String("gitlab", a.GitlabURL)).Error("Status: Gitlab ping error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	status.Gitlab = &gitlabStatus{
		Url:    a.GitlabURL,
		Status: resp.StatusCode,
	}

	w.Header().Set("content-type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(status)

}
