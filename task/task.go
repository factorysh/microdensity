package task

import (
	"time"

	"github.com/google/uuid"
)

type Task struct {
	Id       uuid.UUID
	Project  string
	Creation time.Time
	Args     map[string]interface{}
}
