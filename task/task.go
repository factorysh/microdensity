package task

import (
	"time"

	"github.com/google/uuid"
)

type State int

const (
	Ready State = iota
	Running
	Canceled
	Error
	Done
)

type Task struct {
	Id       uuid.UUID
	Project  string
	Creation time.Time
	Args     map[string]interface{}
	State    State
}
