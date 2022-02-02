package middlewares

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
)

var (
	privateRSA1024   = mustParseRSAKey(privateRSA1024txt)
	privateRSA1024_2 = mustParseRSAKey(privateRSA1024txt2)
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
