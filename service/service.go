package service

type Service interface {
	// Validate the input coming from HTTP Body as a JSON, and fight against XSS
	//Validate is sync
	Validate(map[string]interface{}) error
	// When the queue is ok, run async
	// Run can use the user's Sentry ID
	Run(args map[string]interface{}) error
	// What's service name?
	Name() string
}
