package middlewares

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJWK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`
		{
		  "keys": [
		    {
		      "kty": "RSA",
		      "kid": "kewiQq9jiC84CvSsJYOB-N6A8WFLSV20Mb-y7IlWDSQ",
		      "e": "AQAB",
		      "n": "5RyvCSgBoOGNE03CMcJ9Bzo1JDvsU8XgddvRuJtdJAIq5zJ8fiUEGCnMfAZI4of36YXBuBalIycqkgxrRkSOENRUCWN45bf8xsQCcQ8zZxozu0St4w5S-aC7N7UTTarPZTp4BZH8ttUm-VnK4aEdMx9L3Izo0hxaJ135undTuA6gQpK-0nVsm6tRVq4akDe3OhC-7b2h6z7GWJX1SD4sAD3iaq4LZa8y1mvBBz6AIM9co8R-vU1_CduxKQc3KxCnqKALbEKXm0mTGsXha9aNv3pLNRNs_J-cCjBpb1EXAe_7qOURTiIHdv8_sdjcFTJ0OTeLWywuSf7mD0Wpx2LKcD6ImENbyq5IBuR1e2ghnh5Y9H33cuQ0FRni8ikq5W3xP3HSMfwlayhIAJN_WnmbhENRU-m2_hDPiD9JYF2CrQneLkE3kcazSdtarPbg9ZDiydHbKWCV-X7HxxIKEr9N7P1V5HKatF4ZUrG60e3eBnRyccPwmT66i9NYyrcy1_ZNN8D1DY8xh9kflUDy4dSYu4R7AEWxNJWQQov525v0MjD5FNAS03rpk4SuW3Mt7IP73m-_BpmIhW3LZsnmfd8xHRjf0M9veyJD0--ETGmh8t3_CXh3I3R9IbcSEntUl_2lCvc_6B-m8W-t2nZr4wvOq9-iaTQXAn1Au6EaOYWvDRE",
		      "use": "sig",
		      "alg": "RS256"
		    },
		    {
		      "kty": "RSA",
		      "kid": "4i3sFE7sxqNPOT7FdvcGA1ZVGGI_r-tsDXnEuYT4ZqE",
		      "e": "AQAB",
		      "n": "4cxDjTcJRJFID6UCgepPV45T1XDz_cLXSPgMur00WXB4jJrR9bfnZDx6dWqwps2dCw-lD3Fccj2oItwdRQ99In61l48MgiJaITf5JK2c63halNYiNo22_cyBG__nCkDZTZwEfGdfPRXSOWMg1E0pgGc1PoqwOdHZrQVqTcP3vWJt8bDQSOuoZBHSwVzDSjHPY6LmJMEO42H27t3ZkcYtS5crU8j2Yf-UH5U6rrSEyMdrCpc9IXe9WCmWjz5yOQa0r3U7M5OPEKD1-8wuP6_dPw0DyNO_Ei7UerVtsx5XSTd-Z5ujeB3PFVeAdtGxJ23oRNCq2MCOZBa58EGeRDLR7Q",
		      "use": "sig",
		      "alg": "RS256"
		    }
		  ]
		}
`))
	}))
	defer srv.Close()
	j, err := NewJWTAuthenticator(srv.URL)
	fmt.Println("j", j)
	assert.NoError(t, err)
	assert.Len(t, j.Keys, 2)

}

var (
	privateRSA1024 = mustParseRSAKey(privateRSA1024txt)
	publicRSA1024  = privateRSA1024.PublicKey
)

const (
	// openssl genrsa -out rs256-1024-private.rsa 1024
	privateRSA1024txt = `-----BEGIN RSA PRIVATE KEY-----
MIICXwIBAAKBgQDbug04O5k4KLo1uQrQmDv9otupdu9buj5uqecy/lDZKVsx0LAU
C86eRZPzXcNltO0b47c+uFSyVxOjEvwua1Z8hfm9pqR8kcmxxceoSZb7iHdXC4L8
QR/dl/rb9K6mv7YR8TM5MxhPeSu4hbxhBM6EErOeXavUJd3PxsnGdeA/cwIDAQAB
AoGBAKo7DH7ifaRquUlh4SUWrHOmtvQl9u9j7XajH0H8kfqM9eA0RBZjx2ILmcJU
hEvJzmFrHM701HmOyOHwlXwJIOjM4xLI3U2zOfSWTvq729Wnu7nGl+GaSMc/40Lj
P4SPWSKHj0xg6RdzF4E8ItdZwX593mparjBoHrVIACIBBqMxAkEA9EyhWR0+UXM+
s9W9m8kjlSgYGpwmH6IBZktyeBfs0+KlSZSEJd7pv9qLCIk2f2B6WAKQRuFIxyYk
umgXwcFPvQJBAOZAIq+vPj+nWoO0IzHpD/0k/JjxClWX3m3p7M6HTIGqAq32rfOX
9l3Pm6W8nL0RjBZKCWom7XmGYY2mag5/5u8CQQCaspvJbncz5KJkBolWyPu7S/RX
hWGuzkvMlyIZYi0Zz3+TJHS59npWfvFjql/UMSfH63epKqeHVGQVlizVCLCRAkEA
pwH+JtBFpoYM8VrH7HvQTR122rh7dnohrDfwvB0HMUXPi79RjU68NG9RxnV4eusv
YTtyeLyjo3IFcGk0pC/BoQJBAJQbcTj2IbCHrx/QB/EcenbyjIQwT9QuAx9nkSLm
JBT2XB3g+NSO5sd54NgvmSA78gXsQ8i4HQSaY3fL6K4tI0o=
-----END RSA PRIVATE KEY-----`
)

func mustParseRSAKey(s string) *rsa.PrivateKey {
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
