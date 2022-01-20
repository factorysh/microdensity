package middlewares

import (
	"embed"
	"fmt"
	"net/http"
	"net/url"
	"text/template"
	"time"

	"github.com/factorysh/microdensity/oauth"
	"github.com/go-chi/chi/v5"
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
			project := chi.URLParam(r, "project")
			if project == "" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			// get an auth token
			token, _, err := getToken(r)
			// if a token is found, pass to next middleware
			if token != "" && err == nil {
				next.ServeHTTP(w, r)
				return
			}

			// get the login template
			template, err := template.New("login").Parse(loginTemplate)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			state := uuid.NewString()

			// Golang stdlib wants you to set cookies first (if not they will vanish)
			addOAuthFlowCookies(w, r, *oauthConfig, state, project)

			// No token means forbidden + ask to login page
			w.WriteHeader(http.StatusForbidden)

			// /oauth/authorize?client_id=APP_ID&redirect_uri=REDIRECT_URI&response_type=code&state=STATE&scope=REQUESTED_SCOPES
			values := url.Values{}
			values.Add("client_id", oauthConfig.AppID)
			values.Add("redirect_uri", fmt.Sprintf("https://%s%s", oauthConfig.AppDomain, oauth.CallbackEndpoint))
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
		})
	}
}

// addOAuthFlowCookies adds a set cookie directive use in the OAuth2 flow
func addOAuthFlowCookies(w http.ResponseWriter, r *http.Request, oauthConfig oauth.Config, state string, project string) {
	// state is used as a random value to prevent CSRF attacks
	stateCookie := http.Cookie{
		Name:    oauth.StateCookieName,
		Domain:  oauthConfig.AppDomain,
		Path:    oauth.CallbackEndpoint,
		Expires: time.Now().Add(30 * time.Second),
	}
	stateCookie.Value = state
	http.SetCookie(w, &stateCookie)

	// remember where the user comes from
	originURI := http.Cookie{
		Name:    oauth.OriginURICookieName,
		Domain:  oauthConfig.AppDomain,
		Path:    oauth.CallbackEndpoint,
		Expires: time.Now().Add(30 * time.Second),
	}
	originURI.Value = fmt.Sprintf("https://%s%s", oauthConfig.AppDomain, r.RequestURI)
	http.SetCookie(w, &originURI)

	// remember the requested project
	originProject := http.Cookie{
		Name:    oauth.OriginProjectCookieName,
		Domain:  oauthConfig.AppDomain,
		Path:    oauth.CallbackEndpoint,
		Expires: time.Now().Add(30 * time.Second),
	}
	originProject.Value = project
	http.SetCookie(w, &originProject)

}
