package jwt

import (
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
	privateRSA1024   = MustParseRSAKey(privateRSA1024txt)
	privateRSA1024_2 = MustParseRSAKey(privateRSA1024txt2)
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
	privateRSA1024txt2 = `-----BEGIN RSA PRIVATE KEY-----
MIICXgIBAAKBgQDRhfCQOaKVTZdX3W4uSQP4rHe8lie+7Ti6gmm2nDA5sKRPbcXm
xK2Mhv6id4V1c8ydbIfsiCBBLx8yY7nKJEg9hJtpeoCzS0h6cX5/kGnSo6fZiRQ5
iHi8k4rxmOhUOF6WGVv2N1T6XmONB2RJ5t9j4XDaG0oqyo9DKsRrNaeY5wIDAQAB
AoGAWbly8EBOOIO2uODRSy7nbXll+TOQJ7nsnio03Qd7u2jCpGUM56r36wLwTmDC
nS6OxCdy+b69mUx1np2INWFeMXuZDg0mXzJZ420OwSzpfRHiT2x4o2EYgAomePDf
9MTiRz5hli+F/qSBmnoV7QeqveSZ/B8ny5+fmqtw0dlw9gkCQQDoRj19otEjs+tJ
1WIfYSkSs6l3pnBVbiDD033PG0CvwzkjSwRqRjh73bF9B8zKo68/zbvWQ+p4RKKu
qMcTXbPrAkEA5uzHAkWjrsf2P6h5b84zoOwmFPMCbA/qjY1o3dCjEujOF2X7Ffvd
ZoCThxLk2IQUvlEd6aA8g44+RW/w6vn79QJBAOPPCD4xszd2Hf2TSCKIs7UA+uQ8
HI7dbUtDIXBARWhda6vexpzI9FsgKxT60nOoqJhGWsUiZVPB1WDCbkXjMDMCQQCR
x/aWg5ois8/MPjJzl8xWEd60qPjleWLMe/Iw3g6k2F2KvfG13ivWEuOPiSj5WuCx
iQoGPAcX0guT0GhaHvilAkEA4m0ZkENRjT8ZsvskRlFR5dI081eywzWJ7uPnh/kk
ZjKyM9LAK2CCWDAnXCA65Yc9Fas4iKS0icxLdDH3MWPffQ==
-----END RSA PRIVATE KEY-----`
)
