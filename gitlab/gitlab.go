package gitlab

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// FetchProject will try to fetch requested project by impersonating gitlab user using the access token
func FetchProject(token string, gitlabDomain string, requestedProject string) (*ProjectInfo, error) {
	if requestedProject == "" {
		return nil, fmt.Errorf("error requested project can't be blank")
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("https://%s/api/v4/projects/%s", gitlabDomain, requestedProject), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error when getting project for user, status code : %v", resp.StatusCode)
	}

	// Decode json to gitlab project struct
	var project ProjectInfo
	err = json.Unmarshal(body, &project)
	if err != nil {
		return nil, fmt.Errorf("error when decoding project response body : %v", err)
	}

	return &project, err
}

// ProjectInfo struct contains all the required data about a Gitlab Project
type ProjectInfo struct {
	Name              string            `json:"name"`
	ID                int               `json:"id"`
	Archived          bool              `json:"archived"`
	PathWithNamespace string            `json:"path_with_namespace"`
	Permissions       projectPermission `json:"permissions"`
}

type permissionAccess struct {
	AccessLevel int `json:"access_level"`
}

type projectPermission struct {
	ProjectAccess *permissionAccess `json:"project_access"`
	GroupAccess   *permissionAccess `json:"group_access"`
}
