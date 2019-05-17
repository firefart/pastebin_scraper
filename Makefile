GOPATH := $(or $(GOPATH), $(HOME)/go)

.DEFAULT_GOAL := build

.PHONY: build
build: deps test
	go build .

.PHONY: test
test: deps
	go test -v -race ./...

.PHONY: deps
deps:
	go get -u
	go mod tidy -v

.PHONY: lint
lint: deps
	go get -u github.com/golangci/golangci-lint@master
	golangci-lint run ./...
	go mod tidy
