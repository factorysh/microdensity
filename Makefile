GIT_VERSION?=$(shell git describe --tags --always --abbrev=42 --dirty)

build:
	CGO_ENABLED=0 go build \
		-ldflags "-X github.com/factorysh/microdensity/version.version=$(GIT_VERSION)" \
	.

test:
	go test --cover \
		github.com/factorysh/microdensity/task \
		github.com/factorysh/microdensity/middlewares \
		github.com/factorysh/microdensity/queue \
		github.com/factorysh/microdensity/sessions \
		github.com/factorysh/microdensity/badge \
		github.com/factorysh/microdensity/application \
		github.com/factorysh/microdensity/gitlab \
		github.com/factorysh/microdensity/oauth \
		github.com/factorysh/microdensity/volumes \
		github.com/factorysh/microdensity/run
