package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/factorysh/microdensity/httpcontext"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestTokens(t *testing.T) {
	tests := []struct {
		name       string
		headers    map[string]string
		cookies    map[string]string
		wantJWT    string
		wantAccess string
	}{
		{name: "With JWT", headers: map[string]string{"Authorization": "Bearer token"}, wantJWT: "token"},
		{name: "With Access", cookies: map[string]string{"OAUTH2_SESSION": "access"}, wantAccess: "access"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// create a router
			router := chi.NewRouter()
			// add tokens middleware
			router.Use(Tokens())
			// add handler that tests inside an http handler context
			router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				value, err := httpcontext.GetJWT(r)
				// jwt
				if tc.wantJWT != "" {
					assert.Equal(t, tc.wantJWT, value)
					assert.NoError(t, err)
				} else {
					assert.Error(t, err)
				}

				value, err = httpcontext.GetAccessToken(r)
				if tc.wantAccess != "" {
					assert.Equal(t, tc.wantAccess, value)
					assert.NoError(t, err)
				} else {
					assert.Error(t, err)
				}

				// access token
				w.WriteHeader(http.StatusOK)
			})

			ts := httptest.NewServer(router)
			defer ts.Close()

			// do request
			cli := http.Client{}
			req, err := http.NewRequest(http.MethodGet, ts.URL, nil)
			assert.NoError(t, err)
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}
			for k, v := range tc.cookies {
				cookie := http.Cookie{Name: k, Value: v, Expires: time.Now().Add(10 * time.Minute)}
				req.AddCookie(&cookie)
			}

			_, err = cli.Do(req)
			assert.NoError(t, err)
		})
	}

}
