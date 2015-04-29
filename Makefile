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

all: test build;

build:
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