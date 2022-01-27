package task

import (
	"errors"
	"fmt"
	"strings"
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

func (t *Task) Validate() error {
	if t.Id == uuid.Nil {
		return errors.New("empty id not allowed")
	}
	if strings.ContainsRune(t.Project, '/') {
		return fmt.Errorf("project name must be url escaped, without any / : %s", t.Project)
	}
	if strings.ContainsRune(t.Branch, '/') {
		return fmt.Errorf("branch name must be url escaped, without any / : %s", t.Branch)
	}
	// FIXME assert commit is a sha
	return nil
}
