package application

import (
	"io/ioutil"
	"net/http/httptest"
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

}
