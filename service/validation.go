package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/compose-spec/compose-go/types"
	"github.com/factorysh/microdensity/run"
)

// ValidateServicesDefinitions takes a root dir of services and inspect all services compose files
func ValidateServicesDefinitions(servicesDir string) error {
	dirs, err := os.ReadDir(servicesDir)
	if err != nil {
		return err
	}

	for _, dir := range dirs {
		err := validateServiceDefinition(filepath.Join(servicesDir, dir.Name()))
		if err != nil {
			return fmt.Errorf("error when reading service subdir %s: %v", dir.Name(), err)
		}
	}

	return nil
}

func validateServiceDefinition(path string) error {
	p, _, err := run.LoadCompose(path, map[string]string{})
	if err != nil {
		return err
	}

	for _, fn := range []validatorFunc{volumesValidator} {
		err := fn(p)
		if err != nil {
			return fmt.Errorf("error when validating docker-compose.yml file in directory %s: %v", path, err)
		}

	}

	return nil
}

const volumeMaxDeep = 15

func volumesValidator(p *types.Project) error {
	for _, svc := range p.Services {
		for _, vol := range svc.Volumes {
			if vol.Type != "bind" {
				return fmt.Errorf("found mount of type %s in service %s", vol.Type, svc.Name)
			}

			if !strings.HasPrefix(vol.Source, "./") {
				return fmt.Errorf("found a none relative mount %s in service %s", vol.Source, svc.Name)
			}

			if strings.Contains(vol.Source, "..") {
				return fmt.Errorf("found a path trying to access a parent directory %s in service %s", vol.Source, svc.Name)
			}

			if len(strings.Split(vol.Source, "/")) > volumeMaxDeep {
				return fmt.Errorf("path is too %s is too deep (> %d)", vol.Source, volumeMaxDeep)
			}
		}

	}

	return nil
}

type validatorFunc func(*types.Project) error
