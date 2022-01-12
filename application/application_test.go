package application

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/cristalhq/jwt/v3"
	"github.com/factorysh/microdensity/claims"
	"github.com/factorysh/microdensity/service"
	"github.com/stretchr/testify/assert"
)

func TestApplication(t *testing.T) {
	secret := "s3cr37"
	a, err := New(nil, secret)
	assert.NoError(t, err)
	a.Services = append(a.Services, &service.Service{
		Name: "demo",
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
