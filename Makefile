build:
	CGO_ENABLED=0 go build .

test:
	go test --cover \
		github.com/factorysh/microdensity/middlewares \
		github.com/factorysh/microdensity/queue \
		github.com/factorysh/microdensity/sessions \
		github.com/factorysh/microdensity/application \
		github.com/factorysh/microdensity/gitlab \
		github.com/factorysh/microdensity/oauth \
		github.com/factorysh/microdensity/run
