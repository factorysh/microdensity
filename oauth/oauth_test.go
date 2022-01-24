package oauth

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/factorysh/microdensity/conf"
	"github.com/factorysh/microdensity/gitlab"
	"github.com/factorysh/microdensity/server"
	"github.com/factorysh/microdensity/sessions"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestOAuthCallback(t *testing.T) {
	// gitlab
	mockUP := gitlab.TestMockup()

	// sessions
	s := sessions.New()
	s.Put("session", "access_token", time.Now().Add(10*time.Minute), &gitlab.DummyProject)

	// app server
	router := chi.NewRouter()

	ts := httptest.NewServer(router)
	defer ts.Close()

	oauthConf := conf.OAuthConf{
		ProviderURL: mockUP.URL,
		AppID:       "id",
		AppSecret:   "secret",
		AppURL:      ts.URL,
	}

	router.Get(server.OAuthCallbackEndpoint, CallbackHandler(&oauthConf, &s))
	router.Get("/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	cli := http.Client{}
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s", ts.URL, "oauth/callback?state=state&code=code"), nil)
	pCookie := http.Cookie{Name: OriginProjectCookieName, Value: "group/project", Expires: time.Now().Add(10 * time.Minute)}
	sCookie := http.Cookie{Name: StateCookieName, Value: "state", Expires: time.Now().Add(10 * time.Minute)}
	for _, c := range []http.Cookie{pCookie, sCookie} {
		req.AddCookie(&c)
	}
	assert.NoError(t, err)

	res, err := cli.Do(req)
	assert.NoError(t, err)
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)
}
