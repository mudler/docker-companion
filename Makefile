NAME ?= docker-companion
PACKAGE_NAME ?= $(NAME)
PACKAGE_CONFLICT ?= $(PACKAGE_NAME)-beta
REVISION := $(shell git rev-parse --short HEAD || echo unknown)
VERSION := $(shell git describe --tags || cat main.go | grep -o 'VERSION = "[^"]*"' | awk '{ print $3 }' | sed 's:"::g' || echo dev)
VERSION := $(shell echo $(VERSION) | sed -e 's/^v//g')
ITTERATION := $(shell date +%s)
BUILD_PLATFORMS ?= -os="linux" -os="darwin"

all: deps test lint build

help:
	# make all => deps test lint build
	# make deps - install all dependencies
	# make test - run project tests
	# make lint - check project code style
	# make build - build project for all supported OSes

deps:
	# Installing dependencies...
	go get github.com/Masterminds/glide
	go get -u github.com/golang/lint/golint
	go get github.com/mitchellh/gox
	go get golang.org/x/tools/cmd/cover
	go get github.com/tcnksm/ghr
	go get github.com/axw/gocov/gocov
	go get github.com/mattn/goveralls
	glide install
	go build

deps-add:
	glide guess glide.yaml

deps-update:
	glide up

build:
	# Building gitlab-ci-multi-runner for $(BUILD_PLATFORMS)
	gox $(BUILD_PLATFORMS) -output="release/$(NAME)-$(VERSION)-{{.OS}}-{{.Arch}}"

lint:
	# Checking project code style...
	golint ./... | grep -v "be unexported"

test:
	# Running tests... ${TOTEST}
	go test -cover

build-and-deploy:
	make build BUILD_PLATFORMS="-os=linux -arch=amd64"
