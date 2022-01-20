package sessions

import (
	"net/url"
	"sync"
	"time"

	"github.com/factorysh/microdensity/gitlab"
)

// UserData user sessions contains user data
type UserData struct {
	Expires time.Time
	Project *gitlab.ProjectInfo
}

// IsValid is used to check is user data is expired
func (ud *UserData) IsValid() bool {
	return ud.Expires.After(time.Now())
}

// MatchRequestedProject check is this user session match the required project
func (ud *UserData) MatchRequestedProject(requestedProject string) bool {
	requestedProject, err := url.PathUnescape(requestedProject)
	if err != nil {
		return false
	}

	// TODO : handle both ID and full project name

	return ud.Project.PathWithNamespace == requestedProject
}

// Sessions pool
type Sessions struct {
	sync.RWMutex
	pool map[string]*UserData
}

// Authorize user to access to matching ressource
func (s *Sessions) Authorize(accessToken string, projectName string) bool {
	s.RLock()
	ud, found := s.pool[accessToken]
	s.RUnlock()
	if !found {
		return false
	}

	return ud.IsValid() && ud.MatchRequestedProject(projectName)
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

// Get UserData from session pool
func (s *Sessions) Get(accessToken string) (*UserData, bool) {
	s.RLock()
	u, found := s.pool[accessToken]
	s.RUnlock()

	return u, found
}

// Remove UserData from session pool
func (s *Sessions) Remove(accessToken string) {
	s.Lock()
	delete(s.pool, accessToken)
	s.Unlock()
}

// TODO: prune sessions

// New inits a new sessions struct
func New() Sessions {
	return Sessions{
		pool: make(map[string]*UserData),
	}
}
