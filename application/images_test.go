package application

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/factorysh/microdensity/mockup"
	"github.com/factorysh/microdensity/service"
	"github.com/stretchr/testify/assert"
)

func TestPostImageHandler(t *testing.T) {
	gitlab := httptest.NewServer(mockup.GitlabJWK(&key.PublicKey))
	defer gitlab.Close()

	cfg, cb, err := SpawnConfig(gitlab.URL)
	defer cb()
	assert.NoError(t, err)

	dataPath, err := ioutil.TempDir(os.TempDir(), "-tasks-test")
	assert.NoError(t, err)
	defer os.RemoveAll(dataPath)
	cfg.DataPath = dataPath

	svc, err := service.NewFolder("../demo/services/demo")
	assert.NoError(t, err)

	app, err := New(cfg)
	assert.NoError(t, err)
	app.Services = map[string]service.Service{
		"demo": svc,
	}

	srvApp := httptest.NewServer(app.Router)
	defer srvApp.Close()

	mockupCommit := "50ccd600c79e35c2d488e4d36814d05f5d57baee"
	mockupGroup := url.PathEscape("group/project")
	cli := http.Client{}

	req, err := mkRequest(key)
	assert.NoError(t, err)
	req.Method = http.MethodPost
	req.URL, err = url.Parse(srvApp.URL)
	assert.NoError(t, err)
	req.URL.Path = fmt.Sprintf("/service/demo/%s/master/%s/_image", mockupGroup, mockupCommit)
	assert.NoError(t, err)
	b := new(bytes.Buffer)
	err = json.NewEncoder(b).Encode(ImageParams{
		Name:      "busybox:latest",
		AuthToken: "",
	})
	// trick to make the buffer a ReaderCloser
	req.Body = ioutil.NopCloser(b)
	r, err := cli.Do(req)
	assert.NoError(t, err)
	defer r.Body.Close()
	assert.Equal(t, 200, r.StatusCode)

}
