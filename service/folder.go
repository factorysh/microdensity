package service

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/dop251/goja"
	"github.com/factorysh/microdensity/task"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var _ Service = (*FolderService)(nil)

type FolderService struct {
	name      string
	jsruntime *goja.Runtime
	validate  func(map[string]interface{}) (Arguments, error)
	logger    *zap.Logger
}

type Arguments struct {
	Environments map[string]string `json:"environments"`
	Files        map[string]string `json:"files"`
}

func NewFolder(_path string) (*FolderService, error) {
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

	_, name := path.Split(_path)
	service := &FolderService{
		name:   name,
		logger: logger,
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
		vm.Set("debug", func(a interface{}) {
			spew.Dump(a)
		})
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
		l.Info("js is ready", zap.Float64("js cooking time (ms)", float64(time.Since(chrono))/1000000))
		service.jsruntime = vm
	}

	return service, nil
}

func (f *FolderService) Name() string {
	return f.name
}
func (f *FolderService) Validate(args map[string]interface{}) (Arguments, error) {
	chrono := time.Now()
	defer f.logger.Info("Validate", zap.String("name", f.name), zap.Float64("validation time (µs)", float64(time.Since(chrono))/1000))
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

func (f *FolderService) Run(id uuid.UUID) error {
	return nil
}
