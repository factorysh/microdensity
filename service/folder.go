package service

import (
	"fmt"
	"os"
	"path"

	"github.com/factorysh/microdensity/queue"
	"github.com/factorysh/microdensity/task"
	"github.com/google/uuid"
)

var _ Service = (*FolderService)(nil)

type FolderService struct {
	name string
	qeue *queue.Queue
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
	return &FolderService{
		name: name,
	}, nil
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
