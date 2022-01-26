package application

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/cristalhq/jwt/v3"
	"github.com/factorysh/microdensity/claims"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type NaiveService struct {
	name string
}

func (n *NaiveService) Name() string {
	return n.name
}

func (n *NaiveService) Validate(map[string]interface{}) error {
	return nil
}

func (n *NaiveService) New(project string, args map[string]interface{}) (uuid.UUID, error) {
	return uuid.Nil, nil
}

func (n *NaiveService) Run(id uuid.UUID) error {
	return nil
}

func TestApplication(t *testing.T) {
	secret := "s3cr37"
	dir, err := ioutil.TempDir(os.TempDir(), "queue-")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)
	a, err := New(nil, secret, dir)
	assert.NoError(t, err)
	a.Services = append(a.Services, &NaiveService{
		name: "demo",
	})

	ts := httptest.NewServer(a.router)
	defer ts.Close()

	cli := http.Client{}

	req, err := mkRequest(secret)
	assert.NoError(t, err)
	req.Method = http.MethodGet
	req.URL, err = url.Parse(fmt.Sprintf("%s/services", ts.URL))
	assert.NoError(t, err)

	r, err := cli.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)
	dec := json.NewDecoder(r.Body)
	var services []string
	err = dec.Decode(&services)
	assert.NoError(t, err)
	assert.Equal(t, []string{"demo"}, services)

	req, err = mkRequest(secret)
	assert.NoError(t, err)
	req.Method = http.MethodGet
	req.URL, err = url.Parse(fmt.Sprintf("%s/service/demo", ts.URL))
	assert.NoError(t, err)

	r, err = cli.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)
}

func freshToken(key string) (*jwt.Token, error) {
	signer, err := jwt.NewSignerHS(jwt.HS256, []byte(key))
	if err != nil {
		return nil, err
	}
	clm := claims.Claims{
		Owner: "Bob",
		Admin: false,
		Path:  "group/project",
	}

	builder := jwt.NewBuilder(signer)
	return builder.Build(clm)
}

func mkRequest(key string) (*http.Request, error) {
	t, err := freshToken(key)
	if err != nil {
		return nil, err
	}
	return &http.Request{
		Header: http.Header{
			"Authorization": {fmt.Sprintf("Bearer %s", t.String())},
		},
	}, nil
}
