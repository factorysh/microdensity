package gojatest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGoja(t *testing.T) {
	tt, err := New("../demo/services/demo/meta.js", "../demo/services/demo/meta_test.js")
	assert.NoError(t, err)
	tt.RunAll()
	//assert.True(t, false)
}
