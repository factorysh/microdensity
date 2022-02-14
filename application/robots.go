package application

import "net/http"

func RobotsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(`
User-agent: *
Disallow:
`))
}
