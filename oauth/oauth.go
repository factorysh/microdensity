package oauth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

const (
	// CallbackEndpoint route
	CallbackEndpoint = "/oauth/callback"
	// StateCookieName unify name of oauth2 cookies
	StateCookieName = "OAUTH2_STATE"
	// OriginURICookieName unify name of oauth2 cookies
	OriginURICookieName = "OAUTH2_ORIGIN_URI"
	// OriginProjectCookieName unify name of oauth2 cookie
	OriginProjectCookieName = "OAUTH2_ORIGIN_PROJECT"
)

// Config wraps all the config required for OAuth2 mechanism
type Config struct {
	ProviderDomain string
	AppID          string
	AppSecret      string
	AppDomain      string
}

// NewConfigFromEnv creates an OAuth config from environment variables
func NewConfigFromEnv() (*Config, error) {
	values := make([]string, 4)
	for i, key := range []string{"OAUTH_PROVIDER_DOMAIN", "OAUTH_APPID", "OAUTH_APPSECRET", "OAUTH_APPDOMAIN"} {
		v := os.Getenv(key)
		if v == "" {
			return nil, fmt.Errorf("Missing %s environment variable", key)
		}
		values[i] = v
	}

	return &Config{
		ProviderDomain: values[0],
		AppID:          values[1],
		AppSecret:      values[2],
		AppDomain:      values[3],
	}, nil
}

func newRequestTokenParams(c *Config, code string) url.Values {
	// /oauth/token?&client_id=APP_ID&client_secret=APP_SECRET&code=RETURNED_CODE&grant_type=authorization_code&redirect_uri=REDIRECT_URI
	values := url.Values{}
	values.Add("client_id", c.AppID)
	values.Add("client_secret", c.AppSecret)
	values.Add("code", code)
	values.Add("grant_type", "authorization_code")
	values.Add("redirect_uri", fmt.Sprintf("https://%s%s", c.AppDomain, CallbackEndpoint))

	return values
}

func requestNewTokens(c *Config, callbackCode string) ([]byte, error) {
	// Request JWT tokens
	resp, err := http.Post(fmt.Sprintf("https://%s%s?%s", c.ProviderDomain, "/oauth/token", newRequestTokenParams(c, callbackCode).Encode()), "applications/json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// parse response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error when requesting access token: %v", string(body))
	}

	return body, err
}

// gitlabTokens allows  easy parsing of /oauth/token Gitlab response
type gitlabTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

// CallbackHandler handles the callback from Gitlab OAuth
func CallbackHandler(oauthConfig *Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		callbackState := r.URL.Query().Get("state")
		if callbackState == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		cookieState, err := r.Cookie(StateCookieName)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Prevent CSRF attacks by checking state from cookie and state from callback
		if cookieState.Value != callbackState {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		callbackCode := r.URL.Query().Get("code")
		if callbackCode == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		body, err := requestNewTokens(oauthConfig, callbackCode)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// if response is a success, parse tokens
		var gTokens gitlabTokens
		err = json.Unmarshal(body, &gTokens)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// TODO: check access to requested project using access token
		// if ok add to current sessions (map[string]AppAccess)
		// AppAccess contains : expires, allowed project
		// add Session token to cookies (path is restricted to project)

		originURI := "/"
		if cookieOriginURI, err := r.Cookie(OriginURICookieName); err == nil {
			originURI = cookieOriginURI.Value
		}

		http.Redirect(w, r, originURI, http.StatusTemporaryRedirect)
	}
}
