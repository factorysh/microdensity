package mockup

import (
	"crypto/rsa"
	"encoding/json"
	"net/http"

	"gopkg.in/square/go-jose.v2"
)

func GitlabJWK(public *rsa.PublicKey) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		j := jose.JSONWebKey{
			Key:   public,
			Use:   "sig",
			KeyID: "someID",
		}
		err := json.NewEncoder(w).Encode(j)
		if err != nil {
			panic(err)
		}
	})

}
