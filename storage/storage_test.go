package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/factorysh/microdensity/task"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var dummyTask = &task.Task{
	Id:      uuid.New(),
	Service: "demo",
	Project: "group%20project",
	Branch:  "main",
	Commit:  "01279848527693d126de60ec7b355924c96d2957",
}

const defaultTestDir = "/tmp/microdensity/data"

func cleanUp() {
	os.RemoveAll(defaultTestDir)
}
func TestNew(t *testing.T) {
	_, err := NewFSStore(defaultTestDir)
	defer cleanUp()
	assert.NoError(t, err)
}

func TestUpsert(t *testing.T) {
	s, err := NewFSStore(defaultTestDir)
	defer cleanUp()
	assert.NoError(t, err)

	err = s.Upsert(dummyTask)
	assert.NoError(t, err)

	files, err := os.ReadDir(filepath.Join(defaultTestDir,
		dummyTask.Service,
		dummyTask.Project,
		dummyTask.Branch,
		dummyTask.Id.String()))

	var fnames []string
	for _, file := range files {
		fnames = append(fnames, file.Name())
	}

	assert.Contains(t, fnames, "task.json")
}

func TestGet(t *testing.T) {
	s, err := NewFSStore(defaultTestDir)
	defer cleanUp()
	assert.NoError(t, err)

	err = s.Upsert(dummyTask)
	assert.NoError(t, err)

	task, err := s.Get(dummyTask.Id.String())
	assert.NoError(t, err)
	assert.Equal(t, dummyTask, task)
}

func TestAll(t *testing.T) {
	s, err := NewFSStore(defaultTestDir)
	defer cleanUp()
	assert.NoError(t, err)

	err = s.Upsert(dummyTask)
	assert.NoError(t, err)

	tasks, err := s.All()
	assert.NoError(t, err)
	assert.Len(t, tasks, 1)
}

func TestDelete(t *testing.T) {
	s, err := NewFSStore(defaultTestDir)
	defer cleanUp()
	assert.NoError(t, err)

	err = s.Upsert(dummyTask)
	assert.NoError(t, err)

	err = s.Delete(dummyTask.Id.String())
	assert.NoError(t, err)

	_, err = os.ReadDir(filepath.Join(defaultTestDir,
		dummyTask.Service,
		dummyTask.Project,
		dummyTask.Branch,
		dummyTask.Id.String()))

	assert.Error(t, err)
}

func TestSetLatest(t *testing.T) {
	s, err := NewFSStore(defaultTestDir)
	defer cleanUp()
	assert.NoError(t, err)

	err = s.Upsert(dummyTask)
	assert.NoError(t, err)

	err = s.SetLatest(dummyTask)
	assert.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(defaultTestDir,
		dummyTask.Service,
		dummyTask.Project,
		dummyTask.Branch,
		"latest"))
	assert.NoError(t, err)
	assert.Equal(t, dummyTask.Id.String(), string(content))
}

func TestGetLateset(t *testing.T) {
	s, err := NewFSStore(defaultTestDir)
	defer cleanUp()
	assert.NoError(t, err)

	err = s.Upsert(dummyTask)
	assert.NoError(t, err)

	err = s.SetLatest(dummyTask)
	assert.NoError(t, err)

	task, err := s.GetLatest(dummyTask.Service, dummyTask.Project, dummyTask.Branch)
	assert.NoError(t, err)
	assert.Equal(t, dummyTask, task)
}
