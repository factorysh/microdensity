package jwt

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/factorysh/microdensity/mockup"
	"github.com/stretchr/testify/assert"
)

func TestJWK(t *testing.T) {
	gitlab := httptest.NewServer(mockup.GitlabJWK(&privateRSA1024.PublicKey))
	defer gitlab.Close()

	a, err := NewJWTAuthenticator(gitlab.URL)
	assert.NoError(t, err)
	token, err := BuildJWT(privateRSA1024, map[string]interface{}{
		"name": "Bob",
	})
	assert.NoError(t, err)
	fmt.Println("claims", string(token.RawClaims()))
	err = a.VerifySignature(token)
	assert.NoError(t, err)

	brokenGitlab := httptest.NewServer(http.NewServeMux())
	defer brokenGitlab.Close()
	_, err = NewJWTAuthenticator(brokenGitlab.URL)
	assert.Error(t, err)

}

var (
	privateRSA1024 = MustParseRSAKey(privateRSA1024txt)
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
