package task

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestTask(t *testing.T) {
	task := &Task{
		Id:      uuid.New(),
		Project: "group%20project",
		Branch:  "main",
		Commit:  "01279848527693d126de60ec7b355924c96d2957",
	}

	err := task.Validate()
	assert.NoError(t, err)

}
