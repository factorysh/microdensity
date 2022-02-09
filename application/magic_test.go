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
		{name: "no magic url /metrics", path: "/metrics", want: "/metrics"},
		{name: "no magic url /services", path: "/services", want: "/services"},
		{name: "no magic url with service name", path: "/service/demo", want: "/service/demo"},
		{name: "no magic url with service name", path: "/service/demo", want: "/service/demo"},
		{name: "no magic url with project commit", path: "/service/demo/group%2Fproject/master/commit/status", want: "/service/demo/group%2Fproject/master/commit/status"},
		{name: "magic url with project commit", path: "/service/demo/group/project/-/master/commit/status", want: "/service/demo/group%2Fproject/master/commit/status"},
		{name: "no magic url with project latest", path: "/service/demo/group%2Fproject/master/latest", want: "/service/demo/group%2Fproject/master/latest"},
		{name: "magic url with project latest", path: "/service/demo/group/project/-/master/latest", want: "/service/demo/group%2Fproject/master/latest"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, pathMagic(tc.path, 2))
		})
	}
}

func TestMagicPathHandler(t *testing.T) {
	u, err := url.Parse("https://example.com/robots.txt")
	assert.NoError(t, err)
	assert.Equal(t, "/robots.txt", u.Path) // ok, the path starts with a /

	var snitch *url.URL
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		snitch = r.URL
	})
	srv := httptest.NewServer(MagicPathHandler(h))
	defer srv.Close()

	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "full path",
			path: "service/demo/group/subgroup/project/-/master/latest",
			want: "/service/demo/group%2Fsubgroup%2Fproject/master/latest",
		},
		{
			name: "not a service",
			path: "metrics",
			want: "/metrics",
		},
		{
			name: "with a volume",
			path: "service/demo/factory/check-my-demo/-/master/bf3dfa8fde041eda86e873f5251a6d49158ba5b3/volumes/cache/cache/proof",
			want: "/service/demo/factory%2Fcheck-my-demo/master/bf3dfa8fde041eda86e873f5251a6d49158ba5b3/volumes/cache/cache/proof",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res, err := http.Get(fmt.Sprintf("%s/%s", srv.URL, tc.path))
			assert.NoError(t, err)
			assert.Equal(t, 200, res.StatusCode)
			assert.Equal(t, tc.want, snitch.Path, snitch)
		})
	}
}
