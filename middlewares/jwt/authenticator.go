package jwt

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	_jwt "github.com/cristalhq/jwt/v3"
	"github.com/davecgh/go-spew/spew"
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
	spew.Dump(j.Keys)
	for _, k := range j.Keys {
		if k.Use != "sig" {
			continue
		}
		alg := _jwt.Algorithm(k.Algorithm)
		l = l.With(zap.String("kid", k.KeyID), zap.String("algo", k.Algorithm))
		var err error
		fmt.Println("algo", k.Algorithm)
		switch {
		case strings.HasPrefix(k.Algorithm, "HS"):
			j.verifier[k.KeyID], err = _jwt.NewVerifierHS(alg, k.Key.([]byte))
		case strings.HasPrefix(k.Algorithm, "RS"):
			j.verifier[k.KeyID], err = _jwt.NewVerifierRS(alg, k.Key.(*rsa.PublicKey))
		default:
			err = fmt.Errorf("unhandled algo : %s %v", k.Algorithm, k)
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
