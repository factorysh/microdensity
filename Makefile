GIT_VERSION?=$(shell git describe --tags --always --abbrev=42 --dirty)

build:
	CGO_ENABLED=0 go build \
		-ldflags "-X github.com/factorysh/microdensity/version.version=$(GIT_VERSION)" \
	.

build-linux:
	make build GOOS=linux
	upx microdensity

TESTS= github.com/factorysh/microdensity/task \
		github.com/factorysh/microdensity/middlewares/jwt \
		github.com/factorysh/microdensity/middlewares/project\
		github.com/factorysh/microdensity/middlewares/oauth2 \
		github.com/factorysh/microdensity/middlewares \
		github.com/factorysh/microdensity/sessions \
		github.com/factorysh/microdensity/badge \
		github.com/factorysh/microdensity/gitlab \
		github.com/factorysh/microdensity/oauth \
		github.com/factorysh/microdensity/volumes \
		github.com/factorysh/microdensity/storage \

test:
	go test --cover ${TESTS}

test-all:
	go test --cover ${TESTS} \
		github.com/factorysh/microdensity/run \
		github.com/factorysh/microdensity/queue \
		github.com/factorysh/microdensity/application \
