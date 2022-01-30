package middlewares

import (
	"crypto/rsa"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cristalhq/jwt/v3"
	"github.com/factorysh/microdensity/claims"
	"github.com/factorysh/microdensity/mockup"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestAuth(t *testing.T) {
	gitlab := httptest.NewServer(mockup.GitlabJWK(&privateRSA1024.PublicKey))
	defer gitlab.Close()

	router := chi.NewRouter()
	authenticator, err := NewJWTAuthenticator(gitlab.URL)
	assert.NoError(t, err)
	router.Use(authenticator.Middleware())
	//router.Use(Auth(key))
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	})
	ts := httptest.NewServer(router)
	defer ts.Close()

	res, err := http.Get(ts.URL)
	assert.NoError(t, err)
	assert.Equal(t, 401, res.StatusCode)

	client := &http.Client{}

	type fixture struct {
		claim  claims.Claims
		key    *rsa.PrivateKey
		status int
	}
	for _, a := range []fixture{
		{ // it's ok
			claim: claims.Claims{
				Owner: "bob",
			},
			key:    privateRSA1024,
			status: 200,
		},
		{ // owner is missing
			claim:  claims.Claims{},
			key:    privateRSA1024,
			status: 400,
		},
		{ // wrong key
			claim: claims.Claims{
				Owner: "bob",
			},
			key:    privateRSA1024_2,
			status: 401,
		},
	} {
		r, err := http.NewRequest("GET", ts.URL, nil)
		assert.NoError(t, err)
		signer, err := jwt.NewSignerRS(jwt.RS256, a.key)
		assert.NoError(t, err)
		builder := jwt.NewBuilder(signer)
		token, err := builder.Build(a.claim)
		assert.NoError(t, err)
		assert.NoError(t, err)
		r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		res, err = client.Do(r)
		assert.NoError(t, err)
		assert.Equal(t, a.status, res.StatusCode)
	}
}
