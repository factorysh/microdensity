package application

import (
	"io/ioutil"
	"net/http"
	"os"
	"path"

	"github.com/go-chi/chi/v5"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"go.uber.org/zap"
)

func (a *Application) ReadmeHandler(w http.ResponseWriter, r *http.Request) {
	serviceId := chi.URLParam(r, "serviceID")
	md := path.Join(a.serviceFolder, serviceId, "README.md")
	l := a.logger.With(zap.String("path", md))
	_, err := os.Stat(md)
	if err != nil {
		if os.IsNotExist(err) {
			l.Warn("README.md not found", zap.Error(err))
			w.WriteHeader(http.StatusNotFound)
			return
		}
		l.Error("README.md stat error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	raw, err := ioutil.ReadFile(md)
	if err != nil {
		l.Error("README.md read error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
	}

	_md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM, // Github favored markup
			extension.Typographer,
		),
	)
	err = _md.Convert(raw, w)
	if err != nil {
		l.Error("README.md markdown error", zap.Error(err))
	}
}
