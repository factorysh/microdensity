package application

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/cristalhq/jwt/v3"
	"github.com/factorysh/microdensity/claims"
	"github.com/factorysh/microdensity/conf"
	"github.com/factorysh/microdensity/mockup"
	"github.com/factorysh/microdensity/queue"
	"github.com/factorysh/microdensity/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type NaiveService struct {
	name string
}

func (n *NaiveService) Name() string {
	return n.name
}

func (n *NaiveService) Validate(map[string]interface{}) (service.Arguments, error) {
	return service.Arguments{}, nil
}

func (n *NaiveService) New(project string, args map[string]interface{}) (uuid.UUID, error) {
	return uuid.Nil, nil
}

func (n *NaiveService) Run(id uuid.UUID) error {
	return nil
}

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

func TestApplication(t *testing.T) {

	var services = map[string]service.Service{
		"demo": &NaiveService{
			name: "demo",
		},
	}
	key, app, _, _, cleanUp := prepareTestingContext(t, services)
	defer cleanUp()

	cli := http.Client{}
	req, err := mkRequest(key)
	assert.NoError(t, err)
	req.Method = http.MethodGet
	req.URL, err = url.Parse(fmt.Sprintf("%s/services", app.URL))
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
	req.URL, err = url.Parse(fmt.Sprintf("%s/service/demo", app.URL))
	assert.NoError(t, err)

	r, err = cli.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)
}

func prepareTestingContext(t *testing.T, services map[string]service.Service) (key *rsa.PrivateKey,
	gitlab := httptest.NewServer(mockup.GitlabJWK(&key.PublicKey))
	defer gitlab.Close()
	app *httptest.Server,

	q *queue.Storage,
	cleanUp func()) {

	volDir, err := ioutil.TempDir(os.TempDir(), "volumes")
	assert.NoError(t, err)
	queueDir, err := ioutil.TempDir(os.TempDir(), "queue-")
	assert.NoError(t, err)

	cfg := &conf.Conf{
		JWKProvider: gitlab.URL,
		VolumePath:  volDir,
		Queue:       queueDir,
	}

	a, err := New(cfg)
	assert.NoError(t, err)
	a.Services = services

	app = httptest.NewServer(a.Router)

	return key, app, gitlab, q, func() {
		os.RemoveAll(volDir)
		os.RemoveAll(queueDir)
		gitlab.Close()
		app.Close()
	}
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
