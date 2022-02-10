package application

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/factorysh/microdensity/version"
	"github.com/yuin/goldmark"
	"go.uber.org/zap"
)

var (
	// used to ensure embed import
	_ embed.FS
	//go:embed templates/HELP.md
	helpTemplate string
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
			w.Write([]byte(`
      _               _                                     _
  ___| |__   ___  ___| | __  _ __ ___  _   _  __      _____| |__
/  __| '_ \ / _ \/ __| |/ / | '_ ' _ \| | | | \ \ /\ / / _ \ '_ \
| (__| | | |  __/ (__|   <  | | | | | | |_| |  \ V  V /  __/ |_) |
\ ___|_| |_|\___|\___|_|\_\ |_| |_| |_|\__, |   \_/\_/ \___|_.__/
                                       |___/
	`))
			fmt.Fprintf(w, "Version: %s", version.Version())
			return
		}

		// text/html requests sends a documentation file
		template, err := template.New("help").Parse(helpTemplate)
		if err != nil {
			l.Error("template", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var buffer bytes.Buffer

		template.Execute(&buffer, a.Services)

		err = goldmark.Convert(buffer.Bytes(), w)
		if err != nil {
			l.Error("markdown convert", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("content-type", "text/html")
	}
}
