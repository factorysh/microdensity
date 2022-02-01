package jwt

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/cristalhq/jwt/v3"
	_claims "github.com/factorysh/microdensity/claims"
	"github.com/factorysh/microdensity/httpcontext"
	"github.com/getsentry/sentry-go"
	"go.uber.org/zap"
)

// Auth will ensure JWT token is valid
func (j *JWTAuthenticator) Handler(isBlocking bool) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			l := j.logger.With(zap.String("url", r.URL.String()))
			token, err := getJWTToken(r)
			if err != nil {
				l.Warn("cant't read JWT token", zap.Error(err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if token == nil && isBlocking {
				l.Warn("jwt token not found, middleware blocking")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			if token != nil {
				l = l.With(zap.String("token header", string(token.RawHeader())),
					zap.String("token claims", string(token.RawClaims())))
				if hub := sentry.GetHubFromContext(r.Context()); hub != nil {
					hub.WithScope(func(scope *sentry.Scope) {
						scope.SetExtra("jwt", token.RawClaims())
					})
				}
				err = j.VerifySignature(token)
				if err != nil {
					l.Warn("rotten token", zap.Error(err))
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				l = l.With(zap.ByteString("claims", token.RawClaims()))
				r = r.WithContext(context.WithValue(r.Context(), httpcontext.JWT, token))
				var claims _claims.Claims
				err = json.Unmarshal(token.RawClaims(), &claims)
				if err != nil {
					l.Warn("Can't read claims JSON",
						zap.Error(err))
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				r = r.WithContext(context.WithValue(r.Context(), httpcontext.User, claims.UserLogin))
				r = r.WithContext(context.WithValue(r.Context(), httpcontext.RequestedProject, claims.ProjectPath))
			} else {
				l.Info("No JWT token, but it's not blocking")
			}
			next.ServeHTTP(w, r)
		})
	}
}

func getJWTToken(r *http.Request) (*jwt.Token, error) {
	p := r.Header.Get("PRIVATE-TOKEN")
	if p == "" {
		p = r.URL.Query().Get("private_token")
	}
	if p == "" {
		return nil, nil
	}
	return jwt.ParseString(p)
}
