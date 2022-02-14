package application

import (
	"io"
	"net/http"

	"github.com/factorysh/microdensity/assets"
)

func FaviconHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/png")
	f, err := assets.F.Open("microdensity.png")
	if err != nil {
		panic(err)
	}
	io.Copy(w, f)
}
