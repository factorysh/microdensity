package middlewares

import (
	"net/http"

	"github.com/factorysh/microdensity/conf"
	"github.com/factorysh/microdensity/httpcontext"
	"github.com/factorysh/microdensity/middlewares/jwt"
	"github.com/factorysh/microdensity/middlewares/oauth2"
	_sessions "github.com/factorysh/microdensity/sessions"
	"go.uber.org/zap"
)

type JWTOrOAuth2 struct {
	authenticator *jwt.JWTAuthenticator
	oauthConfig   *conf.OAuthConf
	sessions      *_sessions.Sessions
	logger        *zap.Logger
}

func NewJWTOrOauth2(
	authenticator *jwt.JWTAuthenticator,
	oauthConfig *conf.OAuthConf,
	sessions *_sessions.Sessions,
) (*JWTOrOAuth2, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	return &JWTOrOAuth2{
		authenticator: authenticator,
		oauthConfig:   oauthConfig,
		sessions:      sessions,
		logger:        logger,
	}, nil
}

func (j *JWTOrOAuth2) Middleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			l := j.logger.With(zap.String("url", r.URL.String()))
			project, err := httpcontext.GetRequestedProject(r)
			if err != nil {
				l.Warn("Project error", zap.Error(err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			l = l.With(zap.String("project", project))
			j.authenticator.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				jwtRaw := r.Context().Value(httpcontext.JWT)
				if jwtRaw == nil {
					oauth2.OAuth2(j.oauthConfig, j.sessions)(next)
				} else {
					next.ServeHTTP(w, r)
				}
			}))
		})
	}
}
