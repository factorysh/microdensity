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
	Id       uuid.UUID              `json:"id"`
	Service  string                 `json:"service"`
	Project  string                 `json:"project"`
	Branch   string                 `json:"branch"`
	Commit   string                 `json:"commit"`
	Creation time.Time              `json:"creation"`
	Args     map[string]interface{} `json:"Args"`
	State    State
}
