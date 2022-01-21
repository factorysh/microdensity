package sessions

import (
	"crypto/rand"
	"encoding/base64"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/factorysh/microdensity/gitlab"
)

const idLen = 256

// GenID generates a session id
func GenID() (string, error) {
	s := make([]byte, idLen)
	_, err := rand.Read(s)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(s), nil
}

// UserData user sessions contains user data
type UserData struct {
	accessToken string
	Expires     time.Time
	Project     *gitlab.ProjectInfo
}

// GetToken return the accessToken value
func (ud *UserData) GetToken() string {
	return ud.accessToken
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

	// if it's an int, it's a project id
	if pID, err := strconv.Atoi(requestedProject); err == nil {
		return pID == ud.Project.ID
	}

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
func (s *Sessions) Put(sessionID string, accessToken string, expires time.Time, project *gitlab.ProjectInfo) {
	s.Lock()
	s.pool[sessionID] = &UserData{
		accessToken: accessToken,
		Expires:     expires,
		Project:     project,
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
