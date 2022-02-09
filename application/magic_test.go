package application

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPathMagic(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{name: "no magic url /", path: "/", want: "/"},
		{name: "no magic url /metrics", path: "metrics", want: "metrics"},
		{name: "no magic url /services", path: "services", want: "services"},
		{name: "no magic url with service name", path: "service/demo", want: "service/demo"},
		{name: "no magic url with service name", path: "service/demo", want: "service/demo"},
		{name: "no magic url with project commit", path: "service/demo/group%2Fproject/master/commit/status", want: "service/demo/group%2Fproject/master/commit/status"},
		{name: "magic url with project commit", path: "service/demo/group/project/-/master/commit/status", want: "service/demo/group%2Fproject/master/commit/status"},
		{name: "no magic url with project latest", path: "service/demo/group%2Fproject/master/latest", want: "service/demo/group%2Fproject/master/latest"},
		{name: "magic url with project latest", path: "service/demo/group/project/-/master/latest", want: "service/demo/group%2Fproject/master/latest"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, pathMagic(tc.path, 2))
		})
	}
}

func TestMagicPathHandler(t *testing.T) {
	var snitch *url.URL
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		snitch = r.URL
	})
	srv := httptest.NewServer(MagicPathHandler(h))
	defer srv.Close()

	res, err := http.Get(fmt.Sprintf("%s/service/demo/group/subgroup/project/-/master/latest", srv.URL))
	assert.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, "service/demo%2Fgroup%2Fsubgroup%2Fproject/master/latest", snitch.Path)

}
