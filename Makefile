APP_NAME=$(shell basename $(CURDIR))
GIT_COMMIT=$(shell git rev-parse --short HEAD)

OUT_DIR=build

.PHONY: build run dev

## build: builds the project in a single file executable in "build" directory
build: clean
	CGO_ENABLED=0 go build -o $(OUT_DIR)/$(APP_NAME) -ldflags "-X main.gitCommit=$(GIT_COMMIT)"

## run: run application
run: build
	./$(OUT_DIR)/$(APP_NAME)

## docker-build: builds a docker image
docker-build:
	docker build -t $(APP_NAME):$(GIT_COMMIT) .

## docker-run: runs a docker container. Needs "config" argument (i.e make docker-run config=$(pwd)/config.json)
docker-run: docker-build
	docker run --network=host -v $(config):/app/config.json $(APP_NAME):$(GIT_COMMIT)

## test: executes all unit tests in the repository
test:
	CGO_ENABLED=0 go test -v -cover ./... 


## test-race: executes all tests with a race detector. Takes longer
test-race:
	go test -race ./...


## clean: remove build files 
clean:
	rm -f build/*

.PHONY: help
## help: prints this help message
help:
	@echo "Usage: \n"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

