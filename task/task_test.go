package task

import (
	"strings"
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

	err = (&Task{}).Validate()
	assert.Error(t, err)
	assert.True(t, strings.HasPrefix(err.Error(), "empty id"))

	err = (&Task{
		Id:      uuid.New(),
		Project: "group/project",
	}).Validate()
	assert.Error(t, err)
	assert.True(t, strings.HasPrefix(err.Error(), "project name must be url escaped"), err.Error())

	err = (&Task{
		Id:      uuid.New(),
		Project: "group%20project",
		Branch:  "feature/machin",
	}).Validate()
	assert.Error(t, err)
	assert.True(t, strings.HasPrefix(err.Error(), "branch name must be url escaped"), err.Error())

	err = (&Task{
		Id:      uuid.New(),
		Project: "group%20project",
		Branch:  "feature%20machin",
		Commit:  "beuha aussi",
	}).Validate()
	assert.Error(t, err)
	assert.True(t, strings.HasPrefix(err.Error(), "bad commit format"), err.Error())
}
