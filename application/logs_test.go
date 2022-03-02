package application

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDoubleLogger(t *testing.T) {
	var out bytes.Buffer
	dl := newDoubleLogger(&out)
	dl.Stderr.Write([]byte("2022-03-02T14:11:16.733625031Z an error log line"))
	dl.Stdout.Write([]byte("2022-03-02T14:13:16.733625031Z an stdout log line"))

	expexted := `<span class="stderr-prefix">2022-03-02T14:11:16.733625031Z</span> an error log line
<span class="stdout-prefix">2022-03-02T14:13:16.733625031Z</span> an stdout log line
`

	assert.Equal(t, expexted, out.String())
}
