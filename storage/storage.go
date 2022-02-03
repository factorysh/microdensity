package storage

import (
	"encoding/json"
	"fmt"
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
const latestFile = "latest"

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
	EnsureVolumesDir(*task.Task) error
}

// FSStore contains all storage data and primitives directly on the FS
type FSStore struct {
	root    string
	volumes *volumes.Volumes
}

var _ Storage = (*FSStore)(nil)

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

func (s *FSStore) taskLatestPath(t *task.Task) string {
	return filepath.Join(s.root, t.Service, t.Project, t.Branch, latestFile)
}

func taskFromJSON(path string) (*task.Task, error) {

	// read json file
	jsonFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	// decode json
	var t task.Task
	json.NewDecoder(jsonFile).Decode(&t)
	if err != nil {
		return nil, err
	}

	return &t, err
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
	defer f.Close()

	return json.NewEncoder(f).Encode(t)
}

// Get takes an id and return a task
func (s *FSStore) Get(id string) (*task.Task, error) {

	taskRootPath := ""
	err := filepath.WalkDir(s.root, func(path string, d fs.DirEntry, err error) error {
		// check for errors
		if err != nil {
			return err
		}

		// do not work inside volumes dir
		if d.Name() == volumesDir {
			return filepath.SkipDir
		}

		// if task not found
		if taskRootPath == "" && d.Name() == id {
			taskRootPath = path
		}

		return nil
	})

	// if an error occured return
	if err != nil {
		return nil, err
	}

	// if not found
	if taskRootPath == "" {
		return nil, fmt.Errorf("task with id %s not found", id)
	}

	return taskFromJSON(filepath.Join(taskRootPath, taskFile))
}

// All returns all the tasks for this storage
func (s *FSStore) All() ([]*task.Task, error) {
	var tasks []*task.Task
	err := filepath.WalkDir(s.root, func(path string, d fs.DirEntry, err error) error {
		// check for errors
		if err != nil {
			return err
		}

		// do not work inside volumes dir
		if d.Name() == volumesDir {
			return filepath.SkipDir
		}

		if d.Name() == taskFile {
			task, err := taskFromJSON(path)
			if err != nil {
				// FIXME: handle
				fmt.Println(err)
				return nil
			}

			tasks = append(tasks, task)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return tasks, nil
}

// Filter return all the tasks matching the required predicates from the filter function
func (s *FSStore) Filter(filterFn func(*task.Task) bool) ([]*task.Task, error) {
	// TODO: later
	return nil, nil
}

// Delete takes an id and delete a task in the fs
func (s *FSStore) Delete(id string) error {
	t, err := s.Get(id)
	if err != nil {
		return err
	}

	return os.RemoveAll(s.taskRootPath(t))
}

// SetLatest is used save task as latest for this branch
func (s *FSStore) SetLatest(t *task.Task) error {
	return os.WriteFile(s.taskLatestPath(t), []byte(t.Id.String()), 0664)
}

// GetLatest is used to get latest task for a service, project, branch
func (s *FSStore) GetLatest(service, project, branch string) (*task.Task, error) {
	content, err := os.ReadFile(filepath.Join(s.root, service, project, branch, latestFile))
	if err != nil {
		return nil, err
	}

	return s.Get(string(content))
}

// GetVolumePath is used to get the root volume path of a task
func (s *FSStore) GetVolumePath(t *task.Task) string {
	return filepath.Join(s.taskRootPath(t), volumesDir)
}

// EnsureVolumesDir is used to create required volume dirs
func (s *FSStore) EnsureVolumesDir(t *task.Task) error {
	return s.volumes.Create(t)
}
