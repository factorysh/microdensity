package application

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	_jwt "github.com/factorysh/microdensity/middlewares/jwt"
	"github.com/factorysh/microdensity/mockup"
	"github.com/factorysh/microdensity/queue"
	"github.com/stretchr/testify/assert"
	"go.etcd.io/bbolt"
)

type rc struct {
	*bytes.Buffer
}

func (r *rc) Close() error {
	return nil
}

func TestCreateTask(t *testing.T) {

	block, _ := pem.Decode([]byte(applicationPrivateRSA))
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)

	gitlab := httptest.NewServer(mockup.GitlabJWK(&key.PublicKey))
	defer gitlab.Close()

	jwtAuth, err := _jwt.NewJWTAuthenticator(gitlab.URL)
	assert.NoError(t, err)

	dir, err := ioutil.TempDir(os.TempDir(), "queue-")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)
	s, err := bbolt.Open(
		fmt.Sprintf("%s/bbolt.store", dir),
		0600, &bbolt.Options{})
	assert.NoError(t, err)
	q, err := queue.New(s)
	assert.NoError(t, err)
	a, err := New(q, nil, jwtAuth, dir, gitlab.URL)
	assert.NoError(t, err)
	a.Services = append(a.Services, &NaiveService{
		name: "demo",
	})

	ts := httptest.NewServer(a.Router)
	defer ts.Close()

	cli := http.Client{}
	req, err := mkRequest(key)
	assert.NoError(t, err)
	req.Method = "POST"
	req.URL, err = url.Parse(fmt.Sprintf("%s/service/demo/group%%2Fproject/main/8e54b1d8c5f0859370196733feeb00da022adeb5", ts.URL))
	assert.NoError(t, err)
	b := &rc{
		&bytes.Buffer{},
	}
	err = json.NewEncoder(b).Encode(map[string]interface{}{
		"name": "Bob",
	})
	assert.NoError(t, err)

	req.Body = b
	r, err := cli.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)

	l, err := q.Length()
	assert.NoError(t, err)
	assert.Equal(t, 1, l)

	req, err = mkRequest(key)
	assert.NoError(t, err)
	req.Method = "GET"
	req.URL, err = url.Parse(fmt.Sprintf("%s/service/demo/group%%2Fproject/main/8e54b1d8c5f0859370196733feeb00da022adeb5", ts.URL))
	assert.NoError(t, err)
	r, err = cli.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)
}
