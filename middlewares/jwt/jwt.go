package jwt

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	_jwt "github.com/cristalhq/jwt/v3"
	owner "github.com/factorysh/microdensity/claims"
	"github.com/factorysh/microdensity/httpcontext"
	"github.com/getsentry/sentry-go"
	"go.uber.org/zap"
	"gopkg.in/square/go-jose.v2"
)

type JWTAuthenticator struct {
	jose.JSONWebKeySet
	verifier map[string]_jwt.Verifier
	logger   *zap.Logger
}

func NewJWTAuthenticator(gitlab string) (*JWTAuthenticator, error) {
	logger, err := zap.NewProduction()
	l := logger.With(zap.String("gitlab", gitlab))
	if err != nil {
		return nil, err
	}
	r, err := http.Get(fmt.Sprintf("%s/-/jwks", gitlab))
	if err != nil {
		l.Error("can't fetch Gitlab's jwks", zap.Error(err))
		return nil, err
	}
	defer r.Body.Close()

	var j JWTAuthenticator
	err = json.NewDecoder(r.Body).Decode(&j)
	if err != nil {
		l.Error("can't parse Gitlab's jwks JSON", zap.Error(err))
		return nil, err
	}
	j.logger = logger
	j.verifier = make(map[string]_jwt.Verifier)
	for _, k := range j.Keys {
		if k.Use != "sig" {
			continue
		}
		alg := _jwt.Algorithm(k.Algorithm)
		l = l.With(zap.String("kid", k.KeyID), zap.String("algo", k.Algorithm))
		var err error
		switch {
		case strings.HasPrefix(k.Algorithm, "HS"):
			j.verifier[k.KeyID], err = _jwt.NewVerifierHS(alg, k.Key.([]byte))
		case strings.HasPrefix(k.Algorithm, "RS"):
			j.verifier[k.KeyID], err = _jwt.NewVerifierRS(alg, k.Key.(*rsa.PublicKey))
		default:
			err = fmt.Errorf("unhandled algo : %s", k.Algorithm)
		}
		if err != nil {
			l.Error("Bad signer", zap.Error(err))
			return nil, err
		}
		l.Info("a signer")
	}
	return &j, err
}

func (j *JWTAuthenticator) Validate(t *_jwt.Token) error {
	l := j.logger.With(zap.ByteString("jwt", t.Raw()))
	for _, k := range j.Key(t.Header().KeyID) {
		if k.Algorithm != t.Header().Algorithm.String() {
			err := fmt.Errorf("algo mismatch : %s vs %s", k.Algorithm, t.Header().Algorithm.String())
			l.Error("Bad algo", zap.Error(err))
			return err
		}
		err := j.verifier[k.KeyID].Verify(t.Payload(), t.Signature())
		if err != nil {
			l.Error("verification faild", zap.Error(err))
			return err
		}
		return nil
	}
	err := fmt.Errorf("can't authenticate key %v", t)
	l.Warn("Signer not found")
	return err
}

func (j *JWTAuthenticator) ParseAndValidate(jwtRaw string) (*_jwt.Token, error) {
	token, err := _jwt.ParseString(jwtRaw)
	if err != nil {
		j.logger.Warn("Can't parse JWT", zap.Error(err))
		return nil, err
	}

	err = j.Validate(token)
	if err != nil {
		return nil, err
	}

	return token, nil
}

// Auth will ensure JWT token is valid
func (j *JWTAuthenticator) Middleware() func(next http.Handler) http.Handler {
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

			j.handleJWTAuth(token, w, r, next)
		})
	}
}

func (j *JWTAuthenticator) handleJWTAuth(token string, w http.ResponseWriter, r *http.Request, next http.Handler) {
	t, err := j.ParseAndValidate(token)

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
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
