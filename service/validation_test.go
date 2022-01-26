package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateServiceDefiniton(t *testing.T) {
	t.Run("valid definition", func(t *testing.T) {
		err := validateServiceDefinition("../fixtures/services/valids/test")
		assert.NoError(t, err)
	})

	t.Run("invalid definition", func(t *testing.T) {
		tests := []struct {
			name       string
			dir        string
			errMessage string
		}{
			{name: "access parent directory", dir: "../fixtures/services/invalids/volumes-parent", errMessage: "error when validating docker-compose.yml file in directory ../fixtures/services/invalids/volumes-parent: found a path trying to access a parent directory ./../cache in service hello"},
			{name: "absolute path", dir: "../fixtures/services/invalids/absolute-path", errMessage: "error when validating docker-compose.yml file in directory ../fixtures/services/invalids/absolute-path: found a none relative mount /cache in service hello"},
		}
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				err := validateServiceDefinition(tc.dir)
				assert.EqualError(t, err, tc.errMessage)
			})
		}
	})

}
