package application

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/docker/go-events"
	_event "github.com/factorysh/microdensity/event"
	"github.com/factorysh/microdensity/task"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type HttpSink struct {
	w             http.ResponseWriter
	flusher       http.Flusher
	json          *json.Encoder
	wg            *sync.WaitGroup
	isEventSource bool
	lock          *sync.Mutex
	cancel        context.CancelFunc
}

func NewHttpSink(r *http.Request, w http.ResponseWriter, waitForEnd bool) (*HttpSink, error) {
	isEventSource := false
	for _, accept := range strings.Split(r.Header.Get("accept"), ", ") {
		// https://developer.mozilla.org/fr/docs/Web/API/EventSource
		if strings.Split(accept, ";")[0] == "text/event-stream" {
			isEventSource = true
			w.Header().Set("Content-Type", "text/event-stream")
			break
		}
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, errors.New("ResponseWriter can't be upgraded to Flusher")
	}

	h := &HttpSink{
		w:             w,
		flusher:       flusher,
		json:          json.NewEncoder(w),
		isEventSource: isEventSource,
		lock:          &sync.Mutex{},
		cancel:        func() {},
	}
	if waitForEnd {
		h.wg = &sync.WaitGroup{}
		h.wg.Add(1)
	}
	var ctx context.Context
	ctx, h.cancel = context.WithCancel(context.TODO())

	go func(ctx context.Context) {
		tick := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-ctx.Done():
				return
			case <-tick.C:
				h.lock.Lock()
				if isEventSource {
					h.w.Write([]byte("event: ping\n\n"))
				} else {
					h.w.Write([]byte("\n"))
				}
				h.flusher.Flush()
				h.lock.Unlock()
			}
		}
	}(ctx)
	return h, nil
}

func (h *HttpSink) Write(event events.Event) error {
	h.lock.Lock()
	defer h.lock.Unlock()
	if h.isEventSource {
		h.w.Write([]byte("event: task\ndata: "))
	}
	err := h.json.Encode(event)
	if err != nil {
		return err
	}
	if h.isEventSource {
		h.w.Write([]byte("\n"))
	}
	h.flusher.Flush()
	e, _ := event.(_event.Event)
	if h.wg != nil && e.State != task.Ready && e.State != task.Running {
		h.wg.Done()
	}
	return nil
}

func (h *HttpSink) Close() error {
	h.cancel()
	return nil
}

func (h *HttpSink) Wait() {
	h.wg.Wait()
}

// SinkAllHandler show 5 minutes of Task events
func (a *Application) SinkAllHandler(w http.ResponseWriter, r *http.Request) {
	h, err := NewHttpSink(r, w, false)
	if err != nil {
		a.logger.With(zap.Error(err))
		a.logger.Error("SinkAlleHandler error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer h.Close()
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

// SinkHandler show the events of a specific run and stop
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

	s, err := NewHttpSink(r, w, true)
	if err != nil {
		l.With(zap.Error(err)).Error("SinkHandler error")
	}
	defer s.Close()
	f := events.NewFilter(s, &TaskMatcher{task: t})
	a.Sink.Add(f)
	defer a.Sink.Remove(f)
	s.Wait()
}
