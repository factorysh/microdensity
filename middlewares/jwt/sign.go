package jwt

import (
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"

	_jwt "github.com/cristalhq/jwt/v3"
)

func BuildJWT(private *rsa.PrivateKey, claims interface{}) (*_jwt.Token, error) {
	signer, err := _jwt.NewSignerRS(_jwt.RS256, private)
	if err != nil {
		return nil, err
	}
	h := sha256.New()
	// What is kid : https://www.rfc-editor.org/rfc/rfc7515#section-4.1.4
	kid := base64.RawURLEncoding.EncodeToString(h.Sum(private.PublicKey.N.Bytes()))[:16]
	fmt.Println("kid", kid)
	j, err := _jwt.NewBuilder(signer,
		_jwt.WithKeyID(kid),
	).Build(claims)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func MustParseRSAKey(s string) *rsa.PrivateKey {
	block, _ := pem.Decode([]byte(s))
	if block == nil {
		panic("invalid PEM")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		panic(err)
	}
	return key
}
