package oauth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/factorysh/microdensity/conf"
	"github.com/factorysh/microdensity/gitlab"
	_sessions "github.com/factorysh/microdensity/sessions"
)

const (
	// StateCookieName unify name of oauth2 cookies
	StateCookieName = "OAUTH2_STATE"
	// OriginURICookieName unify name of oauth2 cookies
	OriginURICookieName = "OAUTH2_ORIGIN_URI"
	// OriginProjectCookieName unify name of oauth2 cookies
	OriginProjectCookieName = "OAUTH2_ORIGIN_PROJECT"
	// SessionCookieName unify name of oauth2 cookies
	SessionCookieName = "OAUTH2_SESSION"
)

// CallbackHandler handles the callback from Gitlab OAuth
func CallbackHandler(oauthConfig *conf.OAuthConf, sessions *_sessions.Sessions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		callbackState := r.URL.Query().Get("state")
		if callbackState == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		cookieState, err := r.Cookie(StateCookieName)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Prevent CSRF attacks by checking state from cookie and state from callback
		if cookieState.Value != callbackState {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		callbackCode := r.URL.Query().Get("code")
		if callbackCode == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		requestedProject, err := r.Cookie(OriginProjectCookieName)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		gTokens, err := gitlab.RequestNewTokens(oauthConfig, callbackCode)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		project, err := gitlab.FetchProject(gTokens.AccessToken, oauthConfig.ProviderDomain, requestedProject.Value)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		sessionID, err := _sessions.GenID()
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		expiresDate := time.Now().Add(time.Second * time.Duration(gTokens.ExpiresIn))

		sessions.Put(sessionID, gTokens.AccessToken, expiresDate, project)

		sessionCookie := http.Cookie{
			Name:   SessionCookieName,
			Domain: oauthConfig.AppDomain,
			// TODO : restrict this value ?
			Path:    "/",
			Expires: expiresDate,
		}
		sessionCookie.Value = sessionID
		http.SetCookie(w, &sessionCookie)

		originURI := "/"
		if cookieOriginURI, err := r.Cookie(OriginURICookieName); err == nil {
			originURI = cookieOriginURI.Value
		}

		http.Redirect(w, r, originURI, http.StatusTemporaryRedirect)
	}
}
