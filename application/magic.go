package application

import (
	"net/http"
	"net/url"
	"path"
	"strings"
)

// handles Gitlab like URL Paths and translate it to a Microdensity URL
func pathMagic(p string, baseLen int) string {
	parts := strings.Split(p, "/-/")
	// if there is no magic delimiter
	// return the same path
	if len(parts) < 2 {
		return p
	}

	baseWithProject := strings.Split(parts[0], "/")
	if len(parts) < baseLen {
		return p
	}

	return path.Join(path.Join(baseWithProject[:baseLen]...), url.PathEscape(path.Join(baseWithProject[baseLen:]...)), parts[1])
}

func MagicPathHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		req.URL.Path = pathMagic(req.URL.Path, 2)
		next.ServeHTTP(w, req)
	})
}
