package service

import (
	"fmt"
	"os"
	"path"
)

type FolderService struct {
	name string
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
func (f *FolderService) Run(args map[string]interface{}) error {
	return nil
}
