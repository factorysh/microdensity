package middlewares

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cristalhq/jwt/v3"
	"github.com/factorysh/microdensity/claims"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestAuth(t *testing.T) {
	key := "plop"
	router := chi.NewRouter()
	router.Use(Tokens())
	router.Use(Auth(key))
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
		key    []byte
		status int
	}
	for _, a := range []fixture{
		{ // it's ok
			claim: claims.Claims{
				Owner: "bob",
			},
			key:    []byte(key),
			status: 200,
		},
		{ // owner is missing
			claim:  claims.Claims{},
			key:    []byte(key),
			status: 400,
		},
		{ // wrong key
			claim: claims.Claims{
				Owner: "bob",
			},
			key:    []byte("wrong key"),
			status: 401,
		},
	} {
		r, err := http.NewRequest("GET", ts.URL, nil)
		assert.NoError(t, err)
		signer, err := jwt.NewSignerHS(jwt.HS256, []byte(a.key))
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
