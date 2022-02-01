package mockup

import (
	"crypto/rsa"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"net/http"

	"gopkg.in/square/go-jose.v2"
)

// GitlabJWK exposes -/jwks
func GitlabJWK(public *rsa.PublicKey) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		kid := Kid(public)
		j := jose.JSONWebKey{
			Key:       public,
			Use:       "sig",
			KeyID:     kid,
			Algorithm: "RS256",
		}
		jwks := jose.JSONWebKeySet{
			Keys: []jose.JSONWebKey{j},
		}
		err := json.NewEncoder(w).Encode(jwks)
		if err != nil {
			panic(err)
		}
	})

}

func Kid(public *rsa.PublicKey) string {
	h := sha1.New()
	return base64.RawURLEncoding.EncodeToString(h.Sum(public.N.Bytes()))[:16]
}
