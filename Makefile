export GO111MODULE=on
SHELL:=/bin/bash
DOCKER_C := docker-compose
.DEFAULT_GOAL := help
.PHONY: *

build-server: ## Build server component
	go clean -cache
	go mod tidy
	env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/zc cmd/server/main.go
	
build-client: ## Build server component
	go clean -cache
	go mod tidy
	env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/zc-cli cmd/client/main.go

tests: up-build ## Run all tests
	go clean -cache
	go test -count=1 -covermode=count -coverprofile=coverage.out github.com/jmsilvadev/zc/...
	go tool cover -func coverage.out

tests-coverage: up-build ## Run all tests with coverage in html
	go clean -cache
	go test -count=1 -covermode=count -coverprofile=coverage.out github.com/jmsilvadev/zc/...
	go tool cover -html=coverage.out -o coverage.html
	open coverage.html >&- 2>&- || \
	xdg-open coverage.html >&- 2>&- || \
	gnome-open coverage.html >&- 2>&-

tests-pkg: ## Run package tests
	go clean -cache
	go test -count=1 -covermode=count -coverprofile=coverage.out github.com/jmsilvadev/zc/pkg/...
	go tool cover -func coverage.out

tests-pkg-cover: ## Run package tests with coverage
	go clean -cache
	go test -count=1 -covermode=count -coverprofile=coverage.out github.com/jmsilvadev/zc/pkg/...
	go tool cover -html=coverage.out -o coverage.html
	open coverage.html >&- 2>&- || \
	xdg-open coverage.html >&- 2>&- || \
	gnome-open coverage.html >&- 2>&-

tests-client: up-build ## Run unit tests ion the client
	go clean -cache
	go test -count=1 -covermode=count -coverprofile=coverage.out github.com/jmsilvadev/zc/cmd/client/...
	go tool cover -func coverage.out

tests-server: ## Run unit tests in the server
	go clean -cache
	go test -count=1 -covermode=count -coverprofile=coverage.out github.com/jmsilvadev/zc/cmd/server/...
	go tool cover -func coverage.out 

clean: ## Clean all builts
	rm -rf ./bin

clean-tests: ## Clean tests
	go clean -cache
	rm *.out

up: ## Start docker container
	$(DOCKER_C) pull
	$(DOCKER_C) up -d 

up-build: ## Start docker container and rebuild the image
	go mod tidy
	go mod vendor
	$(DOCKER_C) pull
	$(DOCKER_C) up --build -d

down: ## Stop docker container
	$(DOCKER_C) down --remove-orphans

build-image:  ## Build docker image in daemon mode
	go mod tidy
	go mod vendor
	docker build . -t zc
	
logs: ## Watch docker log files
	$(DOCKER_C) logs --tail 100 -f

help:
	@grep -E '^[a-zA-Z._-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
