package sessions

import (
	"sync"
	"time"

	"github.com/factorysh/microdensity/gitlab"
)

// UserData user sessions contains user data
type UserData struct {
	Expires time.Time
	Project *gitlab.ProjectInfo
}

// Sessions pool
type Sessions struct {
	sync.RWMutex
	pool map[string]*UserData
}

// Put UserData into session pool
func (s *Sessions) Put(accessToken string, expires time.Time, project *gitlab.ProjectInfo) {
	s.Lock()
	s.pool[accessToken] = &UserData{
		Expires: expires,
		Project: project,
	}
	s.Unlock()
}

// New inits a new sessions struct
func New() Sessions {
	return Sessions{
		pool: make(map[string]*UserData),
	}
}
