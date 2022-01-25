package volumes

import (
	"os"
	"strings"
	"testing"

	"github.com/factorysh/microdensity/task"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

const testRootDir = "/tmp/microdensity/volumes"

func cleanDir() {
	os.RemoveAll(testRootDir)
}

func TestNewVolumes(t *testing.T) {
	defer cleanDir()

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
	defer cleanDir()

	v, err := New(testRootDir)
	assert.NoError(t, err)
	id, err := uuid.NewUUID()
	assert.NoError(t, err)
	err = v.Request(&task.Task{
		Project: "group/project",
		Branch:  "master",
		Id:      id,
	})
	assert.NoError(t, err)
	id2, err := uuid.NewUUID()
	assert.NoError(t, err)
	err = v.Request(&task.Task{
		Project: "group/project",
		Branch:  "dev",
		Id:      id2,
	})
	assert.NoError(t, err)

	dirs, err := v.ByProject("group/project")
	assert.NoError(t, err)
	assert.Len(t, dirs, 2, "one folder should be found")
	assert.True(t, strings.HasSuffix(dirs[0], "/another"), dirs[0])
	assert.True(t, strings.HasSuffix(dirs[1], "/uuid"), dirs[1])
	cleanDir()
}

func TestListByProjectByBranch(t *testing.T) {
	defer cleanDir()

	v, err := New(testRootDir)
	assert.NoError(t, err)
	for i := 0; i < 10; i++ {
		id, err := uuid.NewUUID()
		assert.NoError(t, err)
		err = v.Request(&task.Task{
			Project: "group/project",
			Branch:  "master",
			Id:      id,
		})
		assert.NoError(t, err)
	}

	dirs, err := v.ByProjectByBranch("group/project", "master")
	assert.NoError(t, err)
	assert.Len(t, dirs, 10, "one folder should be found")
	assert.Contains(t, dirs[0], "0")
	assert.Contains(t, dirs[9], "9")
	cleanDir()
}
