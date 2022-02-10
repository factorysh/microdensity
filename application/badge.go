package application

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/factorysh/microdensity/badge"
	_badge "github.com/factorysh/microdensity/badge"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// BadgeMyTaskHandler generates a badge for a given task
func (a *Application) BadgeMyTaskHandler(latest bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		// get the task
		t, err := a.storage.GetByCommit(service, project, branch, commit, latest)
		if err != nil {
			l.Warn("Task get error", zap.Error(err))
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// try to get the output badge for this task in this service
		p := filepath.Join(a.storage.GetVolumePath(t), "/data", fmt.Sprintf("%s.badge", bdg))
		_, err = os.Stat(p)

		// if not found
		if err != nil {
			// fallback to status badge
			if os.IsNotExist(err) {
				// use the service name, task status and colors from badge pkg
				badge.WriteBadge(service, t.State.String(), _badge.Colors.Get(t.State), w)
				return
			}
			w.WriteHeader(http.StatusBadRequest)
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
}
