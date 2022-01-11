package application

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/factorysh/microdensity/service"
	"github.com/stretchr/testify/assert"
)

func TestApplication(t *testing.T) {
	a, err := New(nil)
	assert.NoError(t, err)
	a.Services = append(a.Services, &service.Service{
		Name: "demo",
	})

	ts := httptest.NewServer(a.router)
	defer ts.Close()

	r, err := http.Get(fmt.Sprintf("%s/services", ts.URL))
	assert.NoError(t, err)
	dec := json.NewDecoder(r.Body)
	var services []string
	err = dec.Decode(&services)
	assert.NoError(t, err)
	assert.Equal(t, []string{"demo"}, services)

	r, err = http.Get(fmt.Sprintf("%s/service/demo", ts.URL))
	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)
}
