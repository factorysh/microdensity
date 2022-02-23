package application

import (
	"bytes"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/cristalhq/jwt/v3"
	"github.com/factorysh/microdensity/claims"
	"github.com/factorysh/microdensity/conf"
	"github.com/factorysh/microdensity/mockup"
	"github.com/factorysh/microdensity/service"
	"github.com/factorysh/microdensity/storage"
	"github.com/factorysh/microdensity/task"
	"github.com/stretchr/testify/assert"
)

const applicationPrivateRSA = `-----BEGIN RSA PRIVATE KEY-----
MIICXQIBAAKBgQDPAPJtd+Jd7zaM/PiRXQ8HWRvh5NxF28n6DNvX3Oc8K9Z2nIQJ
JVvKoSx+eCMvWYH9jfCWfJc6dn74SPL4EhNDfzyNNKlFXIGwm0b+BJw7pFvc5oag
3c9H0OqCxRJlsLHUdIHG/APhTGmfN59d7OlYvYOuBGjQDpjPB6ayInTsYwIDAQAB
AoGAMHmFS9M+JEcnXB7FSq0jHtJkMCL63jUY+EBYnxUw5StS3pXKaaXg9/OESt1x
R95LDYhWpbbpZxxmoVfb5fG9pq1hkPQF7s41MjVCs/z0zsA0p3ITFBJTnpwRLlZv
2oml9sAVCLyyysRn6k58Lw+Dl06vE4989LroXQ3yRkcqKDECQQD+jR6BI0DjcInu
F6vKjyv5QaVZHKrpGiD4of+ueP7bIgNvGdB7kQP/brlc42QlmeE/Ddg8vrzH6jBQ
0n7cJ67rAkEA0C6NFXVSp0XScpY0Qc+dNGDmOK73e9gEAQEpXmPfnIWVowJ/J7tA
VMq+yAbljL//kjy/1dWbxvjzP0n2f8sKaQJAHh0Bs9NA1Oc2WgVQ3GitkhIzBmS+
z065AdDgV3qW48OVVmpeYI/aQjiOEzAPY+ddX0E7CIyj9p580sLkIRVMuwJBALON
MsmrItp6cgO6YN/R/LhMSsPgxDrgGLP1GIT8hsQswt6RLLJL4jQ/mURUDm/SuM6b
7qizT2PRG5seY6fcquECQQCApqO/lNdrqIpDWHSA4aCVZdYCW5u8SNe1xekqmXsy
bhfhLtK7l19RUDS9g702dcr+z7UxZS97SztCWyEO/mjs
-----END RSA PRIVATE KEY-----`

var key = MustParseRSAKey(applicationPrivateRSA)

func TestApplication(t *testing.T) {
	gitlab := httptest.NewServer(mockup.GitlabJWK(&key.PublicKey))
	defer gitlab.Close()

	cfg, cb, err := SpawnConfig(gitlab.URL)
	defer cb()
	assert.NoError(t, err)

	dataPath, err := ioutil.TempDir(os.TempDir(), "-tasks-test")
	assert.NoError(t, err)
	// defer os.RemoveAll(dataPath)
	cfg.DataPath = dataPath

	app, err := New(cfg)
	assert.NoError(t, err)
	svc, err := service.NewFolder("../demo/services/demo")
	app.Services = map[string]service.Service{"demo": svc}

	srvApp := httptest.NewServer(app.Router)
	defer srvApp.Close()

	cli := http.Client{}
	req, err := mkRequest(key)
	assert.NoError(t, err)
	req.Method = http.MethodGet
	req.URL, err = url.Parse(fmt.Sprintf("%s/services", srvApp.URL))
	assert.NoError(t, err)

	r, err := cli.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)
	dec := json.NewDecoder(r.Body)
	var servicesList []string
	err = dec.Decode(&servicesList)
	assert.NoError(t, err)
	assert.Equal(t, []string{"demo"}, servicesList)

	req, err = mkRequest(key)
	assert.NoError(t, err)
	req.Method = http.MethodGet
	req.URL, err = url.Parse(fmt.Sprintf("%s/service/demo", srvApp.URL))
	assert.NoError(t, err)
	r, err = cli.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)
	assert.Equal(t, "text/html; charset=utf-8", r.Header.Get("content-type"))

	req, err = mkRequest(key)
	assert.NoError(t, err)
	req.Method = http.MethodGet
	req.URL, err = url.Parse(fmt.Sprintf("%s/service/wombat", srvApp.URL))
	assert.NoError(t, err)
	r, err = cli.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 404, r.StatusCode)

	// dummy task
	mockupCommit := "50ccd600c79e35c2d488e4d36814d05f5d57baee"
	mockupGroup := url.PathEscape("group/project")
	req, err = mkRequest(key)
	assert.NoError(t, err)
	req.Method = http.MethodPost
	req.URL, err = url.Parse(srvApp.URL)
	assert.NoError(t, err)
	req.URL.Path = fmt.Sprintf("/service/demo/%s/master/%s", mockupGroup, mockupCommit)
	assert.NoError(t, err)
	b := new(bytes.Buffer)
	err = json.NewEncoder(b).Encode(map[string]interface{}{"HELLO": "Bob"})
	// trick to make the buffer a ReaderCloser
	req.Body = ioutil.NopCloser(b)
	r, err = cli.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)

	// get the status badge
	req, err = mkRequest(key)
	assert.NoError(t, err)
	req.Method = http.MethodGet
	req.URL, err = url.Parse(srvApp.URL)
	assert.NoError(t, err)
	req.URL.Path = fmt.Sprintf("/service/demo/%s/master/%s/status", mockupGroup, mockupCommit)
	assert.NoError(t, err)
	r, err = cli.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)
	assert.Equal(t, "image/svg+xml", r.Header["Content-Type"][0])

	// get the status badge
	req, err = mkRequest(key)
	assert.NoError(t, err)
	req.Method = http.MethodGet
	req.URL, err = url.Parse(srvApp.URL)
	assert.NoError(t, err)
	req.URL.Path = fmt.Sprintf("/service/demo/%s/master/latest/status", mockupGroup)
	assert.NoError(t, err)
	r, err = cli.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)
	assert.Equal(t, "image/svg+xml", r.Header["Content-Type"][0])

	// FIXME: PubSub on a task avoid pooling/waiting
	time.Sleep(3 * time.Second)

	// get the volume
	req, err = mkRequest(key)
	assert.NoError(t, err)
	req.Method = http.MethodGet
	req.URL, err = url.Parse(srvApp.URL)
	assert.NoError(t, err)
	req.URL.Path = fmt.Sprintf("/service/demo/%s/master/%s/volumes/cache/proof", mockupGroup, mockupCommit)
	assert.NoError(t, err)
	r, err = cli.Do(req)
	assert.NoError(t, err)
	defer r.Body.Close()
	assert.Equal(t, 200, r.StatusCode)
	data, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)
	assert.Equal(t, "proof\n", string(data))

	req, err = mkRequest(key)
	assert.NoError(t, err)
	req.Method = http.MethodGet
	req.URL, err = url.Parse(fmt.Sprintf("%s/service/demo/group/project/-/master/%s/volumes/cache/proof", srvApp.URL, mockupCommit))
	assert.NoError(t, err)
	r, err = cli.Do(req)
	assert.NoError(t, err)
	defer r.Body.Close()
	assert.Equal(t, 200, r.StatusCode)
	data, err = ioutil.ReadAll(r.Body)
	assert.NoError(t, err)
	assert.Equal(t, "proof\n", string(data))

	req, err = mkRequest(key)
	assert.NoError(t, err)
	req.Method = http.MethodGet
	req.URL, err = url.Parse(fmt.Sprintf("%s/service/demo/group/project/-/master/%s/volumes/data/result.html", srvApp.URL, mockupCommit))
	assert.NoError(t, err)
	r, err = cli.Do(req)
	assert.NoError(t, err)
	defer r.Body.Close()
	assert.Equal(t, 200, r.StatusCode)
	data, err = ioutil.ReadAll(r.Body)
	assert.NoError(t, err)
	assert.Contains(t, string(data), "Bob")

}

