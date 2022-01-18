package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCue(t *testing.T) {
	f := &FolderService{}
	err := f.parseCue(`
	#args: {
	  url: string
	  url: =~"^https?://[\\w./]+$"
	}
	#args
`)
	assert.NoError(t, err)
	err = f.Validate(map[string]interface{}{
		"url": "http://factory.sh",
	})
	assert.NoError(t, err)
}
