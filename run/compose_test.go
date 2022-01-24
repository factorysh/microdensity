package run

import (
	"bytes"
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
	if os.Getenv("CI") != "" {
		t.Skip("Skipping testing in CI environment")
	}

	cr, err := NewComposeRun("../demo/")
	assert.NoError(t, err)
	buff := &bytes.Buffer{}

	err = cr.Prepare(map[string]string{})
	assert.NoError(t, err)
	rcode, err := cr.Run(&MockupReaderCloser{buff}, os.Stderr)
	assert.NoError(t, err)
	assert.Equal(t, 0, rcode)
	out, err := ioutil.ReadAll(buff)
	assert.NoError(t, err)
	fmt.Println(string(out))
	assert.Equal(t, "World", strings.TrimSpace(string(out)))

	buff.Reset()
	err = cr.Prepare(map[string]string{
		"HELLO": "Bob",
	})
	assert.NoError(t, err)
	rcode, err = cr.Run(&MockupReaderCloser{buff}, os.Stderr)
	assert.NoError(t, err)
	assert.Equal(t, 0, rcode)
	out, err = ioutil.ReadAll(buff)
	assert.NoError(t, err)
	fmt.Println(string(out))
	assert.Equal(t, "Bob", strings.TrimSpace(string(out)))

}
