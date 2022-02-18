package gojatest

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGoja(t *testing.T) {
	vm, err := ReadAll("../demo/services/demo/meta.js", "../demo/services/demo/meta_test.js")
	assert.NoError(t, err)
	fmt.Println(vm.GlobalObject().Keys())
	assert.True(t, false)
}
