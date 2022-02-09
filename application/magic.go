package application

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// handles Gitlab like URL Paths and translate it to a Microdensity URL
func pathMagic(p string, baseLen int) string {
	if !strings.HasPrefix(p, "/service/") { // early exit, magic happens only on /service/*
		return p
	}
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

	return fmt.Sprintf("%s/%s/%s",
		strings.Join(baseWithProject[:baseLen+1], "/"),
		url.PathEscape(strings.Join(baseWithProject[baseLen+1:], "/")),
		parts[1],
	)
}

func MagicPathHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = pathMagic(r.URL.Path, 2)
		next.ServeHTTP(w, r)
	})
}
