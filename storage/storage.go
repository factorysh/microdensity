package storage

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/factorysh/microdensity/task"
	"github.com/factorysh/microdensity/volumes"
)

// DirMode is the default dirmode for volumes
const DirMode fs.FileMode = 0755

const volumesDir = "volumes"

const taskFile = "task.json"

// Storage describe all storage primitives
type Storage interface {
	Upsert(*task.Task) error
	Get(id string) (*task.Task, error)
	All() ([]*task.Task, error)
	Filter(func(*task.Task) bool) ([]*task.Task, error)
	Delete(id string) error
	SetLatest(*task.Task) error
	GetLatest(service, project, branch string) (*task.Task, error)
	GetVolumePath(*task.Task) string
}

// FSStore contains all storage data and primitives directly on the FS
type FSStore struct {
	root    string
	volumes *volumes.Volumes
}

// NewFSStore inits a new filesystem store
func NewFSStore(root string) (*FSStore, error) {
	err := os.MkdirAll(root, DirMode)
	if err != nil {
		return nil, err
	}

	v, err := volumes.New(root)
	if err != nil {
		return nil, err
	}

	return &FSStore{
		root:    root,
		volumes: v,
	}, nil
}

func (s *FSStore) taskRootPath(t *task.Task) string {
	return filepath.Join(s.root, t.Service, t.Project, t.Branch, t.Id.String())
}

func (s *FSStore) taskFilePath(t *task.Task) string {
	return filepath.Join(s.taskRootPath(t), taskFile)
}

// Upsert takes a task and write it to the underlying fs
func (s *FSStore) Upsert(t *task.Task) error {
	// construct the tree on the FS
	err := os.MkdirAll(s.GetVolumePath(t), DirMode)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(s.taskFilePath(t), os.O_CREATE+os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	return json.NewEncoder(f).Encode(t)
}

// Get takes an id and return a task
func (s *FSStore) Get(id string) (*task.Task, error) {
	return nil, nil
}

// All returns all the tasks for this storage
func (s *FSStore) All() ([]*task.Task, error) {
	return nil, nil
}

// Filter return all the tasks matching the required predicates from the filter function
func (s *FSStore) Filter(filterFn func(*task.Task) bool) ([]*task.Task, error) {
	return nil, nil
}

// Delete takes an id and delete a task in the fs
func (s *FSStore) Delete(id string) error {
	return nil
}

// SetLatest is used save task as latest for this branch
func (s *FSStore) SetLatest(t *task.Task) error {
	return nil
}

// GetLatest is used to get latest task for a service, project, branch
func (s *FSStore) GetLatest(service, project, branch string) (*task.Task, error) {
	return nil, nil
}

// GetVolumePath is used to get the root volume path of a task
func (s *FSStore) GetVolumePath(t *task.Task) string {
	return filepath.Join(s.taskRootPath(t), volumesDir)
}

var _ Storage = (*FSStore)(nil)
