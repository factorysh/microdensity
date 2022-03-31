package storage

/*
µdensity can be rebooted without loosing task.Task
*/

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

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
	GetByCommit(service, project, branch, commit string, latest bool) (*task.Task, error)
	All() ([]*task.Task, error)
	Filter(func(*task.Task) bool) ([]*task.Task, error)
	Delete(id string) error
	SetLatest(*task.Task) error
	GetLatest(service, project, branch string) (*task.Task, error)
	GetVolumePath(*task.Task) string
	EnsureVolumesDir(*task.Task) error
	Prune(time.Duration, bool) (int64, error)
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
	path = filepath.Clean(path)
	// read json file
	jsonFile, err := os.Open(path) //#nosec I don't know what a correct is
	if err != nil {
		return nil, err
	}

	// decode json
	var t task.Task
	err = json.NewDecoder(jsonFile).Decode(&t)
	err2 := jsonFile.Close()
	if err != nil {
		return nil, err
	}
	if err2 != nil {
		return nil, err2
	}

	return &t, nil
}

// Upsert takes a task and write it to the underlying fs
func (s *FSStore) Upsert(t *task.Task) error {
	// construct the tree on the FS
	err := os.MkdirAll(s.GetVolumePath(t), DirMode)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(s.taskFilePath(t), os.O_CREATE+os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	err = json.NewEncoder(f).Encode(t)
	err2 := f.Close()
	if err != nil {
		return err
	}
	return err2
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

// GetByCommit gets the task using the full path from service to commit
func (s *FSStore) GetByCommit(service, project, branch, commit string, latest bool) (*task.Task, error) {

	// if latest return early
	if latest {
		latest, err := s.GetLatest(service, project, branch)
		if err != nil {
			return nil, err
		}
		return latest, nil
	}

	// if not latest, do the heavy stuff
	var t *task.Task
	basePath := filepath.Join(s.root, service, project, branch)
	dirs, err := os.ReadDir(basePath)
	if err != nil {
		return nil, err
	}

	for _, dir := range dirs {
		t, err = s.Get(dir.Name())
		if err != nil {
			continue
		}

		if t.Commit == commit {
			return t, nil
		}
	}

	return nil, fmt.Errorf("task with commit `%s` not found", commit)
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
	all, err := s.All()
	if err != nil {
		return nil, err
	}

	tasks := make([]*task.Task, 0)
	for _, t := range all {
		if filterFn(t) {
			tasks = append(tasks, t)
		}
	}

	return tasks, nil
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
	return os.WriteFile(s.taskLatestPath(t), []byte(t.Id.String()), 0600)
}

// GetLatest is used to get latest task for a service, project, branch
func (s *FSStore) GetLatest(service, project, branch string) (*task.Task, error) {
	pth := filepath.Clean(filepath.Join(s.root, service, project, branch, latestFile))
	if !strings.HasPrefix(pth, filepath.Join(s.root, service, project, branch)) {
		panic("path escape: " + pth)
	}
	content, err := os.ReadFile(pth)
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

// Prune will delete run directory older than provided date
func (s *FSStore) Prune(duration time.Duration, dry bool) (int64, error) {
	limit := time.Now().Add(-duration)
	if time.Now().Before(limit) {
		return 0, fmt.Errorf("computed limit time can't be in the future")
	}

	toPrune, err := s.Filter(func(t *task.Task) bool {
		return t.Creation.Before(limit)
	})
	if err != nil {
		return 0, err
	}

	workers := 5
	jobs := make(chan string, workers)
	results := make(chan int64, len(toPrune))

	// start workers
	pruneWorker(jobs, results, workers, dry)

	for _, t := range toPrune {
		jobs <- s.taskRootPath(t)
	}

	close(jobs)

	size := int64(0)
	for i := 0; i < len(toPrune); i++ {
		size += <-results
	}

	close(results)

	return size, nil
}

func pruneWorker(jobs <-chan string, results chan<- int64, workers int, dry bool) {
	for i := 0; i < workers; i++ {
		go func() {
			for p := range jobs {
				// count the size in bytes
				var size int64
				err := filepath.Walk(p, func(_ string, info os.FileInfo, err error) error {
					if !info.IsDir() {
						size += info.Size()
					}

					return err
				})

				// TODO: better logging
				if err != nil {
					fmt.Printf("error in prune worker : %v\n", err)
					return
				}

				// send results of sizes to channel
				results <- size

				// if not dry, remove dir
				if !dry {
					err := os.RemoveAll(p)
					// TODO: better logging
					if err != nil {
						fmt.Printf("error in prune worker : %v\n", err)
						return
					}
				}
			}
		}()
	}
}
