package application

import (
	"fmt"
	"net/http"

	"github.com/factorysh/microdensity/version"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "text/plain")
	w.Write([]byte(`
   _               _                                     _
___| |__   ___  ___| | __  _ __ ___  _   _  __      _____| |__
/ __| '_ \ / _ \/ __| |/ / | '_ ' _ \| | | | \ \ /\ / / _ \ '_ \
| (__| | | |  __/ (__|   <  | | | | | | |_| |  \ V  V /  __/ |_) |
\___|_| |_|\___|\___|_|\_\ |_| |_| |_|\__, |   \_/\_/ \___|_.__/
									 |___/
	`))
	fmt.Fprintf(w, "Version: %s", version.Version())
}
