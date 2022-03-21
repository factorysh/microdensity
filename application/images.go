package application

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

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

// PostImageHandler is used to pull the request image, see https://docs.docker.com/engine/api/sdk/examples/#pull-an-image-with-authentication
func (a *Application) PostImageHandler(w http.ResponseWriter, r *http.Request) {
	var imageParams ImageParams

	serviceID := chi.URLParam(r, "serviceID")
	project := chi.URLParam(r, "project")
	commit := chi.URLParam(r, "commit")
	l := a.logger.With(
		zap.String("service", serviceID),
		zap.String("project", project),
		zap.String("commit", commit),
	)

	d := json.NewDecoder(r.Body)
	err := d.Decode(&imageParams)
	if err != nil {
		l.Warn("decoding post image handler params error", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if err := verifyImageName(imageParams.Name, project, commit); err != nil {
		l.Warn("verifying image name error", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	authConfig := types.AuthConfig{
		Username: "gitlab-ci-token",
		Password: imageParams.AuthToken,
	}
	encoded, err := json.Marshal(authConfig)
	if err != nil {
		l.Warn("encoding auth config", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	auth := base64.URLEncoding.EncodeToString(encoded)
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		l.Warn("new client error", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	ended := make(chan bool)
	f, flushable := w.(http.Flusher)
	if flushable {
		go func() {
			ticker := time.NewTicker(1 * time.Second)
			for {
				select {
				case <-ticker.C:
					io.WriteString(w, "#")
					if flushable {
						f.Flush()
					}
				case <-ended:
					return
				case <-r.Context().Done():
					return
				}
			}
		}()
	}

	out, err := cli.ImagePull(r.Context(), imageParams.Name, types.ImagePullOptions{RegistryAuth: auth})
	if err != nil {
		l.Warn("error when downloading image", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer out.Close()

	// ⚠ consume the reader or nothing happens
	_, err = ioutil.ReadAll(out)
	if flushable {
		ended <- true
	}
	if err != nil {
		l.Warn("error when downloading image", zap.Error(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Length", "0")
	io.WriteString(w, " Download ended")
	w.WriteHeader(http.StatusOK)
}

func verifyImageName(name string, project string, commit string) error {
	parts := strings.Split(name, ":")

	if len(parts) < 2 {
		return fmt.Errorf("invalid image name %s", name)
	}

	if parts[len(parts)-1] != commit {
		return fmt.Errorf("image label name is not equal to commit sha : %s != %s", parts[len(parts)-1], commit)
	}

	unescaped, err := url.PathUnescape(project)
	if err != nil {
		return err
	}

	if !strings.Contains(name, unescaped) {
		return fmt.Errorf("image %s do not match project %s", name, unescaped)
	}

	return nil
}
