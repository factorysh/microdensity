package run

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/docker/compose/v2/cmd/formatter"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/stretchr/testify/assert"
)

func TestCompose(t *testing.T) {
	cr, err := NewComposeRun("../demo/")
	assert.NoError(t, err)
	ctx := context.TODO()
	buff := &bytes.Buffer{}
	consumer := formatter.NewLogConsumer(ctx, buff, false, false)
	go cr.service.Logs(context.TODO(), "demo", consumer, api.LogOptions{
		Follow: true,
	})

	ctxRun := context.TODO()
	defer ctxRun.Done()
	err = cr.Run(ctxRun, map[string]string{})
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
