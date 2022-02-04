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

type rc struct {
	*bytes.Buffer
}

func (r *rc) Close() error {
	return nil
}

func TestCreateTask(t *testing.T) {
	gitlab := httptest.NewServer(mockup.GitlabJWK(&key.PublicKey))
	defer gitlab.Close()

	cfg, cb, err := SpawnConfig(gitlab.URL)
	defer cb()
	assert.NoError(t, err)

	tests := []struct {
		name         string
		args         map[string]interface{}
		createStatus int
		getStatus    int
		sLen         int
	}{
		{
			name:         "Valid args",
			sLen:         1,
			args:         map[string]interface{}{"HELLO": "Bob"},
			createStatus: http.StatusOK,
			getStatus:    http.StatusOK},
		{
			name:         "Invalid args",
			sLen:         0,
			args:         map[string]interface{}{"nop": "Bob"},
			createStatus: http.StatusBadRequest,
			getStatus:    http.StatusNotFound},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dataPath, err := ioutil.TempDir(os.TempDir(), "-tasks-test")
			assert.NoError(t, err)
			defer os.RemoveAll(dataPath)
			cfg.DataPath = dataPath

			svc, err := service.NewFolder("../demo")
			assert.NoError(t, err)

			app, err := New(cfg)
			assert.NoError(t, err)
			app.Services = map[string]service.Service{
				"demo": svc,
			}

			srvApp := httptest.NewServer(app.Router)
			defer srvApp.Close()

			taskPath := "service/demo/group%2Fproject/main/8e54b1d8c5f0859370196733feeb00da022adeb5"
			cli := http.Client{}
			req, err := mkRequest(key)
			assert.NoError(t, err)
			req.Method = "POST"
			req.URL, err = url.Parse(fmt.Sprintf("%s/%s", srvApp.URL, taskPath))
			assert.NoError(t, err)
			b := &rc{
				&bytes.Buffer{},
			}
			err = json.NewEncoder(b).Encode(tc.args)
			assert.NoError(t, err)

			req.Body = b
			r, err := cli.Do(req)
			assert.NoError(t, err)
			rbody, err := ioutil.ReadAll(r.Body)
			assert.NoError(t, err)
			fmt.Println("body", string(rbody))
			assert.Equal(t, tc.createStatus, r.StatusCode)

			/*
				FIXME
				l, err := app.storage.All()
				assert.NoError(t, err)
				assert.Len(t, l, tc.sLen, tc)
			*/

			req, err = mkRequest(key)
			assert.NoError(t, err)
			req.Method = "GET"
			req.URL, err = url.Parse(fmt.Sprintf("%s/%s", srvApp.URL, taskPath))
			assert.NoError(t, err)
			r, err = cli.Do(req)
			assert.NoError(t, err)
			assert.Equal(t, tc.getStatus, r.StatusCode)
		})
	}

}
