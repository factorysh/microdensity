package volumes

import (
	"io/fs"
	"os"
	"path/filepath"
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
	return filepath.Join(elem...)
}
