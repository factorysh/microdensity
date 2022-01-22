package gitlab

import (
	"net/http"
	"net/http/httptest"
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

	return httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.Header["Authorization"][0] != authHeader {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			switch r.URL.Path {
			case "/api/v4/projects/group/project":
				w.Write([]byte(pInfo))
				w.WriteHeader(http.StatusOK)
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))
}
