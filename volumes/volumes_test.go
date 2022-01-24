package volumes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewVolumes(t *testing.T) {
	tests := []struct {
		name     string
		root     string
		validate func(*testing.T, error)
	}{
		{name: "Valid dir", root: "/tmp/microdensity/volumes", validate: func(t *testing.T, err error) {
			assert.NoError(t, err)
		}},
		{name: "Invalid dir", root: "/no", validate: func(t *testing.T, err error) {
			assert.Error(t, err)
		}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := New(tc.root)
			tc.validate(t, err)
		})
	}
}
