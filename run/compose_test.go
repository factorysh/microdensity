package run

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompose(t *testing.T) {
	cr, err := NewComposeRun("../demo/")
	assert.NoError(t, err)
	fmt.Println(cr)
}
