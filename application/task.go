package application

import (
	"net/http"

	"github.com/go-chi/render"
)

func (a *Application) newTask(w http.ResponseWriter, r *http.Request) {
	var args map[string]string
	err := render.DecodeJSON(r.Body, args)
	if err != nil {
		panic(err)
	}
	a.queue.Put(nil)

}

func (a *Application) task(w http.ResponseWriter, r *http.Request) {
}
