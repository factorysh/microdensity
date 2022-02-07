package badge

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/factorysh/microdensity/storage"
	"github.com/factorysh/microdensity/task"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestBadge(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "data-")
	assert.NoError(t, err)
	// defer os.RemoveAll(dir)
	store, err := storage.NewFSStore(dir)
	assert.NoError(t, err)

	err = store.Upsert(
		&task.Task{
			Id:      uuid.MustParse("63E322B7-A9D0-4BDA-85AD-5867F90A1DBA"),
			State:   task.Running,
			Project: "42",
		})
	r := chi.NewRouter()
	r.Route("/s/{service:[a-z-]+}/{project}/{id}/badge", func(r chi.Router) {
		//r.Use(_project.AssertProject)
		r.Get("/", StatusBadge(store, false))
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	u := fmt.Sprintf("%s/s/demo/42/63E322B7-A9D0-4BDA-85AD-5867F90A1DBA/badge", ts.URL)
	fmt.Println("url", u)
	resp, err := http.Get(u)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	defer resp.Body.Close()
	f, err := os.Create(path.Join(dir, "toto.svg"))
	assert.NoError(t, err)
	defer f.Close()
	io.Copy(f, resp.Body)
	// If you want to see the svg, comment the `defer os.RemoveAll(dir)`
	fmt.Println(dir)
}
