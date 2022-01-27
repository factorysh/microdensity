package volumes

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/factorysh/microdensity/task"
)

// DirMode is the default dirmode for volumes
const DirMode fs.FileMode = 0755

// New inits a new volumes struct
func New(root string) (*Volumes, error) {
	err := os.MkdirAll(root, DirMode)
	if err != nil {
		return nil, err
	}

	return &Volumes{
		root: root,
	}, nil
}

// Volumes struct used to handle volumes CRUD
type Volumes struct {
	root string
}

// Create a new volume
func (v *Volumes) Create(t *task.Task) error {
	if strings.ContainsRune(t.Project, '/') {
		return fmt.Errorf("project name must be url escaped, without any / : %s", t.Project)
	}
	if strings.ContainsRune(t.Branch, '/') {
		return fmt.Errorf("branch name must be url escaped, without any / : %s", t.Branch)
	}
	err := os.MkdirAll(v.Path(t.Service, t.Project, t.Branch, t.Id.String(), "volumes"), DirMode)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(v.Path(t.Service, t.Project, t.Branch, t.Id.String(), "task.json"), os.O_CREATE+os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	return json.NewEncoder(f).Encode(t)
}

func (v *Volumes) Get(service, project, branch, commit string) (*task.Task, error) {
	p := v.Path(service, project, branch)
	commits, err := ioutil.ReadDir(p)
	if err != nil {
		fmt.Println("path", p)
		return nil, err
	}
	for _, com := range commits {
		f, err := os.OpenFile(v.Path(service, project, branch, com.Name(), "task.json"), os.O_RDONLY, 0)
		if err != nil {
			return nil, err
		}
		var t task.Task
		err = json.NewDecoder(f).Decode(&t)
		if err != nil {
			return nil, err
		}
		if t.Commit == commit {
			return &t, nil
		}
	}
	return nil, nil
}

// Path will return the full path for this volume
func (v *Volumes) Path(elem ...string) string {
	return filepath.Join(v.root, filepath.Join(elem...))
}

// ByProjectByBranch returns all subvolumes for a specific branch of a project
func (v *Volumes) ByProjectByBranch(project string, branch string) ([]string, error) {
	vols, err := ioutil.ReadDir(v.Path(project, branch))
	if err != nil || !containsFiles(vols) {
		return nil, err
	}

	var res []string
	for _, run := range vols {
		if run.IsDir() {
			res = append(res, v.Path(project, branch, run.Name()))
		}
	}
	sort.Strings(res)

	return res, nil
}

func containsFiles(files []fs.FileInfo) bool {
	return len(files) > 0
}
