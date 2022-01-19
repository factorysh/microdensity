package middlewares

import (
	"embed"
	"fmt"
	"net/http"
	"net/url"
	"text/template"
	"time"

	"github.com/factorysh/microdensity/oauth"
	"github.com/google/uuid"
)

var (
	// used to ensure embed import
	_ embed.FS
	//go:embed templates/login.html
	loginTemplate string
)

// OAuth will trigger an OAuth flow if no auth token is found see https://docs.gitlab.com/ee/api/oauth2.html#authorization-code-flow
func OAuth(oauthConfig *oauth.Config) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// get an auth token
			token, _, err := getToken(r)
			// if a token is found, pass to next middleware
			if token != "" && err == nil {
				next.ServeHTTP(w, r)
				return
			}

			// No token means forbidden + redirect
			w.WriteHeader(http.StatusForbidden)

			// get the login template
			template, err := template.New("login").Parse(loginTemplate)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			state := uuid.NewString()

			// /oauth/authorize?client_id=APP_ID&redirect_uri=REDIRECT_URI&response_type=code&state=STATE&scope=REQUESTED_SCOPES
			values := url.Values{}
			values.Add("client_id", oauthConfig.AppID)
			values.Add("redirect_uri", fmt.Sprintf("https://%s/oauth/callback", oauthConfig.AppDomain))
			values.Add("response_type", "code")
			values.Add("state", state)
			values.Add("scope", "read_user")

			// prepare data for template
			data := struct {
				AuthURL string
			}{
				AuthURL: fmt.Sprintf("https://%s/oauth/authorize?%s", oauthConfig.ProviderDomain, values.Encode()),
			}

			// write filled template to response body
			template.Execute(w, data)

			// set cookies
			// state is used as a random value to prevent CSRF attacks
			stateCookie := http.Cookie{
				Name:    "OAUTH2_STATE",
				Domain:  oauthConfig.AppDomain,
				Path:    "/oauth/callback",
				Expires: time.Now().AddDate(0, 1, 0),
			}
			stateCookie.Value = state
			http.SetCookie(w, &stateCookie)

			// TODO: save requestURI using state
		})
	}
}
