package middlewares

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cristalhq/jwt/v3"
	owner "github.com/factorysh/microdensity/claims"
	"github.com/getsentry/sentry-go"
)

// Auth will ensure JWT token is valid
func Auth(key string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, add, err := getToken(r)
			if err != nil {
				fmt.Println(err)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			verifier, err := jwt.NewVerifierHS(jwt.HS256, []byte(key))
			if err != nil {
				fmt.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			t, err := jwt.ParseAndVerify([]byte(token), verifier)
			if err != nil {
				fmt.Println(err)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// owner claims
			var claims owner.Claims
			err = json.Unmarshal(t.RawClaims(), &claims)
			if err != nil {
				fmt.Println(err)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			err = claims.Validate()
			if err != nil {
				fmt.Println(err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			if hub := sentry.GetHubFromContext(r.Context()); hub != nil {
				hub.WithScope(func(scope *sentry.Scope) {
					scope.SetExtra("jwt", claims)
				})
			}

			ctx := claims.ToCtx(r.Context())

			if add {
				addCookie(w, token)
			}

			next.ServeHTTP(w, r.WithContext(ctx))

		})
	}
}

// getToken from Header or Cookie or Param
func getToken(r *http.Request) (string, bool, error) {
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
