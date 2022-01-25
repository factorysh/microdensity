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
	Failed
	Done
)

func (s State) String() string {
	return []string{"Ready", "Running", "Canceled", "Failed", "Done"}[s]
}

type Task struct {
	Id       uuid.UUID
	Service  string
	Commit   string
	Branch   string
	Project  string
	Creation time.Time
	Args     map[string]interface{}
	State    State
}
