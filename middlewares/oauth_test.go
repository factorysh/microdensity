package middlewares

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/factorysh/microdensity/conf"
	"github.com/factorysh/microdensity/gitlab"
	"github.com/factorysh/microdensity/sessions"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestOAuthRedirect(t *testing.T) {
	mockUP := gitlab.TestMockup()
	defer mockUP.Close()

	s := sessions.New()
	router := chi.NewRouter()
	router.Group(func(r chi.Router) {
		r.Use(OAuth(&conf.OAuthConf{
			ProviderURL: mockUP.URL,
			AppID:       "id",
			AppSecret:   "secret",
			AppURL:      "url",
		}, &s))
		r.Get("/{project}", func(w http.ResponseWriter, _ *http.Request) {
			w.Write([]byte("i am protected"))
			w.WriteHeader(http.StatusOK)
		})

	})

	ts := httptest.NewServer(router)
	defer ts.Close()

	res, err := http.Get(fmt.Sprintf("%s%s", ts.URL, "/project"))
	assert.NoError(t, err)
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)
	assert.Equal(t, res.StatusCode, http.StatusUnauthorized, "expected a redirect status code")
	assert.Contains(t, string(body), "Please login with <a href=")
}
