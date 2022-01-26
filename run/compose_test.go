package run

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/factorysh/microdensity/volumes"
	"github.com/stretchr/testify/assert"
)

type MockupReaderCloser struct {
	io.Writer
}

func (m *MockupReaderCloser) Close() error {
	return nil
}

const microdensityVolumesRoot = "/tmp/microdensity/volumes/uuid"

func TestCompose(t *testing.T) {
	os.MkdirAll(microdensityVolumesRoot, volumes.DirMode)
	defer func() {
		os.RemoveAll(microdensityVolumesRoot)
	}()

	if os.Getenv("CI") != "" {
		t.Skip("Skipping testing in CI environment")
	}

	cr, err := NewComposeRun("../demo/")
	assert.NoError(t, err)
	buff := &bytes.Buffer{}

	err = cr.Prepare(map[string]string{}, microdensityVolumesRoot)
	assert.NoError(t, err)
	rcode, err := cr.Run(&MockupReaderCloser{buff}, os.Stderr)
	assert.NoError(t, err)
	assert.Equal(t, 0, rcode)
	out, err := ioutil.ReadAll(buff)
	assert.NoError(t, err)
	fmt.Println(string(out))
	assert.Equal(t, "World", strings.TrimSpace(string(out)))
	// check for volumes dirs
	dirs, err := os.ReadDir(fmt.Sprintf("%s/volumes/cache", microdensityVolumesRoot))
	assert.NoError(t, err)
	assert.Len(t, dirs, 1, "expected 1 file in cache dir")

	buff.Reset()
	err = cr.Prepare(map[string]string{
		"HELLO": "Bob",
	}, microdensityVolumesRoot)
	assert.NoError(t, err)
	rcode, err = cr.Run(&MockupReaderCloser{buff}, os.Stderr)
	assert.NoError(t, err)
	assert.Equal(t, 0, rcode)
	out, err = ioutil.ReadAll(buff)
	assert.NoError(t, err)
	fmt.Println(string(out))
	assert.Equal(t, "Bob", strings.TrimSpace(string(out)))

}
