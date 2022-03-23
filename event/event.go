package event

import (
	"encoding/json"

	"github.com/factorysh/microdensity/task"
	"github.com/google/uuid"
)

type Event struct {
	Id    uuid.UUID  `json:"id"`
	State task.State `json:"state"`
	Error error      `json:"error,omit_empty"`
}

func (e Event) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Id    uuid.UUID `json:"id"`
		State string    `json:"state"`
		Error error     `json:"error"`
	}{
		Id:    e.Id,
		State: e.State.String(),
		Error: e.Error,
	})
}
