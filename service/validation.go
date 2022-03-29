package service

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/compose-spec/compose-go/types"
	"github.com/factorysh/microdensity/run"
	"gopkg.in/yaml.v3"
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
	err := validateImages(path)
	if err != nil {
		return err
	}

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

func validateImages(path string) error {
	file, err := os.Open(filepath.Clean(filepath.Join(path, "docker-compose.yml")))
	if err != nil {
		return err
	}

	data := map[string]interface{}{}
	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(data)
	if err != nil {
		return err
	}

	services, ok := data["services"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("unable to get services from docker-compose file %s", path)
	}

	for service, rawConfig := range services {
		config, ok := rawConfig.(map[string]interface{})
		if !ok {
			return fmt.Errorf("unable to get config for service %s from docker-compose file %s", service, path)
		}

		image, ok := config["image"].(string)
		if !ok {
			return fmt.Errorf("unable to get image for service %s from docker-compose file %s", service, path)
		}

		err = validateImage(image)
		if err != nil {
			return fmt.Errorf("error when valdiating image name for service %s from docker-compose file %s: %v", service, path, err)
		}

	}

	return nil
}

var variableRegex = regexp.MustCompile(`\${([a-zA-Z0-9_\-:]+)}`)
var variableDefaultRegex = regexp.MustCompile(`[a-zA-Z0-9_\-]+:-[a-zA-Z0-9_\-]+`)

func validateImage(name string) error {

	variables := variableRegex.FindAllSubmatch([]byte(name), -1)
	if variables == nil {
		return nil
	}

	for _, match := range variables {
		if !variableDefaultRegex.Match(match[1]) {
			return fmt.Errorf("missing a default variable in image name definition %s", name)
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
