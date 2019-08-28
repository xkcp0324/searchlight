#
#   make              - default to 'build' target
#   make test         - run unit test
#   make build        - build local binary targets
#   make docker-build - build local binary targets by docker
#   make container    - build containers
#   make push         - push containers
#   make clean        - clean up targets
#
# The makefile is also responsible to populate project version information.

#
# Tweak the variables based on your project.
#

# Current version of the project.
VERSION ?= v8.0.0
DOCKER_VERSION := 8.0.0-k8s-1

# Target binaries. You can build multiple binaries for a single project.
TARGETS ?= hyperalert

# Container registries.
REGISTRY := registry.cn-hangzhou.aliyuncs.com/dmall

# Container image prefix and suffix added to targets.
# The final built images are:
#   $[REGISTRY]$[IMAGE_PREFIX]$[TARGET]$[IMAGE_SUFFIX]:$[VERSION]
# $[REGISTRY] is an item from $[REGISTRIES], $[TARGET] is an item from $[TARGETS].
IMAGE_PREFIX ?= $(strip )
IMAGE_SUFFIX ?= $(strip )

# This repo's root import path (under GOPATH).
ROOT := github.com/appscode/searchlight

# Project main package location (can be multiple ones).
CMD_DIR := ./cmd

# Project output directory.
OUTPUT_DIR := ./bin

# docker file direcotory.
DOCKER_DIR := ./docker

# Git commit sha.
COMMIT := $(strip $(shell git rev-parse --short HEAD 2>/dev/null))
COMMIT := $(COMMIT)$(shell git diff-files --quiet || echo '-dirty')
COMMIT := $(if $(COMMIT),$(COMMIT),"Unknown")


GO_VERSION := 1.12.9
ARCH     ?= $(shell go env GOARCH)
BuildDate = $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
Commit    = $(shell git rev-parse --short HEAD)
GOENV     := CGO_ENABLED=0 GOOS=$(shell uname -s | tr A-Z a-z) GOARCH=$(ARCH)
GO        := $(GOENV) go build -v -a

#
# Define all targets. At least the following commands are required:
#

.PHONY: build container push test clean

build:
	$(GO) -o $(OUTPUT_DIR)/hyperalert -ldflags "-s -w -X ./cmd/hyperalert/main.Version=$(VERSION) -X ./cmd/hyperalert/main.CommitHash=$(COMMIT) -X ./cmd/hyperalert/main.BuildTimestamp=$(BuildDate)"  $(CMD_DIR)/hyperalert


docker-build:
	docker run --rm -v "$$PWD":/go/src/${ROOT} -w /go/src/${ROOT} golang:${GO_VERSION} make build


container:
	docker build -t $(REGISTRY)/icinga:$(DOCKER_VERSION) -f ./Dockerfile .


push: container
	docker push $(REGISTRY)/icinga:$(DOCKER_VERSION)


test:
	@go test ./...

clean:
	@rm -vrf ${OUTPUT_DIR}/*
