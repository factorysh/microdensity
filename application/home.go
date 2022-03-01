package application

import (
	"embed"
	"fmt"
	"net/http"
	"strings"

	"github.com/factorysh/microdensity/html"
	"github.com/factorysh/microdensity/version"
	"go.uber.org/zap"
)

var (
	// used to ensure embed import
	_ embed.FS
	//go:embed templates/home.html
	homeTemplate string
)

func acceptsHTML(r *http.Request) bool {
	accepts, found := r.Header["Accept"]
	if !found {
		return false
	}

	for _, h := range accepts {
		if strings.Contains(h, "text/html") {
			return true
		}
	}

	return false
}

const logo = `
      _               _                                     _
  ___| |__   ___  ___| | __  _ __ ___  _   _  __      _____| |__
/  __| '_ \ / _ \/ __| |/ / | '_ ' _ \| | | | \ \ /\ / / _ \ '_ \
| (__| | | |  __/ (__|   <  | | | | | | |_| |  \ V  V /  __/ |_) |
\ ___|_| |_|\___|\___|_|\_\ |_| |_| |_|\__, |   \_/\_/ \___|_.__/
                                       |___/
	`

// HomeHandler display the home page
func (a *Application) HomeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := a.logger.With(
			zap.String("path", "home"),
		)

		// if you ask for something that is not html
		// nice geeky ascii art
		if !acceptsHTML(r) {
			w.Header().Set("content-type", "text/plain")
			w.Write([]byte(logo))
			fmt.Fprintf(w, "Version: %s", version.Version())
			return
		}

		p := html.Page{
			Detail: "Home",
			Domain: a.Domain,
			Partial: html.Partial{
				Data:     a.Services,
				Template: homeTemplate,
			},
		}
		err := p.Render(w)
		if err != nil {
			l.Error("html render error", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func AdminHomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "text/plain")
	w.Write([]byte(logo))
	fmt.Fprintf(w, "Version: %s", version.Version())
	w.Write([]byte(`
/metrics Prometheus export
`))
}
