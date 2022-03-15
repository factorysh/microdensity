package application

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// ImageParams describe body data
type ImageParams struct {
	Name      string `json:"name"`
	AuthToken string `json:"auth_token"`
}

// PostImageHandler is used to pull the request image
func (a *Application) PostImageHandler(w http.ResponseWriter, r *http.Request) {
	var imageParams ImageParams

	serviceID := chi.URLParam(r, "serviceID")
	project := chi.URLParam(r, "project")
	l := a.logger.With(
		zap.String("service", serviceID),
		zap.String("project", project),
	)

	d := json.NewDecoder(r.Body)
	err := d.Decode(&imageParams)
	if err != nil {
		l.Warn("decoding post image handler params error", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		l.Warn("new client error", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	out, err := cli.ImagePull(r.Context(), imageParams.Name, types.ImagePullOptions{RegistryAuth: imageParams.AuthToken})
	if err != nil {
		l.Warn("error when downloading image", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer out.Close()

	text, err := ioutil.ReadAll(out)
	if err != nil {
		l.Warn("error when downloading image", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("content-type", "text/plain")
	w.Write(text)
}
