package storage

import (
	"path/filepath"

	"github.com/factorysh/microdensity/task"
	"github.com/factorysh/microdensity/volumes"
)

// Storage describe all storage primitives
type Storage interface {
	Upsert(*task.Task) error
	Get(id string) (*task.Task, error)
	All() ([]*task.Task, error)
	Filter(func(*task.Task) bool) ([]*task.Task, error)
	Delete(id string) error
	SetLatest(*task.Task) error
	GetLatest(service, project, branch string) (*task.Task, error)
	GetVolumePath(*task.Task) error
}

// Store contains all storage data and primitives
type Store struct {
	root    string
	volumes *volumes.Volumes
}

// New inits a new store
func New(root string) (*Store, error) {
	v, err := volumes.New(filepath.Join(root, "volumes"))
	if err != nil {
		return nil, err
	}

	return &Store{
		root:    root,
		volumes: v,
	}, nil
}

// Upsert takes a task and write it to the underlying fs
func (s *Store) Upsert(t *task.Task) error {
	return nil
}

// Get takes an id and return a task
func (s *Store) Get(id string) (*task.Task, error) {
	return nil, nil
}

// All returns all the tasks for this storage
func (s *Store) All() ([]*task.Task, error) {
	return nil, nil
}

// Filter return all the tasks matching the required predicates from the filter function
func (s *Store) Filter(filterFn func(*task.Task) bool) ([]*task.Task, error) {
	return nil, nil
}

// Delete takes an id and delete a task in the fs
func (s *Store) Delete(id string) error {
	return nil
}

// SetLatest is used save task as latest for this branch
func (s *Store) SetLatest(t *task.Task) error {
	return nil
}

// GetLatest is used to get latest task for a service, project, branch
func (s *Store) GetLatest(service, project, branch string) (*task.Task, error) {
	return nil, nil
}

// GetVolumePath is used to get the root volume path of a task
func (s *Store) GetVolumePath(*task.Task) error {
	return nil
}

var _ Storage = (*Store)(nil)
