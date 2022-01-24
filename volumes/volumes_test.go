package volumes

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testRootDir = "/tmp/microdensity/volumes"

func cleanDir() {
	os.RemoveAll(testRootDir)
}

func TestNewVolumes(t *testing.T) {
	tests := []struct {
		name     string
		root     string
		validate func(*testing.T, error)
	}{
		{name: "Valid dir", root: testRootDir, validate: func(t *testing.T, err error) {
			assert.NoError(t, err)
		}},
		{name: "Invalid dir", root: "/no", validate: func(t *testing.T, err error) {
			assert.Error(t, err)
		}},
	}

	cleanDir()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := New(tc.root)
			tc.validate(t, err)
			cleanDir()
		})
	}
}

func TestListByProject(t *testing.T) {

	cleanDir()
	v, err := New(testRootDir)
	assert.NoError(t, err)
	err = v.Request("group/project", "master", "uuid")
	assert.NoError(t, err)

	dirs, err := v.ByProject("group/project")
	assert.NoError(t, err)
	assert.Len(t, dirs, 1, "one folder should be found")

}
