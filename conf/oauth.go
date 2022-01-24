package conf

import (
	"fmt"
	"os"
)

// OAuthConf wraps all the config required for OAuth2 mechanism
type OAuthConf struct {
	ProviderURL string
	AppID       string
	AppSecret   string
	AppURL      string
}

// NewOAuthConfigFromEnv creates an OAuth config from environment variables
func NewOAuthConfigFromEnv() (*OAuthConf, error) {
	values := make([]string, 4)
	for i, key := range []string{"OAUTH_PROVIDER_URL", "OAUTH_APPID", "OAUTH_APPSECRET", "OAUTH_APPURL"} {
		v := os.Getenv(key)
		if v == "" {
			return nil, fmt.Errorf("Missing %s environment variable", key)
		}
		values[i] = v
	}

	return &OAuthConf{
		ProviderURL: values[0],
		AppID:       values[1],
		AppSecret:   values[2],
		AppURL:      values[3],
	}, nil
}
