package application

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// VolumesHandler expose volumes of a task
func (a *Application) VolumesHandler(basePathLen int) http.HandlerFunc {
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

		t, err := a.storage.GetByCommit(service, project, branch, commit, false)
		if err != nil {
			l.Error("Get task", zap.Error(err))
			w.WriteHeader(http.StatusNotFound)
			return
		}

		filePath, err := extractPathFromURL(r.URL.Path, basePathLen)
		if err != nil {
			l.Error("Extrat path", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if strings.Contains(filePath, "..") {
			l.Warn("Dangerous path trying to access parent directory")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		fullPath := filepath.Join(a.storage.GetVolumePath(t), filePath)

		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			l.Warn("Path not found", zap.Error(err))
			w.WriteHeader(http.StatusNotFound)
			return
		}

		http.ServeFile(w, r, fullPath)
	}
}

func extractPathFromURL(p string, basePathLen int) (string, error) {
	urlPath, file := path.Split(strings.TrimLeft(p, "/"))

	parts := strings.Split(urlPath, "/")
	if len(parts) < basePathLen {
		return "", fmt.Errorf("can not split path `%s` in more than %d elements", p, basePathLen)
	}

	folderPath := parts[basePathLen:]
	folderPath = append(folderPath, file)

	return path.Join(folderPath...), nil
}
