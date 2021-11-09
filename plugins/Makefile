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
	@make -C gateway/kong/ docker
