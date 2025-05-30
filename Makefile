# Project Variables
PROJECT_NAME := friemon
GOFILES := $(shell find . -name '*.go' -not -path './vendor/*')

# Git Variables
GIT_TAG := $(shell git describe --tags --always --dirty)
GIT_COMMIT := $(shell git rev-parse --short HEAD)
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD)

# Commands
GOCMD := go
GOTEST := $(GOCMD) test
GOBUILD := $(GOCMD) build
GORUN := $(GOCMD) run
GOFMT := $(GOCMD) fmt
GOVET := $(GOCMD) vet
GOMOD := $(GOCMD) mod

# Build Flags
SYNC ?= false

DEFAULT_FLAGS ?= -commit=$(GIT_COMMIT) -branch=$(GIT_BRANCH)

default: build

.PHONY: build
build:
	@echo "Building $(PROJECT_NAME)..."
	$(GOBUILD) $(DEFAULT_FLAGS) -o bin/$(PROJECT_NAME) ./main.go

.PHONY: run
run:
	@echo "Running $(PROJECT_NAME)..."
	$(GORUN) ./cmd/friemon/main.go $(DEFAULT_FLAGS)

.PHONY: test
test:
	@echo "Running tests..."
	$(GOTEST) ./... -v

.PHONY: fmt
fmt:
	@echo "Formatting code..."
	$(GOFMT) $(GOFILES)

.PHONY: vet
vet:
	@echo "Running 'go vet'..."
	$(GOVET) ./...

.PHONY: tidy
tidy:
	@echo "Tidying up Go modules..."
	$(GOMOD) tidy

.PHONY: clean
clean:
	@echo "Cleaning up..."
	rm -rf bin/

.PHONY: all
all: fmt vet test build
