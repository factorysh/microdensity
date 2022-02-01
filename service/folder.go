package service

import (
	"fmt"
	"os"
	"path"

	"github.com/dop251/goja"
	"github.com/factorysh/microdensity/queue"
	"github.com/factorysh/microdensity/task"
	"github.com/google/uuid"
)

var _ Service = (*FolderService)(nil)

type FolderService struct {
	name      string
	qeue      *queue.Storage
	jsruntime *goja.Runtime
	validate  func(map[string]interface{}) map[string]interface{}
}

func NewFolder(_path string) (*FolderService, error) {
	stat, err := os.Stat(_path)
	if err != nil {
		return nil, err
	}
	if !stat.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", _path)
	}

	_, name := path.Split(_path)
	service := &FolderService{
		name: name,
	}

	jsPath := path.Join(_path, "meta.js")
	_, err = os.Stat(jsPath)
	if err != nil {
		if os.IsNotExist(err) {
			// there is meta.js file
		} else {
			return nil, err
		}
	} else {
		vm := goja.New()
		_, err := vm.RunScript("meta", jsPath)
		if err != nil {
			return nil, err
		}
		err = vm.ExportTo(vm.Get("validate"), &service.validate)
		if err != nil {
			return nil, err
		}
		service.jsruntime = vm
	}

	return service, nil
}

func (f *FolderService) Name() string {
	return f.name
}
func (f *FolderService) Validate(map[string]interface{}) error {
	return nil
}
func (f *FolderService) New(project string, args map[string]interface{}) (uuid.UUID, error) {
	t := &task.Task{
		Id:      uuid.New(),
		Project: project,
		Args:    args,
		State:   task.Ready,
	}
	err := f.qeue.Put(t)
	return t.Id, err
}

func (f *FolderService) Run(id uuid.UUID) error {
	return nil
}
