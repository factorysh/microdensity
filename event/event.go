package event

import (
	"github.com/factorysh/microdensity/task"
	"github.com/google/uuid"
)

type Event struct {
	Id    uuid.UUID  `json:"id"`
	State task.State `json:"state"`
	Error error      `json:"error,omit_empty"`
}
