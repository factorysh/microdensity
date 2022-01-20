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
	rcode, err := cr.Run(ctxRun, map[string]string{}, &MockupReaderCloser{buff}, os.Stderr)
	assert.NoError(t, err)
	assert.Equal(t, 0, rcode)
	out, err := ioutil.ReadAll(buff)
	assert.NoError(t, err)
	fmt.Println(string(out))
	assert.Equal(t, "World", strings.TrimSpace(string(out)))

	buff.Reset()
	ctxRun = context.TODO()
	rcode, err = cr.Run(ctxRun, map[string]string{
		"HELLO": "Bob",
	}, &MockupReaderCloser{buff}, os.Stderr)
	assert.NoError(t, err)
	assert.Equal(t, 0, rcode)
	out, err = ioutil.ReadAll(buff)
	assert.NoError(t, err)
	fmt.Println(string(out))
	assert.Equal(t, "Bob", strings.TrimSpace(string(out)))

}
