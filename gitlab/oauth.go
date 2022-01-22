package gitlab

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/factorysh/microdensity/conf"
	"github.com/factorysh/microdensity/server"
)

// OAuthResponse allows easy parsing of /oauth/token Gitlab response
type OAuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

// RequestNewTokens uses Gitlab OAuth to generate new access tokens
func RequestNewTokens(c *conf.OAuthConf, callbackCode string) (*OAuthResponse, error) {
	// Request JWT tokens
	resp, err := http.Post(fmt.Sprintf("%s%s?%s", c.ProviderDomain, "/oauth/token", newRequestTokenParams(c, callbackCode).Encode()), "applications/json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		// in case of error, just get the raw message for debug
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("error when requesting access token: %v", string(body))
	}

	var gTokens OAuthResponse
	err = json.NewDecoder(resp.Body).Decode(&gTokens)
	if err != nil {
		return nil, fmt.Errorf("error when parsing oauth response: %v", err)
	}

	return &gTokens, err
}

func newRequestTokenParams(c *conf.OAuthConf, code string) url.Values {
	// /oauth/token?&client_id=APP_ID&client_secret=APP_SECRET&code=RETURNED_CODE&grant_type=authorization_code&redirect_uri=REDIRECT_URI
	values := url.Values{}
	values.Add("client_id", c.AppID)
	values.Add("client_secret", c.AppSecret)
	values.Add("code", code)
	values.Add("grant_type", "authorization_code")
	values.Add("redirect_uri", fmt.Sprintf("%s%s", c.AppDomain, server.OAuthCallbackEndpoint))

	return values
}
