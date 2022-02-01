package jwt

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/cristalhq/jwt/v3"
	"github.com/factorysh/microdensity/claims"
	"github.com/factorysh/microdensity/httpcontext"
	"github.com/getsentry/sentry-go"
	"go.uber.org/zap"
)

// Auth will ensure JWT token is valid
func (j *JWTAuthenticator) Middleware(isBlocking bool) func(next http.Handler) http.Handler {
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
					zap.String("token payload", string(token.Payload())))
				if hub := sentry.GetHubFromContext(r.Context()); hub != nil {
					hub.WithScope(func(scope *sentry.Scope) {
						scope.SetExtra("jwt", token.RawClaims())
					})
				}
				err = j.Validate(token)
				if err != nil {
					l.Warn("rotten token", zap.Error(err))
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				l = l.With(zap.ByteString("payload", token.Payload()))
				r = r.WithContext(context.WithValue(r.Context(), httpcontext.JWT, token))
				var payload claims.Claims
				err = json.Unmarshal(token.Payload(), &payload)
				if err != nil {
					l.Warn("rotten payload",
						zap.Error(err))
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				user := payload.Owner
				r = r.WithContext(context.WithValue(r.Context(), httpcontext.User, user))
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
	return jwt.Parse([]byte(p))
}
