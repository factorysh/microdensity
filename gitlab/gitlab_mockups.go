package gitlab

import (
	"net/http"
	"net/http/httptest"
	"strings"
)

// DummyProject is a Gitlab project usable in various tests
var DummyProject = ProjectInfo{
	PathWithNamespace: "group/project",
	Name:              "group/project",
	ID:                42,
	Archived:          false,
}

// TestMockup spins an Gitlab API httptest server
func TestMockup() *httptest.Server {
	authHeader := "Bearer access_token"
	pInfo := `
		{
			"name": "project",
			"archived": false,
			"path_with_namespace": "group/project",
			"permissions": {
				"project_access": null,
				"group_access": {
					"access_level": 30,
					"notification_level": 3
				}
			}
		}
	`

	// from https://docs.gitlab.com/ee/api/oauth2.html
	OAuthResp := `
		{
			"access_token": "access_token",
			"token_type": "bearer",
			"expires_in": 7200,
			"refresh_token": "8257e65c97202ed1726cf9571600918f3bffb2544b26e00a61df9897668c33a1",
			"created_at": 1607635748
		}
	`

	return httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api/v4") && r.Header["Authorization"][0] != authHeader {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			switch r.URL.Path {
			case "/api/v4/projects/group/project":
				_, err := w.Write([]byte(pInfo))
				if err != nil {
					panic(err)
				}
				w.WriteHeader(http.StatusOK)
			case "/oauth/token":
				_, err := w.Write([]byte(OAuthResp))
				if err != nil {
					panic(err)
				}
				w.WriteHeader(http.StatusOK)
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
}
