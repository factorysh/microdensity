package application

import (
	"io/ioutil"
	"net/http"
	"os"
	"path"

	"github.com/yuin/goldmark"
	"go.uber.org/zap"
)

func (a *Application) ReadmeHandler(w http.ResponseWriter, r *http.Request) {
	md := path.Join(a.serviceFolder, "README.md")
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

	err = goldmark.Convert(raw, w)
	if err != nil {
		l.Error("README.md markdown error", zap.Error(err))
	}
}
