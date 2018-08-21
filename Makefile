GOPATH := $(or $(GOPATH), $(HOME)/go)

.DEFAULT_GOAL := build

.PHONY: build
build: deps test
	go build .

.PHONY: test
test: deps lint
	go test -v ./...

.PHONY: deps
deps:
	go get -u gopkg.in/gomail.v2

.PHONY: lint
lint: deps
	go get -u github.com/alecthomas/gometalinter
	gometalinter --deadline=5m  ./...
