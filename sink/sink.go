package sink

/*
µdensity raise events htrought its workflow
*/

import "github.com/docker/go-events"

var _ events.Sink = (*VoidSink)(nil)

// VoidSink accepts events and does nothing with them.
type VoidSink struct {
}

func (v *VoidSink) Write(events.Event) error {
	return nil
}

func (v *VoidSink) Close() error {
	return nil
}
