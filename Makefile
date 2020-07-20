# The binary to build (just the basename).
BIN := $(shell basename $$PWD)

# Where to push the docker image.
REGISTRY ?= quay.io/oriedge

# Image URL to use all building/pushing image targets
IMG ?= $(REGISTRY)/$(BIN)

setup:
	-./test/kind-with-registry.sh &>/dev/null

up: setup
	tilt up

down:
	tilt down

nuke: 
	kind delete cluster --name kind

build:
	CGO_ENABLED=0 go build cmd/coredns.go

.PHONY: test
test:
	go test -race ./... -short

clean:
	go clean
	rm -f coredns

image: 
	docker build . -t ${IMG}