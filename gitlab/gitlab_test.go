package gitlab

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFetchProject(t *testing.T) {
	mockUP := TestMockup()
	defer mockUP.Close()

	tests := []struct {
		name        string
		token       string
		projectName string
		project     *ProjectInfo
		err         error
	}{
		{name: "valid token, valid project", token: "access_token", project: &DummyProject, projectName: "group/project", err: nil},
		{name: "invalid token, valid project", token: "nop", project: nil, projectName: "group/project", err: fmt.Errorf("error when getting project `group/project`, status code : 403")},
		{name: "valid token, invalid project", token: "access_token", project: nil, projectName: "group/nop", err: fmt.Errorf("error when getting project `group/nop`, status code : 404")},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := FetchProject(tc.token, mockUP.URL, tc.projectName)
			assert.Equal(t, err, tc.err, "with test `%s`, unexpected error received", tc.name)
		})
	}

}
