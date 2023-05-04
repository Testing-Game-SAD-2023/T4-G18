APP_NAME=$(shell basename $(CURDIR))
GIT_COMMIT=$(shell git rev-parse --short HEAD)

OUT_DIR=build

.PHONY: build run dev

## build: build the project in a single file executable in "build" directory
build:
	CGO_ENABLED=0 go build -o $(OUT_DIR)/$(APP_NAME) -ldflags "-X main.gitCommit=$(GIT_COMMIT)"

## run: run application
run: build
	./$(OUT_DIR)/$(APP_NAME)

## docker-build: build a docker image
docker-build:
	docker build -t $(APP_NAME):$(GIT_COMMIT) .

## docker-run: runs a docker container exposing port 3000 and using config.json
docker-run:
	docker run -p 3000:3000 -v $(shell pwd)/config.json:/app/config.json $(APP_NAME):$(GIT_COMMIT)

## test: execute all tests in the repository
test: test-unit test-race

## test-unit: execute all unit tests in the repository
test-unit:
	CGO_ENABLED=0 go test -v -cover ./... 


## test-with-race: execute all tests with a race detector. Takes longer
test-race:
	go test -race ./...


.PHONY: help
## help: prints this help message
help:
	@echo "Usage: \n"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

