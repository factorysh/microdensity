package gitlab

import (
	"testing"

	"github.com/factorysh/microdensity/conf"
	"github.com/stretchr/testify/assert"
)

func TestRequestNewTokens(t *testing.T) {
	mockUP := TestMockup()

	oauthConf := conf.OAuthConf{
		ProviderURL: mockUP.URL,
		AppID:       "some",
		AppSecret:   "some",
		AppURL:      "http://localhost:3000",
	}

	resp, err := RequestNewTokens(&oauthConf, "code")
	assert.NoError(t, err)
	assert.Equal(t, "access_token", resp.AccessToken)
}
