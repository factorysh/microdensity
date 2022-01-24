package badge

import (
	"fmt"
	"net/http"

	"github.com/factorysh/microdensity/httpcontext"
	"github.com/factorysh/microdensity/queue"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/narqo/go-badge"
)

func BadgeMyProject(q *queue.Queue, label string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		project := r.Context().Value(httpcontext.RequestedProject).(string)
		id := chi.URLParam(r, "id")
		uid, err := uuid.Parse(id)
		if err != nil {
			panic(err)
		}
		fmt.Println("id", id, uid)
		t, err := q.Get(uid)
		if err != nil {
			panic(err)
		}
		w.Header().Set("content-type", "image/svg+xml")
		if t == nil {
			badge.Render(label, "?!", "#5272B4", w)
			return
		}
		if t.Project != project {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		badge.Render(label, fmt.Sprintf("%d", t.State), "#5272B4", w)
	}
}
