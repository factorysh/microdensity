package jwt

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/factorysh/microdensity/mockup"
	"github.com/stretchr/testify/assert"
)

func TestMiddleware(t *testing.T) {
	gitlab := httptest.NewServer(mockup.GitlabJWK(&privateRSA1024.PublicKey))
	defer gitlab.Close()

	auth, err := NewJWTAuthenticator(gitlab.URL)
	assert.NoError(t, err)
	srv := httptest.NewServer(auth.Middleware(true)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))
	defer srv.Close()

	res, err := http.Get(srv.URL)
	assert.NoError(t, err)
	assert.Equal(t, 401, res.StatusCode)

}
