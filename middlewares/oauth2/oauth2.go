package oauth2

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"time"

	"github.com/factorysh/microdensity/conf"
	"github.com/factorysh/microdensity/httpcontext"
	"github.com/factorysh/microdensity/oauth"
	"github.com/factorysh/microdensity/server"
	_sessions "github.com/factorysh/microdensity/sessions"
	"github.com/go-chi/chi/v5"
)

var (
	// used to ensure embed import
	_ embed.FS
	//go:embed templates/login.html
	loginTemplate string
)

// OAuth2 will trigger an OAuth2 flow if no auth token is found see https://docs.gitlab.com/ee/api/oauth2.html#authorization-code-flow
func OAuth2(oauthConfig *conf.OAuthConf, sessions *_sessions.Sessions) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			project := chi.URLParam(r, "project")
			// if context contains a JWT token
			if _, err := httpcontext.GetJWT(r); err == nil {
				next.ServeHTTP(w, r)
				return
			}

			// if context contains a session access token
			if accessToken, err := httpcontext.GetAccessToken(r); err == nil {
				// verify token access
				if sessions.Authorize(accessToken, project) {
					// if authorized, add value to context
					ctx := context.WithValue(r.Context(), httpcontext.IsOAuth, true)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}

			// fallback to oauth flow
			template, err := template.New("login").Parse(loginTemplate)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			state, err := _sessions.GenID()
			if err != nil {
				fmt.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			// Golang stdlib wants you to set cookies first (if not they will vanish)
			addOAuthFlowCookies(w, r, *oauthConfig, state, project)

			// No token means forbidden + ask to login page
			w.WriteHeader(http.StatusUnauthorized)

			// /oauth/authorize?client_id=APP_ID&redirect_uri=REDIRECT_URI&response_type=code&state=STATE&scope=REQUESTED_SCOPES
			values := url.Values{}
			values.Add("client_id", oauthConfig.AppID)
			values.Add("redirect_uri", fmt.Sprintf("%s%s", oauthConfig.AppURL, server.OAuthCallbackEndpoint))
			values.Add("response_type", "code")
			values.Add("state", state)
			values.Add("scope", "read_api")

			// prepare data for template
			data := struct {
				AuthURL string
			}{
				AuthURL: fmt.Sprintf("%s/oauth/authorize?%s", oauthConfig.ProviderURL, values.Encode()),
			}

			// write filled template to response body
			err = template.Execute(w, data)
			if err != nil {
				fmt.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		})
	}
}

// addOAuthFlowCookies adds a set cookie directive use in the OAuth2 flow
func addOAuthFlowCookies(w http.ResponseWriter, r *http.Request, oauthConfig conf.OAuthConf, state string, project string) {
	// state is used as a random value to prevent CSRF attacks
	stateCookie := http.Cookie{
		Name:    oauth.StateCookieName,
		Domain:  oauthConfig.AppURL,
		Path:    server.OAuthCallbackEndpoint,
		Expires: time.Now().Add(30 * time.Second),
	}
	stateCookie.Value = state
	http.SetCookie(w, &stateCookie)

	// remember where the user comes from
	originURI := http.Cookie{
		Name:    oauth.OriginURICookieName,
		Domain:  oauthConfig.AppURL,
		Path:    server.OAuthCallbackEndpoint,
		Expires: time.Now().Add(30 * time.Second),
	}
	originURI.Value = fmt.Sprintf("%s%s", oauthConfig.AppURL, r.RequestURI)
	http.SetCookie(w, &originURI)

	// remember the requested project
	originProject := http.Cookie{
		Name:    oauth.OriginProjectCookieName,
		Domain:  oauthConfig.AppURL,
		Path:    server.OAuthCallbackEndpoint,
		Expires: time.Now().Add(30 * time.Second),
	}
	originProject.Value = project
	http.SetCookie(w, &originProject)

}
