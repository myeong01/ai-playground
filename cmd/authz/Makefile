REPO_PATH := $(shell git rev-parse --show-toplevel)
CHANGED_FILES := $(shell git diff-files)

ifeq ($(strip $(CHANGED_FILES)),)
GIT_VERSION := $(shell git describe --tags --long --always)
else
GIT_VERSION := $(shell git describe --tags --long --always)-dirty-$(shell git diff | shasum -a256 | cut -c -6)
endif

#IMG ?= myoon/ai-playground-authz
IMG ?= registry.gitlab.com/myeongsuk.yoon/temp
TAG ?= v00.00.01
#TAG ?= $(GIT_VERSION)

all: build

build:
#	CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o bin/oidc-authservice
	GOOS=linux GOARCH=amd64 go build -a -ldflags '-extldflags "-static"' -o bin/oidc-authservice
	#chmod +x bin/oidc-authservice

docker-build-amd64: build
	docker buildx build --platform linux/amd64 -t $(IMG):$(TAG) .

docker-build:
	docker build -t $(IMG):$(TAG) .

docker-push:
	docker push $(IMG):$(TAG)

publish: docker-build-amd64 docker-push

