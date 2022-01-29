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
	verifier []jwt.Verifier
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
	j.verifier = make([]jwt.Verifier, len(j.Keys))
	for i, k := range j.Keys {
		if k.Use != "sig" {
			continue
		}
		alg := jwt.Algorithm(k.Algorithm)
		var err error
		switch {
		case strings.HasPrefix(k.Algorithm, "HS"):
			j.verifier[i], err = jwt.NewVerifierHS(alg, k.Key.([]byte))
		case strings.HasPrefix(k.Algorithm, "RS"):
			j.verifier[i], err = jwt.NewVerifierRS(alg, k.Key.(*rsa.PublicKey))
		}
		if err != nil {
			return nil, err
		}
	}
	return &j, err
}

func (j *JWTAuthenticator) Validate(t *jwt.Token) error {
	for _, k := range j.Key(t.Header().KeyID) {
		if k.Algorithm == t.Header().Algorithm.String() {
		}
	}
	return fmt.Errorf("can't authenticate key %v", t)
}

func ParseAndValidate(jwk *jose.JSONWebKeySet, jwtRaw string) (*jwt.Token, error) {
	jj, err := jwt.ParseString(jwtRaw)
	if err != nil {
		return nil, err
	}
	// FIXME alg in RS*
	alg := jj.Header().Algorithm
	if alg != jwt.RS256 {
		return nil, fmt.Errorf("bad algo : %s", alg.String())
	}
	/*
		for _, k := range j.Keys {
			if k.verifier.Algorithm() == jj.Header().Algorithm {
				err = k.verifier.Verify(jj.Payload(), jj.Signature())
				if err != nil {
					return nil, err
				}
				return jj, nil
			}
		}
	*/

	return nil, fmt.Errorf("can't validate %s", jj.String())
}

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
