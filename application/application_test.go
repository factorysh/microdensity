package application

import (
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

	"github.com/cristalhq/jwt/v3"
	"github.com/factorysh/microdensity/claims"
	"github.com/factorysh/microdensity/mockup"
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

	block, _ := pem.Decode([]byte(applicationPrivateRSA))
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)

	gitlab := httptest.NewServer(mockup.GitlabJWK(&key.PublicKey))
	defer gitlab.Close()

	dir, err := ioutil.TempDir(os.TempDir(), "queue-")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)
	a, err := New(nil, gitlab.URL, dir)
	assert.NoError(t, err)
	a.Services = append(a.Services, &NaiveService{
		name: "demo",
	})

	ts := httptest.NewServer(a.Router)
	defer ts.Close()

	cli := http.Client{}

	req, err := mkRequest("fixme")
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

	req, err = mkRequest("fixme")
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
