MAKEFLAGS += --no-builtin-rules
.SUFFIXES:
.SECONDARY:
.DELETE_ON_ERROR:

ARGS  ?= -v
TESTS ?= ./... -cover
COVER ?=
SRC   := $(shell find . -name '*.go')

# If the first argument to make is "migrate" use the rest as arguments for the migrate target...
ifeq (migrate,$(firstword $(MAKECMDGOALS)))
  RUN_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  $(eval $(RUN_ARGS):;@:) # ...and turn them into do-nothing targets
endif

# Prefix the user's GOPATH with our godep path. I don't want to totally blast the user's GOPATH here.
# TODO: Is it worth ensuring that godep is installed here?
export GOPATH := $(shell godep path):$(GOPATH)

all: test build;

build:
	@echo 'Building with GOPATH: $(GOPATH)'
	@go build $(GOFLAGS) -o build/migrate cmd/migrate/main.go

clean:
	@rm -rf build

test: $(SRC)
	@go test $(TESTS) $(ARGS)

test-cover: $(SRC)
	@go test $(COVER) -coverprofile=coverage.out
	@go tool cover -html=coverage.out

migrate:
	@go run cmd/migrate/main.go $(RUN_ARGS)

.PHONY: migrate