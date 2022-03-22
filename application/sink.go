package application

import (
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/docker/go-events"
	_event "github.com/factorysh/microdensity/event"
	"github.com/factorysh/microdensity/task"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type HttpSink struct {
	w       http.ResponseWriter
	flusher http.Flusher
	json    *json.Encoder
	wg      *sync.WaitGroup
}

func NewHttpSink(w http.ResponseWriter, waitForEnd bool) (*HttpSink, error) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, errors.New("ResponseWriter can't be upgraded to Flusher")
	}
	h := &HttpSink{
		w:       w,
		flusher: flusher,
		json:    json.NewEncoder(w),
	}
	if waitForEnd {
		h.wg = &sync.WaitGroup{}
		h.wg.Add(1)
	}
	return h, nil
}

func (h *HttpSink) Write(event events.Event) error {
	err := h.json.Encode(event)
	if err != nil {
		return err
	}
	h.flusher.Flush()
	e, _ := event.(_event.Event)
	if h.wg != nil && e.State != task.Ready && e.State != task.Running {
		h.wg.Done()
	}
	return nil
}

func (h *HttpSink) Close() error {
	return nil
}

func (h *HttpSink) Wait() {
	h.wg.Wait()
}

func (a *Application) SinkAllHandler(w http.ResponseWriter, r *http.Request) {
	h, err := NewHttpSink(w, false)
	if err != nil {
		a.logger.With(zap.Error(err))
		a.logger.Error("SinkAlleHandler error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	a.Sink.Add(h)
	defer a.Sink.Remove(h)
	time.Sleep(5 * time.Minute)
}

type TaskMatcher struct {
	task *task.Task
}

func (t *TaskMatcher) Match(event events.Event) bool {
	e, _ := event.(_event.Event)
	return e.Id == t.task.Id
}

func (a *Application) SinkHandler(w http.ResponseWriter, r *http.Request) {
	l := a.logger.With(
		zap.String("url", r.URL.String()),
		zap.String("service", chi.URLParam(r, "serviceID")),
		zap.String("project", chi.URLParam(r, "project")),
		zap.String("branch", chi.URLParam(r, "branch")),
		zap.String("commit", chi.URLParam(r, "commit")),
	)

	t, err := a.storage.GetByCommit(
		chi.URLParam(r, "serviceID"),
		chi.URLParam(r, "project"),
		chi.URLParam(r, "branch"),
		chi.URLParam(r, "commit"),
		true,
	)
	if err != nil {
		l.Warn("Task get error", zap.Error(err))
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(http.StatusText(http.StatusNotFound)))
		return
	}

	s, err := NewHttpSink(w, true)
	if err != nil {
		l.With(zap.Error(err)).Error("SinkHandler error")
	}
	f := events.NewFilter(s, &TaskMatcher{task: t})
	a.Sink.Add(f)
	defer a.Sink.Remove(f)
	s.Wait()
}
