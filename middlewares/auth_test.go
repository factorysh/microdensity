package middlewares

import (
	"crypto/rsa"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cristalhq/jwt/v3"
	"github.com/factorysh/microdensity/claims"
	_jwt "github.com/factorysh/microdensity/middlewares/jwt"
	"github.com/factorysh/microdensity/mockup"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestRESTAuthJWT(t *testing.T) {
	gitlab := httptest.NewServer(mockup.GitlabJWK(&privateRSA1024.PublicKey))
	defer gitlab.Close()

	router := chi.NewRouter()
	authenticator, err := _jwt.NewJWTAuthenticator(gitlab.URL)
	assert.NoError(t, err)
	router.Use(authenticator.Handler(true))
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
				UserLogin: "bob",
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
				UserLogin: "bob",
			},
			key:    privateRSA1024_2,
			status: 400,
		},
	} {
		a.claim.IssuedAt = jwt.NewNumericDate(time.Now())
		a.claim.ExpiresAt = jwt.NewNumericDate(time.Now().Add(10 * time.Minute))
		r, err := http.NewRequest("GET", ts.URL, nil)
		assert.NoError(t, err)
		signer, err := jwt.NewSignerRS(jwt.RS256, a.key)
		assert.NoError(t, err)
		token, err := jwt.NewBuilder(signer, jwt.WithKeyID(mockup.Kid(&privateRSA1024.PublicKey))).Build(a.claim)
		assert.NoError(t, err)
		r.Header.Set("Private-Token", token.String())
		res, err = client.Do(r)
		assert.NoError(t, err)
		assert.Equal(t, a.status, res.StatusCode)
	}
}
