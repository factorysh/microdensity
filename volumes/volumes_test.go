package volumes

import (
	"os"
	"strings"
	"testing"
	"time"

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

func TestListByProjectByBranch(t *testing.T) {
	defer cleanDir()

	v, err := New(testRootDir)
	assert.NoError(t, err)
	var first uuid.UUID
	var last uuid.UUID
	for i := 0; i < 10; i++ {
		id, err := uuid.NewUUID()
		assert.NoError(t, err)
		c := time.Now()
		if first == uuid.Nil {
			first = id
		}
		last = id
		err = v.Create(&task.Task{
			Project:  "group%2Fproject",
			Branch:   "master",
			Id:       id,
			Creation: c,
		})
		assert.NoError(t, err)
	}

	dirs, err := v.ByProjectByBranch("group%2Fproject", "master")
	assert.NoError(t, err)
	assert.Len(t, dirs, 10, "one folder should be found")
	assert.True(t, strings.HasSuffix(dirs[0], first.String()), dirs[0])
	assert.True(t, strings.HasSuffix(dirs[9], last.String()))
	cleanDir()
}
