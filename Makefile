PACKAGES := $(shell go list ./...)
name := $(shell basename ${PWD})
checksum := $(shell git rev-parse --short HEAD)

all: help

.PHONY: help
help: Makefile
	@echo
	@echo " Choose a make command to run"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo

## build: build a dev binary
.PHONY: build
build: test vet
	go build -ldflags "-X 'github.com/proxati/llm_proxy/version.gitHeadChecksum=$(checksum)'" -o llm_proxy -v

## release: build a binary for release
.PHONY: release
release: test vet
	go build -ldflags "-X 'github.com/proxati/llm_proxy/version.dev=no'" -o llm_proxy -v

## vet: vet code
.PHONY: vet
vet:
	go vet $(PACKAGES)

## test: run unit tests
.PHONY: test
test:
	go test -race -cover $(PACKAGES)
