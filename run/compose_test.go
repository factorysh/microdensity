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
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
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

	cr, err := NewComposeRun("../demo/services/demo")
	assert.NoError(t, err)
	buff := &bytes.Buffer{}

	err = cr.Prepare(map[string]string{},
		microdensityVolumesRoot,
		uuid.New(),
		[]string{})
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
	}, microdensityVolumesRoot, uuid.New(), []string{})
	assert.NoError(t, err)
	rcode, err = cr.Run(&MockupReaderCloser{buff}, os.Stderr)
	assert.NoError(t, err)
	assert.Equal(t, 0, rcode)
	out, err = ioutil.ReadAll(buff)
	assert.NoError(t, err)
	fmt.Println(string(out))
	assert.Equal(t, "Bob", strings.TrimSpace(string(out)))

	buff.Reset()
	err = cr.Prepare(map[string]string{
		"HELLO": "Bob",
	}, microdensityVolumesRoot,
		uuid.New(),
		[]string{"google.dns:8.8.8.8"})
	assert.NoError(t, err)
	for _, s := range cr.project.Services {
		fmt.Println("hosts", s.ExtraHosts)
	}
	rcode, err = cr.runCommand(&MockupReaderCloser{buff}, os.Stderr,
		[]string{"grep", "google.dns", "/etc/hosts"})
	assert.NoError(t, err)
	assert.Equal(t, 0, rcode)
	out, err = ioutil.ReadAll(buff)
	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(string(out), "8.8.8.8"))

	rc, err := cr.Logs(context.TODO())
	assert.NoError(t, err)
	logs, err := ioutil.ReadAll(rc)
	assert.NoError(t, err)
	assert.Contains(t, string(logs), "8.8.8.8\tgoogle.dns")
}
