package badge

import (
	"net/http"

	"github.com/factorysh/microdensity/storage"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/narqo/go-badge"
)

func BadgeMyProject(s storage.Storage, label string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		project := chi.URLParam(r, "project")
		id := chi.URLParam(r, "id")
		uid, err := uuid.Parse(id)
		if err != nil {
			panic(err)
		}
		t, err := s.Get(uid.String())
		if err != nil {
			panic(err)
		}
		w.Header().Set("content-type", "image/svg+xml")
		if t == nil {
			err = badge.Render(label, "?!", "#5272B4", w)
			if err != nil {
				panic(err)
			}
			return
		}
		if t.Project != project {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err = badge.Render(label, t.State.String(), "#5272B4", w)
		if err != nil {
			panic(err)
		}
	}
}
