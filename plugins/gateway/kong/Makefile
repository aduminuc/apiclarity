# Project variables
BINARY_NAME ?= kong-pluging
DOCKER_REGISTRY ?= ghcr.io/apiclarity
VERSION ?= $(shell git rev-parse HEAD)
DOCKER_IMAGE ?= $(DOCKER_REGISTRY)/$(BINARY_NAME)
DOCKER_TAG ?= ${VERSION}

.PHONY: all
all: kong-plugin

.PHONY: docker
docker: ## Build Docker image
	@DOCKER_BUILDKIT=1 docker build -t ${DOCKER_IMAGE}:${DOCKER_TAG} --file Dockerfile .

.PHONY: push-docker
push-docker: docker ## Build and Push Docker image
	@echo "Publishing Docker image ..."
	docker push ${DOCKER_IMAGE}:${DOCKER_TAG}

kong-plugin: plugin.go
	go build -o bin/kong-plugin plugin.go

.PHONY: test
test: ## Run Unit Tests
	@(go test ./...)

.PHONY: lint
lint: bin/golangci-lint
	../../../bin/golangci-lint run

.PHONY: check
check: lint test ## Run tests and linters

.PHONY: license-check
license-check: bin/licensei ## Run license check
	../../../bin/licensei header
	../../../bin/licensei check --config=../.licensei.toml

.PHONY: license-cache
license-cache: bin/licensei ## Generate license cache
	../../../bin/licensei cache

.PHONY: clean
clean:
	@(rm -rf bin ; echo "Kong plugin cleanup done" )