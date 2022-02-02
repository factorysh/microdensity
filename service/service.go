package service

import "github.com/google/uuid"

type Service interface {
	// Validate the input coming from HTTP Body as a JSON, and fight against XSS
	//Validate is sync
	Validate(map[string]interface{}) (Arguments, error)
	// When the queue is ok, run async
	// Run can use the user's Sentry ID
	New(project string, args map[string]interface{}) (uuid.UUID, error)
	Run(id uuid.UUID) error
	// TODO
	//Watch(id uuid.UUID) error
	// What's service name?
	Name() string
}
