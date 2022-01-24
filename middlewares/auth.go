package middlewares

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cristalhq/jwt/v3"
	owner "github.com/factorysh/microdensity/claims"
	"github.com/factorysh/microdensity/httpcontext"
	"github.com/getsentry/sentry-go"
)

// Auth will ensure JWT token is valid
func Auth(key string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// if user was auth using oauth
			isOAuth, err := httpcontext.GetIsOAuth(r)
			if err == nil && isOAuth {
				next.ServeHTTP(w, r)
				return
			}

			token, err := httpcontext.GetJWT(r)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			handleJWTAuth(token, key, w, r, next)
		})
	}
}

func handleJWTAuth(token string, key string, w http.ResponseWriter, r *http.Request, next http.Handler) {
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

	next.ServeHTTP(w, r.WithContext(ctx))
}
