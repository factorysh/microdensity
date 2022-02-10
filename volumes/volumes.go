package volumes

import (
	"io/fs"
	"os"
	"path/filepath"

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

// Path will return the full path for this volume
func (v *Volumes) Path(elem ...string) string {
	return filepath.Join(v.root, filepath.Join(elem...))
}
