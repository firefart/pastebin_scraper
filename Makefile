GOPATH := $(or $(GOPATH), $(HOME)/go)

.DEFAULT_GOAL := build

.PHONY: build
build: deps test
	go vet ./...
	go fmt ./...
	go build -trimpath .

.PHONY: linux
linux: deps test
	GOOS=linux GOARCH=amd64 GO111MODULE=on go build -trimpath .

.PHONY: test
test: deps
	go test -v -race ./...

.PHONY: deps
deps:
	go get -u
	go mod tidy -v

.PHONY: lint
lint: deps
	if [ ! -f "$(go env GOPATH)/bin/golangci-lint" ]; then
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.24.0;
	fi
	golangci-lint run ./...
	go mod tidy
