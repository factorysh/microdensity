package badge

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/factorysh/microdensity/middlewares"
	"github.com/factorysh/microdensity/queue"
	"github.com/factorysh/microdensity/task"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.etcd.io/bbolt"
)

func TestBadge(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "queue-")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)
	s, err := bbolt.Open(
		fmt.Sprintf("%s/bbolt.store", dir),
		0600, &bbolt.Options{})
	assert.NoError(t, err)
	q, err := queue.New(s)
	assert.NoError(t, err)
	q.Put(&task.Task{
		Id:    uuid.MustParse("63E322B7-A9D0-4BDA-85AD-5867F90A1DBA"),
		State: task.Running,
	})
	r := chi.NewRouter()
	r.Route("/{service}/{project}/{id}/badge, ", func(r chi.Router) {
		r.Use(middlewares.Project())
		r.Get("/", BadgeMyProject(q, "status"))
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	resp, err := http.Get(fmt.Sprintf("%s/demo/42/63E322B7-A9D0-4BDA-85AD-5867F90A1DBA/badge", ts.URL))
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}
