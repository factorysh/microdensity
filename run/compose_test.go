package run

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockupReaderCloser struct {
	io.Writer
}

func (m *MockupReaderCloser) Close() error {
	return nil
}

func TestCompose(t *testing.T) {
	cr, err := NewComposeRun("../demo/")
	assert.NoError(t, err)
	buff := &bytes.Buffer{}

	ctxRun := context.TODO()
	defer ctxRun.Done()
	err = cr.Run(ctxRun, map[string]string{}, &MockupReaderCloser{buff}, os.Stderr)
	assert.NoError(t, err)
	out, err := ioutil.ReadAll(buff)
	assert.NoError(t, err)
	fmt.Println(string(out))
	assert.Equal(t, "World", strings.TrimSpace(string(out)))

	/*
		buff.Reset()
		ctxRun = context.TODO()
		cr.Run(ctxRun, map[string]string{
			"HELLO": "Bob",
		})
		out, err = ioutil.ReadAll(buff)
		assert.NoError(t, err)
		fmt.Println(string(out))
		assert.Equal(t, "Bob", strings.TrimSpace(string(out)))
	*/
}
