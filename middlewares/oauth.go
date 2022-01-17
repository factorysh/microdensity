package middlewares

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
)

// OAuth will trigger an OAuth flow if no auth token is found
func OAuth(domain string, appid string, redirectDomain string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// get an auth token
			token, _, err := getToken(r)
			// if a token is found, pass to next middleware
			if token != "" && err == nil {
				next.ServeHTTP(w, r)
				return
			}

			state := uuid.NewString()

			// if token not found, ask for a new one
			values := url.Values{}
			values.Add("client_id", appid)
			values.Add("redirect_uri", fmt.Sprintf("https://%s/oauth/callback", redirectDomain))
			values.Add("response_type", "code")
			values.Add("state", state)
			fmt.Fprintf(w, "Please login with Gitlab : https://%s/oauth/authorize?%s", domain, values.Encode())

			cookie := http.Cookie{
				Name:    "OAUTH2_STATE",
				Domain:  redirectDomain,
				Path:    "/oauth/callback",
				Expires: time.Now().AddDate(0, 1, 0),
			}

			cookie.Value = state

			http.SetCookie(w, &cookie)
			w.WriteHeader(http.StatusOK)
		})
	}
}
