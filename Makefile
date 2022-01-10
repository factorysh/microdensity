build:
	go build .

test:
	go test --cover \
		github.com/factorysh/microdensity/middlewares \
		github.com/factorysh/microdensity/queue
