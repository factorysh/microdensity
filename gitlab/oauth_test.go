package gitlab

import (
	"testing"

	"github.com/factorysh/microdensity/conf"
	"github.com/stretchr/testify/assert"
)

func TestRequestNewTokens(t *testing.T) {
	mockUP := TestMockup()

	oauthConf := conf.OAuthConf{
		ProviderDomain: mockUP.URL,
		AppID:          "some",
		AppSecret:      "some",
		AppDomain:      "http://localhost:3000",
	}

	resp, err := RequestNewTokens(&oauthConf, "code")
	assert.NoError(t, err)
	assert.Equal(t, resp.AccessToken, "de6780bc506a0446309bd9362820ba8aed28aa506c71eedbe1c5c4f9dd350e54")
}
