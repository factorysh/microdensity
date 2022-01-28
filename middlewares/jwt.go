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

type JWKS struct {
	Keys []struct {
		Kty      string `json:"kty"`
		Kid      string `json:"kid"`
		E        string `json:"e"`
		N        string `json:"n"`
		Use      string `json:"use"`
		Alg      string `json:"alg"`
		verifier jwt.Verifier
	} `json:"keys"`
}

func NewJWTAuthenticator(gitlab string) (*JWKS, error) {
	r, err := http.Get(fmt.Sprintf("https://%s/-/jwks", gitlab))
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	var j JWKS
	err = json.NewDecoder(r.Body).Decode(&j)
	if err != nil {
		return nil, err
	}
	for _, k := range j.Keys {
		k.verifier, err = jwt.NewVerifierHS(jwt.Algorithm(k.Alg), []byte(k.N))
		if err != nil {
			return nil, err
		}
	}
	return &j, err
}

func (j *JWKS) ParseAndValidate(jwtRaw string) (*jwt.Token, error) {
	jj, err := jwt.ParseString(jwtRaw)
	if err != nil {
		return nil, err
	}
	// FIXME alg in RS*
	alg := jj.Header().Algorithm
	if alg != jwt.RS256 {
		return nil, fmt.Errorf("bad algo : %s", alg.String())
	}
	for _, k := range j.Keys {
		if k.verifier.Algorithm() == jj.Header().Algorithm {
			err = k.verifier.Verify(jj.Payload(), jj.Signature())
			if err != nil {
				return nil, err
			}
			return jj, nil
		}
	}

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
