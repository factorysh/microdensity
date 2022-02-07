package badge

import (
	"net/http"

	"github.com/factorysh/microdensity/storage"
	"github.com/factorysh/microdensity/task"
	"github.com/go-chi/chi/v5"
	"github.com/narqo/go-badge"
)

var colors = statusToColors{
	c: map[task.State]badge.Color{
		// blue - lapis
		task.Ready: "#2832C2",
		// red - ruby
		task.Canceled: "#900603",
		// orange - fire
		task.Running: "#DD571C",
		// red - ruby
		task.Failed: "#900603",
		// green - emerald
		task.Done: "#028A0F",
	},
	// blue
	Default: "#527284",
}

type statusToColors struct {
	c       map[task.State]badge.Color
	Default badge.Color
}

func (s statusToColors) Get(state task.State) badge.Color {
	c, found := s.c[state]
	if !found {
		return s.Default
	}

	return c
}

type Badge struct {
	Subject string `json:"subject"`
	Status  string `json:"status"`
	Color   string `json:"color"`
}

func (b *Badge) Render(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "image/svg+xml")
	err := badge.Render(b.Subject, b.Status, badge.Color(b.Color), w)
	if err != nil {
		panic(err)
	}
}

// StatusBadge handles request to for a badge task status request
func StatusBadge(s storage.Storage, latest bool) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		service := chi.URLParam(r, "serviceID")
		project := chi.URLParam(r, "project")
		branch := chi.URLParam(r, "branch")
		commit := chi.URLParam(r, "commit")

		t, err := s.GetByCommit(service, project, branch, commit, latest)

		if t == nil || err != nil {
			err = writeBadge("status", "?!", colors.Default, w)
			if err != nil {
				panic(err)
			}
			return
		}

		if t.Project != project {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		writeBadge("status", t.State.String(), colors.Get(t.State), w)
		if err != nil {
			panic(err)
		}
	}
}

// writeBadge is a wrapper use to write a badge into an http response
func writeBadge(label string, content string, color badge.Color, w http.ResponseWriter) error {
	w.Header().Set("content-type", "image/svg+xml")
	return badge.Render(label, content, color, w)
}
