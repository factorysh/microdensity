package service

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/dop251/goja"
	"github.com/factorysh/microdensity/task"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

var _ Service = (*FolderService)(nil)

type FolderService struct {
	name      string
	jsruntime *goja.Runtime
	validate  func(map[string]interface{}) (Arguments, error)
	badge     func(project, branch, commit, badge string) (Badge, error)
	logger    *zap.Logger
	meta      Meta
}

type Arguments struct {
	Environments map[string]string `json:"environments"`
	Files        map[string]string `json:"files"`
}

type Console struct {
}

func (c *Console) Log(args ...interface{}) {
	spew.Dump(args...)
}

func NewFolder(_path string) (*FolderService, error) {
	_path = path.Clean(_path)
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	l := logger.With(zap.String("path", _path))
	stat, err := os.Stat(_path)
	if err != nil {
		l.Error("stat", zap.Error(err))
		return nil, err
	}
	if !stat.IsDir() {
		l.Error("Is not a directory")
		return nil, fmt.Errorf("%s is not a directory", _path)
	}

	var content []byte

	// loop over possible meta.yml files
	for _, name := range []string{"meta.yml", "meta.yaml"} {
		content, err = os.ReadFile(filepath.Join(_path, name))
		if err == nil {
			break
		}
	}

	if err != nil {
		return nil, fmt.Errorf("error with path %s: %v", _path, err)
	}

	var m Meta

	err = yaml.Unmarshal(content, &m)
	if err != nil {
		return nil, fmt.Errorf("error with path %s: %v", _path, err)
	}

	_, name := path.Split(_path)
	service := &FolderService{
		name:   name,
		logger: logger,
		meta:   m,
	}
	l = l.With(zap.String("name", name))

	jsPath := path.Join(_path, "meta.js")
	_, err = os.Stat(jsPath)
	if err != nil {
		if os.IsNotExist(err) {
			// there is meta.js file
		} else {
			return nil, err
		}
	} else {
		chrono := time.Now()
		vm := goja.New()
		vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
		vm.Set("console", &Console{})
		src, err := ioutil.ReadFile(jsPath)
		if err != nil {
			return nil, err
		}
		_, err = vm.RunScript("meta", string(src))
		if err != nil {
			return nil, err
		}
		err = vm.ExportTo(vm.Get("validate"), &service.validate)
		if err != nil {
			return nil, err
		}
		l.Info("js is ready", zap.Float64("js cooking time (µs)", float64(time.Since(chrono))/1000))
		service.jsruntime = vm
	}

	return service, nil
}

func (f *FolderService) Name() string {
	return f.name
}
func (f *FolderService) Validate(args map[string]interface{}) (Arguments, error) {
	chrono := time.Now()
	defer f.logger.Info("Validate",
		zap.String("service", f.name),
		zap.Float64("validation time (µs)", float64(time.Since(chrono))/1000))
	return f.validate(args)
}
func (f *FolderService) New(project string, args map[string]interface{}) (uuid.UUID, error) {
	t := &task.Task{
		Id:      uuid.New(),
		Project: project,
		Args:    args,
		State:   task.Ready,
	}
	return t.Id, nil
}

func (f *FolderService) Badge(project, branch, commit, badge string) (Badge, error) {
	chrono := time.Now()
	defer f.logger.Info("Badge",
		zap.String("service", f.name),
		zap.String("project", project),
		zap.String("name", badge),
		zap.Float64("validation time (µs)", float64(time.Since(chrono))/1000))
	return f.badge(project, branch, commit, badge)
}

func (f *FolderService) Run(id uuid.UUID) error {
	// FIXME
	return nil
}

// Meta data about this service
func (f *FolderService) Meta() Meta {
	return f.meta
}
