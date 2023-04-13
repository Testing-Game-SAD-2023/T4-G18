APP_NAME=$(shell basename $(CURDIR))
GIT_COMMIT=$(shell git rev-parse --short HEAD)
FILES=$(wildcard *.go)

OUT_DIR=build

.PHONY: build run dev

## build: build the project in a single file executable in "build" directory
build:
	CGO_ENABLED=0 go build -o $(OUT_DIR)/$(APP_NAME) -ldflags "-X main.gitCommit=$(GIT_COMMIT)" $(FILES)

## run: run application
run: build
	./$(OUT_DIR)/$(APP_NAME)

.PHONY: help
## help: prints this help message
help:
	@echo "Usage: \n"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

