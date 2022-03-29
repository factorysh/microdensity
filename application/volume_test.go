package application

import (
	"net/http/httptest"
	"testing"

	"github.com/factorysh/microdensity/storage"
	"github.com/factorysh/microdensity/volumes"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestVolume(t *testing.T) {
	v, err := volumes.New("fixtures")
	assert.NoError(t, err)
	s, err := storage.NewFSStore("fixtures")
	assert.NoError(t, err)
	logger, err := zap.NewProduction()
	assert.NoError(t, err)
	a := &Application{
		logger:  logger,
		volumes: v,
		storage: s,
	}
	r := chi.NewRouter()
	r.Get("/service/{serviceID}/{project}/{branch}/{commit}/volumes/*", a.VolumesHandler(6, false))
	ts := httptest.NewServer(r)
	cli := ts.Client()
	defer ts.Close()

	resp, err := cli.Get(ts.URL + "/service/demo/test/main/c31d8da57145e2a98d88679bb7b212f6d71fdd5e/volumes/debug.txt")
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}