func TestApplicationStart(t *testing.T) {
	// state 5 is interrupted
	rawJSON := `{"id":"f79b5c4c-94b4-11ec-a442-00163e007d68","service":"waiter","project":"group%2Fproject","branch":"master","commit":"7e15b158cfc3e8f6bbe3e441a0cdb64bba135ef3","creation":"2022-02-23T15:29:11.082288364+01:00","Args":{"WAIT":10},"State":5}`

	gitlab := httptest.NewServer(mockup.GitlabJWK(&key.PublicKey))
	defer gitlab.Close()

	cfg, cb, err := SpawnConfig(gitlab.URL)
	defer cb()
	assert.NoError(t, err)

	a, err := New(cfg)
	assert.NoError(t, err)

	taskPath := path.Join(cfg.DataPath, "data", "wait", "group%2Fproject", "master", "f79b5c4c-94b4-11ec-a442-00163e007d68")
	err = os.MkdirAll(taskPath, storage.DirMode)
	assert.NoError(t, err)

	err = os.WriteFile(path.Join(taskPath, "task.json"), []byte(rawJSON), 0644)
	assert.NoError(t, err)

	err = a.Run(":9090")
	assert.NoError(t, err)

	tasks, err := a.storage.All()
	assert.NoError(t, err)
	assert.Equal(t, task.Running, tasks[0].State)

	err = a.Shutdown()
	assert.NoError(t, err)
}

	assert.NoError(t, err)
	r, err := cli.Do(req)
	assert.NoError(t, err)
	defer r.Body.Close()
	assert.Equal(t, 200, r.StatusCode)
	assert.Equal(t, "image/svg+xml", r.Header["Content-Type"][0])
	data, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)
	assert.Contains(t, string(data), "<text x=\"122\" y=\"14\">Running</text>\n")

	err = a.Shutdown()
	assert.NoError(t, err)
}

func SpawnConfig(gitlabURL string) (*conf.Conf, func(), error) {
	dataDir, err := ioutil.TempDir(os.TempDir(), "data")
	if err != nil {
		return nil, nil, err
	}

	serviceDir, err := filepath.Abs("../demo/services")
	if err != nil {
		return nil, nil, err
	}

	cfg := &conf.Conf{
		JWKProvider: gitlabURL,
		DataPath:    dataDir,
		Services:    serviceDir,
	}

	return cfg, func() {
		os.RemoveAll(dataDir)
	}, nil

}

func freshToken(key *rsa.PrivateKey) (*jwt.Token, error) {
	signer, err := jwt.NewSignerRS(jwt.RS256, key)
	if err != nil {
		return nil, err
	}
	return jwt.NewBuilder(signer,
		jwt.WithKeyID(mockup.Kid(&key.PublicKey))).Build(
		claims.Claims{
			StandardClaims: jwt.StandardClaims{
				IssuedAt: jwt.NewNumericDate(time.Now()),
			},
			UserLogin:   "Bob",
			ProjectPath: "group/project",
		})

}

func mkRequest(key *rsa.PrivateKey) (*http.Request, error) {
	t, err := freshToken(key)
	if err != nil {
		return nil, err
	}
	return &http.Request{
		Header: http.Header{
			"PRIVATE-TOKEN": {t.String()},
		},
	}, nil
}
