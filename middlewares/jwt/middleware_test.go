package jwt

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cristalhq/jwt/v3"
	"github.com/factorysh/microdensity/mockup"
	"github.com/stretchr/testify/assert"
)

func TestMiddleware(t *testing.T) {
	gitlab := httptest.NewServer(mockup.GitlabJWK(&privateRSA1024.PublicKey))
	defer gitlab.Close()

	auth, err := NewJWTAuthenticator(gitlab.URL)
	assert.NoError(t, err)
	srv := httptest.NewServer(auth.Handler(true)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))
	defer srv.Close()

	client := &http.Client{}

	res, err := http.Get(srv.URL)
	assert.NoError(t, err)
	assert.Equal(t, 401, res.StatusCode)
	signer, err := jwt.NewSignerRS(jwt.RS256, privateRSA1024)
	assert.NoError(t, err)
	token, err := jwt.NewBuilder(signer,
		jwt.WithKeyID(mockup.Kid(&privateRSA1024.PublicKey))).Build(
		jwt.StandardClaims{
			IssuedAt: jwt.NewNumericDate(time.Now()),
		})
	assert.NoError(t, err)
	r, err := http.NewRequest("GET", srv.URL, nil)
	assert.NoError(t, err)
	r.Header.Set("Private-Token", token.String())
	res, err = client.Do(r)
	assert.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)
}
