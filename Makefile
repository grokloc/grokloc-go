IMG_DEV    = grokloc/grokloc-go:dev
DOCKER     = docker
DOCKER_RUN = $(DOCKER) run --rm -it
GO         = go
UNIT_ENVS  = --env-file ./env/unit.env
PORTS      = -p 3000:3000
CWD        = $(shell pwd)
BASE       = /grokloc
RUN        = $(DOCKER_RUN) -v $(CWD):$(BASE) -w $(BASE) $(UNIT_ENVS) $(PORTS) $(IMG_DEV)

.PHONY: docker
docker:
	$(DOCKER) build . -f Dockerfile -t $(IMG_DEV)

# Local go module operations.
.PHONY: mod
mod:
	$(GO) mod tidy
	$(GO) mod download
	$(GO) mod vendor
	$(GO) build ./...

# Shell in container.
.PHONY: shell
shell:
	$(RUN) /bin/bash

# Checks in local shell.
.PHONY: local-check
local-check:
	golangci-lint --timeout=24h run pkg/... && golint pkg/...

# Checks in container.
.PHONY: check
check:
	$(RUN) golangci-lint --timeout=24h run pkg/... && golint pkg/...

# Tests in local shell.
.PHONY: local-test
local-test:
	go test -v -race ./...

# Tests in container.
.PHONY: test
test:
	$(RUN) go test -v -race ./...
