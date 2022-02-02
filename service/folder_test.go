package service

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	service, err := NewFolder("../demo")
	assert.NoError(t, err)
	v, err := service.validate(map[string]interface{}{
		"HELLO": "Alice",
	})
	spew.Dump(v)
	assert.NoError(t, err)
	assert.Equal(t, "Alice", v.Environments["HELLO"], v)
	assert.Equal(t, "Hello Alice", v.Files["hello.txt"])
	_, err = service.Validate(map[string]interface{}{
		"HELLO": "Alice Dupont",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "HELLO is only letters", err.Error())
}
