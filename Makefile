APP_NAME=$(shell basename $(CURDIR))
GIT_COMMIT=$(shell git rev-parse --short HEAD)

CONFIG=$(CURDIR)/config.json

OUT_DIR=build

.PHONY: build run dev

## build: builds the application in "build" directory
build: clean
	CGO_ENABLED=0 go build -o $(OUT_DIR)/$(APP_NAME) -ldflags "-s -w -X main.gitCommit=$(GIT_COMMIT)"

## run: runs the application in "build/game-repository"
run: build
	./$(OUT_DIR)/$(APP_NAME) --config=$(CONFIG)

## dev: executes the application with hot reload
dev: 
	air

## dev-dependecies: installs development dependencies
dev-dependecies:
	go install github.com/cosmtrek/air@latest

## docker-build: builds a docker image
docker-build:
	docker build -t $(APP_NAME):$(GIT_COMMIT) .

## docker-run: runs a docker container. Needs "config" argument (i.e make docker-run config=$(pwd)/config.json)
docker-run: docker-build
	docker run --network=host -v $(CONFIG):/app/config.json $(APP_NAME):$(GIT_COMMIT)

## docker-push: sends the image on a server with ssh (i.e make docker push SSH="10.10.1.1 -p1234")
docker-push: docker-build
	docker save $(APP_NAME):$(GIT_COMMIT) | bzip2 | pv | ssh $(SSH) docker load

## test: executes all unit tests in the repository
test:
	CGO_ENABLED=0 SKIP_INTEGRATION=1 go test -v -cover . 

## test-race: executes all unit tests with a race detector. Takes longer
test-race:
	go test -race .

## test-integration
test-integration:
ifeq ($(CI),)
	$(info CI is not defined)
	@ ID=$$(docker run -p 5432 -e POSTGRES_PASSWORD=postgres --rm -d postgres:14-alpine3.17); \
	PORT=$$(docker port $$ID | awk '{split($$0,a,":"); print a[2]}' ); \
	sleep 5; \
	CGO_ENABLED=0 DB_URI=postgres://postgres:postgres@localhost:$$PORT/postgres?sslmode=disable go test -v -cover . -- ; \
	docker kill $$ID
else
	$(info CI is defined)
	go test -v -coverprofile=coverage.out -c . -o testable .
	DB_URI=$(DB_URI) ./testable -test.coverprofile=coverage.out -test.v 
	go tool cover -func=coverage.out -o=coverage.out
endif

## clean: remove build files 
clean:
	rm -f build/*

.PHONY: help
## help: prints this help message
help:
	@echo "Usage: \n"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

