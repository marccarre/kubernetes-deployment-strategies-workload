.PHONY: all default build test integration-test push clean

NAME := marccarre/kds-service
VERSION := $(shell ./scripts/version)
BUILD_DATE := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
BUILD_IMAGE := golang:1.10-alpine
CURRENT_DIR := $(dir $(realpath $(firstword $(MAKEFILE_LIST))))

GO_SOURCES := $(shell find . -name '*.go')
GO_BINARY := ./cmd/service/service
GO_PROJECT_PATH := github.com/marccarre/kubernetes-deployment-strategies-workload

default: build
all: build test integration-test

# Go flags:
# - The -extldflags "-static" flag creates a static binary, i.e. with no external dependency.
# - The -s -w flags reduce the target's size.
# - The -i flag installs the packages that are dependencies of the target.
# - The -tags netgo flag enforces native Go networking, based on goroutines.
GO_FLAGS := -a -ldflags '-extldflags \"-static\" -s -w' -i -tags netgo

$(GO_BINARY): $(GO_SOURCES)
	docker run --rm \
		-v $(CURRENT_DIR):/go/src/$(GO_PROJECT_PATH) \
		--workdir /go/src/$(GO_PROJECT_PATH) \
		$(BUILD_IMAGE) \
		/bin/sh -c "GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go build $(GO_FLAGS) -o $@ ./$(@D)"

build: $(GO_BINARY)
	docker build \
		-t $(NAME):$(VERSION) \
		-t $(NAME):latest \
		--build-arg=REVISION=$(VERSION) \
		--build-arg=BUILD_DATE=$(BUILD_DATE) \
		$(CURRENT_DIR)

test:
	docker run --rm \
		-v $(CURRENT_DIR):/go/src/$(GO_PROJECT_PATH) \
		--workdir /go/src/$(GO_PROJECT_PATH) \
		$(BUILD_IMAGE) \
		/bin/sh -c "go test ./..."

integration-test:
	DB_CONTAINER="$$(docker run -d -e 'POSTGRES_DB=users_test' postgres:9.6.2)"; \
	docker run --rm \
		-v $(CURRENT_DIR):/go/src/$(GO_PROJECT_PATH) \
		--workdir /go/src/$(GO_PROJECT_PATH) \
		--link "$$DB_CONTAINER":users-db.local \
		$(BUILD_IMAGE) \
		/bin/sh -c "go test -tags integration -timeout 30s ./..."; \
	status=$$?; \
	docker rm -f "$$DB_CONTAINER"; \
	exit $$status

push:
	docker push $(NAME):$(VERSION)
	docker push $(NAME):latest

clean:
	-rm -f $(GO_BINARY)
	-docker rmi -f \
		$(NAME):$(VERSION) \
		$(NAME):latest
