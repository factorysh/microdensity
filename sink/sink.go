package sink

import "github.com/docker/go-events"

var _ events.Sink = (*VoidSink)(nil)

type VoidSink struct {
}

func (v *VoidSink) Write(events.Event) error {
	return nil
}

func (v *VoidSink) Close() error {
	return nil
}
