package conf

import (
	"fmt"
	"os"
)

// OAuthConf wraps all the config required for OAuth2 mechanism
type OAuthConf struct {
	ProviderDomain string
	AppID          string
	AppSecret      string
	AppDomain      string
}

// NewOAuthConfigFromEnv creates an OAuth config from environment variables
func NewOAuthConfigFromEnv() (*OAuthConf, error) {
	values := make([]string, 4)
	for i, key := range []string{"OAUTH_PROVIDER_DOMAIN", "OAUTH_APPID", "OAUTH_APPSECRET", "OAUTH_APPDOMAIN"} {
		v := os.Getenv(key)
		if v == "" {
			return nil, fmt.Errorf("Missing %s environment variable", key)
		}
		values[i] = v
	}

	return &OAuthConf{
		ProviderDomain: values[0],
		AppID:          values[1],
		AppSecret:      values[2],
		AppDomain:      values[3],
	}, nil
}
