package sessions

import (
	"testing"
	"time"

	"github.com/factorysh/microdensity/gitlab"
	"github.com/stretchr/testify/assert"
)

const delta = 10 * time.Second

var dummyProject = gitlab.ProjectInfo{
	PathWithNamespace: "factorysh/test",
	Name:              "factorysh/test",
	ID:                42,
	Archived:          false,
}

func TestUserDataIsValid(t *testing.T) {

	tests := []struct {
		name  string
		input UserData
		want  bool
	}{
		{name: "Valid User Data", input: UserData{Expires: time.Now().Add(delta)}, want: true},
		{name: "Invalid User Data", input: UserData{Expires: time.Now().Add(-delta)}, want: false},
	}

	for _, tc := range tests {
		got := tc.input.IsValid()
		assert.Equal(t, tc.want, got)
	}

}

func TestUserDataMatchRequestedProject(t *testing.T) {

	tests := []struct {
		name             string
		input            UserData
		requestedProject string
		want             bool
	}{
		{name: "Same project name", input: UserData{Project: &dummyProject}, requestedProject: "factorysh/test", want: true},
		{name: "Different project name", input: UserData{Project: &dummyProject}, requestedProject: "factorysh/another", want: false},
		{name: "Same ID", input: UserData{Project: &dummyProject}, requestedProject: "42", want: true},
		{name: "Different ID", input: UserData{Project: &dummyProject}, requestedProject: "41", want: false},
	}

	for _, tc := range tests {
		got := tc.input.MatchRequestedProject(tc.requestedProject)
		assert.Equal(t, tc.want, got)
	}
}

func TestSessionsPut(t *testing.T) {
	s := New()
	s.Put("id", "token", time.Now().Add(delta), &dummyProject)
	assert.Equal(t, s.Len(), 1)
}

func TestSessionsGet(t *testing.T) {
	s := New()
	s.Put("id", "token", time.Now().Add(delta), &dummyProject)
	assert.Equal(t, s.Len(), 1)
	ud, found := s.Get("id")
	assert.NotNil(t, ud)
	assert.True(t, found)
}

func TestSessionsRemove(t *testing.T) {
	s := New()
	s.Put("id", "token", time.Now().Add(delta), &dummyProject)
	assert.Equal(t, s.Len(), 1)
	s.Remove("id")
	assert.Equal(t, s.Len(), 0)
}

func TestSessionsPrune(t *testing.T) {
	s := New()
	s.Put("billy", "token", time.Now().Add(-delta), &dummyProject)
	s.Put("bob", "token", time.Now().Add(-delta), &dummyProject)
	s.Put("alekei", "token", time.Now().Add(-delta), &dummyProject)
	s.Put("nancy", "token", time.Now().Add(delta), &dummyProject)
	s.Put("dustin", "token", time.Now().Add(delta), &dummyProject)
	s.Put("steve", "token", time.Now().Add(delta), &dummyProject)
	assert.Equal(t, s.Len(), 6)

	s.Prune()
	assert.Equal(t, s.Len(), 3)
}
