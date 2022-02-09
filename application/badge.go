package application

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	_badge "github.com/factorysh/microdensity/badge"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func (a *Application) TaskMyBadgeHandler(w http.ResponseWriter, r *http.Request) {
	l := a.logger.With(
		zap.String("url", r.URL.String()),
		zap.String("service", chi.URLParam(r, "serviceID")),
		zap.String("project", chi.URLParam(r, "project")),
		zap.String("branch", chi.URLParam(r, "branch")),
		zap.String("commit", chi.URLParam(r, "commit")),
		zap.String("branch", chi.URLParam(r, "branch")),
	)
	service := chi.URLParam(r, "serviceID")
	project := chi.URLParam(r, "project")
	branch := chi.URLParam(r, "branch")
	commit := chi.URLParam(r, "commit")
	bdg := chi.URLParam(r, "badge")

	t, err := a.storage.GetByCommit(service, project, branch, commit, false)
	if err != nil {
		l.Warn("Task get error", zap.Error(err))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	p := filepath.Join(a.storage.GetVolumePath(t), "/data", fmt.Sprintf("%s.badge", bdg))

	_, err = os.Stat(p)
	if err != nil {
		l.Warn("Task get error", zap.Error(err))
		if os.IsNotExist(err) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
		return
	}
	l = l.With(zap.String("path", p))
	b, err := ioutil.ReadFile(p)
	if err != nil {
		l.Error("reading file", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var badge _badge.Badge
	err = json.Unmarshal(b, &badge)
	if err != nil {
		l.Error("JSON unmarshal", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	badge.Render(w, r)
}
