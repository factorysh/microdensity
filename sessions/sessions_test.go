package sessions

import (
	"testing"
	"time"

	"github.com/factorysh/microdensity/gitlab"
	"github.com/stretchr/testify/assert"
)

const delta = 10 * time.Second

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
		t.Run(tc.name, func(t *testing.T) {
			got := tc.input.IsValid()
			assert.Equal(t, tc.want, got)
		})
	}

}

func TestUserDataMatchRequestedProject(t *testing.T) {

	tests := []struct {
		name             string
		input            UserData
		requestedProject string
		want             bool
	}{
		{name: "Same project name", input: UserData{Project: &gitlab.DummyProject}, requestedProject: "group/project", want: true},
		{name: "Different project name", input: UserData{Project: &gitlab.DummyProject}, requestedProject: "factorysh/another", want: false},
		{name: "Same ID", input: UserData{Project: &gitlab.DummyProject}, requestedProject: "42", want: true},
		{name: "Different ID", input: UserData{Project: &gitlab.DummyProject}, requestedProject: "41", want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.input.MatchRequestedProject(tc.requestedProject)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestSessionsPut(t *testing.T) {
	s := New()
	s.Put("id", "token", time.Now().Add(delta), &gitlab.DummyProject)
	assert.Equal(t, s.Len(), 1)
}

func TestSessionsGet(t *testing.T) {
	s := New()
	s.Put("id", "token", time.Now().Add(delta), &gitlab.DummyProject)
	assert.Equal(t, 1, s.Len())
	ud, found := s.Get("id")
	assert.NotNil(t, ud)
	assert.True(t, found)
}

func TestSessionsRemove(t *testing.T) {
	s := New()
	s.Put("id", "token", time.Now().Add(delta), &gitlab.DummyProject)
	assert.Equal(t, 1, s.Len())
	s.Remove("id")
	assert.Equal(t, 0, s.Len())
}

func TestSessionsPrune(t *testing.T) {
	s := New()
	s.Put("billy", "token", time.Now().Add(-delta), &gitlab.DummyProject)
	s.Put("bob", "token", time.Now().Add(-delta), &gitlab.DummyProject)
	s.Put("alekei", "token", time.Now().Add(-delta), &gitlab.DummyProject)
	s.Put("nancy", "token", time.Now().Add(delta), &gitlab.DummyProject)
	s.Put("dustin", "token", time.Now().Add(delta), &gitlab.DummyProject)
	s.Put("steve", "token", time.Now().Add(delta), &gitlab.DummyProject)
	assert.Equal(t, 6, s.Len())

	s.Prune()
	assert.Equal(t, 3, s.Len())
}
