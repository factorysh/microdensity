package conf

// OAuthConf wraps all the config required for OAuth2 mechanism
type OAuthConf struct {
	ProviderURL string `yaml:"provider_url"`
	AppID       string `yaml:"app_id"`
	AppSecret   string `yaml:"app_secret"`
	AppURL      string `yaml:"app_url"`
}
