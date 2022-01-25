package volumes

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
)

const dirMode fs.FileMode = 0755

// New inits a new volumes struct
func New(root string) (*Volumes, error) {
	err := os.MkdirAll(root, dirMode)
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

// Request a new volume
func (v *Volumes) Request(project string, branch string, taskID string) error {
	return os.MkdirAll(v.Path(project, branch, taskID), dirMode)
}

// Path will return the full path for this volume
func (v *Volumes) Path(elem ...string) string {
	return filepath.Join(v.root, filepath.Join(elem...))
}

// ByProject all subvolumes for a project
func (v *Volumes) ByProject(project string) ([]string, error) {
	var res []string

	branches, err := ioutil.ReadDir(v.Path(project))
	if err != nil {
		return nil, err
	}

	fsSorter(branches)

	for _, branch := range branches {
		subs, err := v.ByProjectByBranch(project, branch.Name())
		if err != nil {
			fmt.Println(err)
			continue
		}

		res = append(res, subs...)
	}

	return res, nil
}

func fsSorter(files []fs.FileInfo) {
	sort.Slice(files, func(i int, j int) bool {
		return files[i].ModTime().Before(files[j].ModTime())
	})
}

// ByProjectByBranch returns all subvolumes for a specific branch of a project
func (v *Volumes) ByProjectByBranch(project string, branch string) ([]string, error) {

	var res []string

	vols, err := ioutil.ReadDir(v.Path(project, branch))
	if err != nil || !containsFiles(vols) {
		return nil, err
	}

	fsSorter(vols)

	for _, run := range vols {
		if run.IsDir() {
			res = append(res, v.Path(project, branch, run.Name()))
		}
	}

	return res, nil
}

func containsFiles(files []fs.FileInfo) bool {
	return len(files) > 0
}
