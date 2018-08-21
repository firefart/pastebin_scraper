GOPATH := $(or $(GOPATH), $(HOME)/go)
BIN_DIR := $(GOPATH)/bin
GOMETALINTER := $(BIN_DIR)/gometalinter

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

$(GOMETALINTER):
	go get -u github.com/alecthomas/gometalinter
	$(GOMETALINTER) --install &> /dev/null

.PHONY: lint
lint: deps $(GOMETALINTER)
	$(BIN_DIR)/gometalinter ./...
