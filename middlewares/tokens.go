package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/factorysh/microdensity/httpcontext"
	"github.com/factorysh/microdensity/oauth"
)

// Tokens check for tokens in headers and add them to context
func Tokens() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Add JWT to context if any
			if token, _, err := getJWT(r); err == nil && token != "" {
				ctx := context.WithValue(r.Context(), httpcontext.JWT, token)
				r = r.WithContext(ctx)
			}

			// Add AccessToken to context if any
			if token, err := r.Cookie(oauth.SessionCookieName); err == nil && token != nil {
				ctx := context.WithValue(r.Context(), httpcontext.AccessToken, token.Value)
				r = r.WithContext(ctx)
			}

			next.ServeHTTP(w, r)
		})
	}
}

// getJWT from Header or Cookie or Param
func getJWT(r *http.Request) (string, bool, error) {
	getters := []func(*http.Request) (string, bool, error){getTokenFromHeader, getTokenFromParam, getTokenFromCookies}

	for _, fun := range getters {
		token, add, err := fun(r)
		if err == nil && token != "" {
			return token, add, err
		}
	}

	return "", false, fmt.Errorf("All authentication mechanisms failed")
}

func getTokenFromHeader(r *http.Request) (string, bool, error) {
	h := r.Header.Get("Authorization")
	bToken := strings.Split(h, " ")
	if len(bToken) != 2 {
		return "", false, fmt.Errorf("Invalid authorization header %v", h)
	}

	if bToken[0] != "Bearer" {
		return "", false, fmt.Errorf("Authorization header is not a bearer token %v", h)
	}

	return bToken[1], false, nil
}

func getTokenFromParam(r *http.Request) (string, bool, error) {
	return r.URL.Query().Get("token"), true, nil
}

func getTokenFromCookies(r *http.Request) (string, bool, error) {
	cookie, err := r.Cookie("TOKEN")
	if err != nil {
		return "", false, err
	}

	return cookie.Value, false, nil
}

func addCookie(w http.ResponseWriter, value string) {
	expires := time.Now().Add(time.Duration(24) * time.Hour)

	cookie := http.Cookie{
		Name:    "TOKEN",
		Expires: expires,
		Value:   value,
	}

	http.SetCookie(w, &cookie)
}
