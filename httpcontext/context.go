package httpcontext

import (
	"fmt"
	"net/http"
)

// Key is a special type used as a context key
type Key string

const (
	// JWT is the key used to read a JWT value from a context
	JWT Key = "JWT"
	// AccessToken is the key used to read an AccessToken value from a context
	AccessToken Key = "AccessToken"
	// IsOAuth is the key used to check is user is auth from a context
	IsOAuth Key = "IsOAuth"
)

// GetAccessToken is used to fetch an access token from request context
func GetAccessToken(r *http.Request) (string, error) {
	rawAccessToken := r.Context().Value(AccessToken)
	if rawAccessToken == nil {
		return "", fmt.Errorf("no access token found in httpcontext")
	}

	accessToken, ok := rawAccessToken.(string)
	if !ok {
		return "", fmt.Errorf("error when casting access token")
	}

	return accessToken, nil
}

// GetJWT is used to fetch a JWT token from request context
func GetJWT(r *http.Request) (string, error) {
	rawJWT := r.Context().Value(JWT)
	if rawJWT == nil {
		return "", fmt.Errorf("no jwt token found in httpcontext")
	}

	token, ok := rawJWT.(string)
	if !ok {
		return "", fmt.Errorf("error when casting JWT token")
	}

	return token, nil

}

// GetIsOAuth is used to fetch if used is auth from request context
func GetIsOAuth(r *http.Request) (bool, error) {
	rawIsAuth := r.Context().Value(IsOAuth)
	if rawIsAuth == nil {
		return false, fmt.Errorf("no IsOAuth found in httpcontext")
	}

	token, ok := rawIsAuth.(bool)
	if !ok {
		return false, fmt.Errorf("error when casting IsOAuth value")
	}

	return token, nil
}
