package application

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/factorysh/microdensity/service"
	"github.com/stretchr/testify/assert"
)

type rc struct {
	*bytes.Buffer
}

func (r *rc) Close() error {
	return nil
}

func TestCreateTask(t *testing.T) {

	svc, err := service.NewFolder("../demo")
	assert.NoError(t, err)

	var services = map[string]service.Service{
		"demo": svc,
	}

	key, app, _, q, cleanUp := prepareTestingContext(t, services)
	defer cleanUp()

	cli := http.Client{}
	req, err := mkRequest(key)
	assert.NoError(t, err)
	req.Method = "POST"
	req.URL, err = url.Parse(fmt.Sprintf("%s/service/demo/group%%2Fproject/main/8e54b1d8c5f0859370196733feeb00da022adeb5", app.URL))
	assert.NoError(t, err)
	b := &rc{
		&bytes.Buffer{},
	}
	err = json.NewEncoder(b).Encode(map[string]interface{}{
		"HELLO": "Bob",
	})
	assert.NoError(t, err)

	req.Body = b
	r, err := cli.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)

	l, err := q.Length()
	assert.NoError(t, err)
	assert.Equal(t, 1, l)

	req, err = mkRequest(key)
	assert.NoError(t, err)
	req.Method = "GET"
	req.URL, err = url.Parse(fmt.Sprintf("%s/service/demo/group%%2Fproject/main/8e54b1d8c5f0859370196733feeb00da022adeb5", app.URL))
	assert.NoError(t, err)
	r, err = cli.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)
}
