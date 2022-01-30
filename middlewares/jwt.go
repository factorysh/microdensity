package middlewares

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/cristalhq/jwt/v3"
	owner "github.com/factorysh/microdensity/claims"
	"github.com/factorysh/microdensity/httpcontext"
	"github.com/getsentry/sentry-go"
	"gopkg.in/square/go-jose.v2"
)

type JWTAuthenticator struct {
	jose.JSONWebKeySet
	verifier map[string]jwt.Verifier
}

func NewJWTAuthenticator(gitlab string) (*JWTAuthenticator, error) {
	r, err := http.Get(fmt.Sprintf("%s/-/jwks", gitlab))
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	var j JWTAuthenticator
	err = json.NewDecoder(r.Body).Decode(&j)
	if err != nil {
		return nil, err
	}
	j.verifier = make(map[string]jwt.Verifier)
	for _, k := range j.Keys {
		if k.Use != "sig" {
			continue
		}
		alg := jwt.Algorithm(k.Algorithm)
		var err error
		switch {
		case strings.HasPrefix(k.Algorithm, "HS"):
			j.verifier[k.KeyID], err = jwt.NewVerifierHS(alg, k.Key.([]byte))
		case strings.HasPrefix(k.Algorithm, "RS"):
			j.verifier[k.KeyID], err = jwt.NewVerifierRS(alg, k.Key.(*rsa.PublicKey))
		}
		if err != nil {
			return nil, err
		}
	}
	return &j, err
}

func (j *JWTAuthenticator) Validate(t *jwt.Token) error {
	for _, k := range j.Key(t.Header().KeyID) {
		if k.Algorithm != t.Header().Algorithm.String() {
			return fmt.Errorf("algo mismatch : %s vs %s", k.Algorithm, t.Header().Algorithm.String())
		}
		err := j.verifier[k.KeyID].Verify(t.Payload(), t.Signature())
		if err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("can't authenticate key %v", t)
}

func (j *JWTAuthenticator) ParseAndValidate(jwtRaw string) (*jwt.Token, error) {
	token, err := jwt.ParseString(jwtRaw)
	if err != nil {
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
