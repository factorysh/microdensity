package gojatest

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGoja(t *testing.T) {
	defer Oups()
	tt, err := New("../demo/services/demo/meta.js", "../demo/services/demo/meta_test.js")
	assert.NoError(t, err)
	err = tt.RunAll()
	assert.NoError(t, err)
}

func Oups() {
	if r := recover(); r != nil {
		fmt.Println("Recovering from panic:", r)
	}
}
