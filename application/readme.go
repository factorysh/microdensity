package application

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/factorysh/microdensity/html"
	"github.com/go-chi/chi/v5"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	_html "github.com/yuin/goldmark/renderer/html"
	"go.uber.org/zap"
)

func (a *Application) ReadmeHandler(w http.ResponseWriter, r *http.Request) {
	serviceID := chi.URLParam(r, "serviceID")
	md := path.Join(a.serviceFolder, serviceID, "README.md")
	l := a.logger.With(zap.String("path", md))
	if strings.Contains(md, "..") {
		l.Error("Path with ..", zap.String("path", md))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
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

	raw, err := ioutil.ReadFile(md) //#nosec path assertion at the begining of the function
	if err != nil {
		l.Error("README.md read error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
	}

	_md := goldmark.New(
		goldmark.WithRendererOptions(
			_html.WithXHTML(),
			_html.WithWriter(_html.DefaultWriter),
		),
		goldmark.WithExtensions(
			extension.GFM, // Github favored markup
			extension.Typographer,
		))

	var buffer bytes.Buffer
	err = _md.Convert(raw, &buffer)
	if err != nil {
		l.Error("README.md markdown error", zap.Error(err))
	}

	p := html.Page{
		Detail: serviceID,
		Domain: a.Domain,
		Partial: html.Partial{
			Template: buffer.String(),
		},
	}
	err = p.Render(w)
	if err != nil {
		l.Error("html render error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

}
