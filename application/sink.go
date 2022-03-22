package application

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/docker/go-events"
)

type HttpSink struct {
	w       http.ResponseWriter
	flusher http.Flusher
	json    *json.Encoder
}

func (h *HttpSink) Write(event events.Event) error {
	err := h.json.Encode(event)
	if err != nil {
		return err
	}
	_, err = h.w.Write([]byte("\n"))
	if err != nil {
		return err
	}
	h.flusher.Flush()
	return nil
}

func (h *HttpSink) Close() error {
	return nil
}

func (a *Application) SinkAllHandler(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		a.logger.Error("ResponseWriter can't be upgraded to Flusher")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	h := &HttpSink{
		w:       w,
		flusher: flusher,
		json:    json.NewEncoder(w),
	}
	a.Sink.Add(h)
	defer a.Sink.Remove(h)
	time.Sleep(5 * time.Minute)
}
