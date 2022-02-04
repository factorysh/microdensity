package volumes

import (
	"encoding/json"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"github.com/factorysh/microdensity/task"
	"go.uber.org/zap"
)

// DirMode is the default dirmode for volumes
const DirMode fs.FileMode = 0755

// New inits a new volumes struct
func New(root string) (*Volumes, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	err = os.MkdirAll(root, DirMode)
	if err != nil {
		logger.Error("Volume mkdir error",
			zap.String("root", root),
			zap.Error(err))
		return nil, err
	}

	return &Volumes{
		root:   root,
		logger: logger,
	}, nil
}

// Volumes struct used to handle volumes CRUD
type Volumes struct {
	root   string
	logger *zap.Logger
}

// Create a new volume
func (v *Volumes) Create(t *task.Task) error {
	l := v.logger.With(
		zap.String("root", v.root),
		zap.Any("task", t),
	)
	err := t.Validate()

	if err != nil {
		l.Warn("Validation failed", zap.Error(err))
		return err
	}
	err = os.MkdirAll(v.Path(t.Service, t.Project, t.Branch, t.Id.String(), "volumes"), DirMode)
	if err != nil {
		l.Error("mkdir all error", zap.Error(err))
	}
	return err
}

func (v *Volumes) Get(service, project, branch, commit string) (*task.Task, error) {
	p := v.Path(service, project, branch)
	l := v.logger.With(
		zap.String("root", v.root),
		zap.String("service", service),
		zap.String("project", project),
		zap.String("branch", branch),
		zap.String("commit", commit),
		zap.String("path", p),
	)
	_, err := os.Stat(p)
	if err != nil {
		if os.IsNotExist(err) {
			l.Warn("Volume not found")
		} else {
			l.Error("Stat error", zap.Error(err))
		}
		return nil, err
	}

	commits, err := ioutil.ReadDir(p)
	if err != nil {
		l.Error("Readall error", zap.Error(err))
		return nil, err
	}
	for _, com := range commits {
		l = l.With(zap.String("commit", com.Name()))
		f, err := os.OpenFile(v.Path(service, project, branch, com.Name(), "task.json"), os.O_RDONLY, 0)
		if err != nil {
			l.Error("open commit file", zap.Error(err))
			return nil, err
		}
		var t task.Task
		err = json.NewDecoder(f).Decode(&t)
		if err != nil {
			l.Error("JSON decode", zap.Error(err))
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
		v.logger.Error("by project by branch",
			zap.String("project", project),
			zap.String("brnach", branch),
			zap.Error(err),
		)
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
